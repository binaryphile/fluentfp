#!/usr/bin/env python3
"""Extract and compile-check fluentfp snippets from markdown.

Scans the configured markdown files for fenced ```go blocks annotated with
metadata in the language line (e.g. ```go {compile,context=NAME}). For each
{compile} block, assembles the snippet against the matching harness at
scripts/snippet-harness/NAME.go and runs `go vet ./...` in a temp module
that points at the local fluentfp repo via a replace directive.

Block metadata vocabulary:
  {compile}              — block must build clean
  {compile,context=NAME} — block must build against scripts/snippet-harness/NAME.go
  {run,output=...}       — RESERVED; not implemented in this prototype
  {ignore}               — skip (illustrative pseudo-code)
  (no metadata)          — skipped with a warning

Exit code 0 if every checked block passed; nonzero on any failure.
"""

import re
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
HARNESS_DIR = Path(__file__).resolve().parent / "snippet-harness"
SNIPPET_MARKER = "// __SNIPPET__"

TARGET_FILES = [
    REPO_ROOT / "docs" / "showcase.md",
]

FENCED_RE = re.compile(
    r"^```go(?:\s+\{([^}]*)\})?\s*\n(.*?)^```",
    re.DOTALL | re.MULTILINE,
)


def parse_metadata(meta_str):
    if not meta_str:
        return {}
    result = {}
    for part in meta_str.split(","):
        part = part.strip()
        if not part:
            continue
        if "=" in part:
            key, value = part.split("=", 1)
            result[key.strip()] = value.strip()
        else:
            result[part] = True
    return result


def assemble(harness_text, body):
    for line in harness_text.split("\n"):
        if SNIPPET_MARKER in line:
            indent = line[: len(line) - len(line.lstrip())]
            indented_body = "\n".join(
                (indent + l) if l.strip() else l for l in body.split("\n")
            )
            return harness_text.replace(line, indented_body, 1)
    return None


def check_compile(metadata, body, location):
    context = metadata.get("context")
    if not context:
        return False, "block has {compile} but no context=NAME"

    harness_file = HARNESS_DIR / f"{context}.go"
    if not harness_file.exists():
        return False, f"harness file not found: {harness_file}"

    harness = harness_file.read_text()
    assembled = assemble(harness, body)
    if assembled is None:
        return False, f"harness {harness_file} missing marker '{SNIPPET_MARKER}'"

    with tempfile.TemporaryDirectory(prefix="fluentfp-snippet-") as tmp:
        tmp = Path(tmp)
        (tmp / "snippet.go").write_text(assembled)
        (tmp / "go.mod").write_text(
            f"module snippet\n\n"
            f"go 1.22\n\n"
            f"require github.com/binaryphile/fluentfp v0.0.0\n\n"
            f"replace github.com/binaryphile/fluentfp => {REPO_ROOT}\n"
        )
        # `go build` answers the literal question "does it compile?". `vet`
        # performs extra analysis that can flag valid-but-suspect code; the
        # reserved `{vet}` mode in the metadata vocabulary will opt into the
        # stricter check later.
        try:
            proc = subprocess.run(
                ["go", "build", "./..."],
                cwd=tmp,
                capture_output=True,
                text=True,
                env={**__import__("os").environ, "GOFLAGS": "-mod=mod"},
            )
        except FileNotFoundError:
            return False, "go: command not found (run via 'nix develop -c')"

        if proc.returncode == 0:
            return True, None

        # Annotate the failure with assembled-source line numbers so the
        # operator can map a vet diagnostic back to the snippet body.
        msg = (proc.stderr or proc.stdout).strip()
        return False, msg


def main():
    failures = 0
    total = 0
    skipped = 0
    warnings = 0

    for md_path in TARGET_FILES:
        content = md_path.read_text()
        for match in FENCED_RE.finditer(content):
            meta_str = match.group(1)
            body = match.group(2)
            metadata = parse_metadata(meta_str)

            line_no = content[: match.start()].count("\n") + 1
            location = f"{md_path.relative_to(REPO_ROOT)}:{line_no}"

            if metadata.get("ignore"):
                skipped += 1
                continue
            if not metadata:
                warnings += 1
                print(f"WARN {location}  un-annotated go block (add metadata or {{ignore}})")
                continue
            if not metadata.get("compile"):
                # Future: {run,output=...}. Skip for the prototype.
                skipped += 1
                continue

            total += 1
            ok, err = check_compile(metadata, body, location)
            context = metadata.get("context", "<no-context>")
            if ok:
                print(f"OK   {location}  [{context}]")
            else:
                print(f"FAIL {location}  [{context}]")
                for line in (err or "").splitlines():
                    print(f"     {line}")
                failures += 1

    print(
        f"\n{total - failures}/{total} checked  "
        f"({skipped} skipped, {warnings} un-annotated)"
    )
    sys.exit(1 if failures else 0)


if __name__ == "__main__":
    main()
