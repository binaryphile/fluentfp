#!/usr/bin/env python3
"""Extract and compile-check fluentfp snippets from markdown.

Scans the configured markdown files for fenced ```go blocks annotated with
metadata in the language line (e.g. ```go {compile,context=NAME}). For each
{compile} block, assembles the snippet against the matching harness at
scripts/snippet-harness/NAME.go and runs `go build ./...` in a temp module
that points at the local fluentfp repo via a replace directive.

Block metadata vocabulary:
  {compile,context=NAME}        — block must build against scripts/snippet-harness/NAME.go
  {compile,context=NAME,slot=S} — block fills the // __SNIPPET_S__ marker (multi-slot harness)
  {run,output=...}              — RESERVED; not implemented in this prototype
  {ignore}                      — skip (illustrative pseudo-code)
  (no metadata)                 — skipped with a warning

Multi-slot semantics: when multiple blocks share a context, they are
assembled into ONE harness file and built TOGETHER. Each block specifies
which slot it fills via the `slot=` metadata (e.g. `slot=construct`,
`slot=call`). The harness file must have a matching `// __SNIPPET_<slot>__`
marker for each. Slots that no block fills remain as Go line comments
(harmless). Blocks omitting `slot=` use the default `// __SNIPPET__` marker
— the single-slot pattern is the special case of "slot=default" and stays
backward-compatible with single-block harnesses.

When the assembled multi-slot build fails, ALL blocks in that context
report FAIL (they share the build). Single-slot contexts behave exactly
as before.

Exit code 0 if every checked block passed; nonzero on any failure.
"""

import collections
import concurrent.futures
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
HARNESS_DIR = Path(__file__).resolve().parent / "snippet-harness"
DEFAULT_MARKER = "// __SNIPPET__"

TARGET_FILES = [
    REPO_ROOT / "README.md",
    REPO_ROOT / "docs" / "design.md",
    REPO_ROOT / "docs" / "parallelism-research.md",
    REPO_ROOT / "docs" / "showcase.md",
    REPO_ROOT / "examples" / "orders" / "README.md",
    REPO_ROOT / "kv" / "README.md",
    REPO_ROOT / "option" / "README.md",
    REPO_ROOT / "seq" / "README.md",
    REPO_ROOT / "slice" / "README.md",
    REPO_ROOT / "stream" / "README.md",
    REPO_ROOT / "web" / "README.md",
]
# All target files share opt-in semantics — un-annotated blocks emit
# warnings only. Per-file expansion intentionally lands one file at a
# time so each addition bundles its `{compile,...}` annotations,
# harnesses, and `{ignore}` rationale in a single commit.

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


def marker_for_slot(slot):
    """Return the harness marker string for a given slot name.

    The default slot (None) uses `// __SNIPPET__` — backward-compatible
    with single-block harnesses. Named slots use `// __SNIPPET_<slot>__`.
    """
    if slot is None:
        return DEFAULT_MARKER
    return f"// __SNIPPET_{slot}__"


def substitute(harness_text, body, marker):
    """Substitute `body` at the marker line in `harness_text`.

    Returns the new harness text, or None if no line matches.

    Match rule: a line MATCHES iff its `.strip()` equals `marker`. The
    replacement reconstructs the file by joining lines around the matched
    index — string-based `.replace()` would substring-match the same
    marker text if it appears in a doc-comment or other line first.

    A snippet body containing the literal marker text as a standalone
    line still creates a theoretical collision risk; in practice the
    markers are verbose enough (`// __SNIPPET[_<slot>]__`) that this
    requires deliberate construction. The default-slot marker is the
    shortest and most likely to collide with operator-written prose.
    """
    lines = harness_text.split("\n")
    for i, line in enumerate(lines):
        if line.strip() == marker:
            indent = line[: len(line) - len(line.lstrip())]
            indented_body = "\n".join(
                (indent + l) if l.strip() else l for l in body.split("\n")
            )
            return "\n".join(lines[:i] + [indented_body] + lines[i + 1 :])
    return None


def assemble_group(harness_text, blocks):
    """Substitute every block in the group at its respective marker.

    Strips any `//go:build ...` constraint lines from the harness first
    (the harness carries `//go:build ignore` to keep the in-tree file out
    of `go build ./...`; the assembled tmpdir file must be fully buildable
    and so the constraint is dropped).

    `blocks` is a list of (metadata, body, location) tuples. Returns
    (assembled_text, None) on success, or (None, error_message) on the
    first missing-marker.
    """
    lines = [
        line for line in harness_text.split("\n")
        if not line.lstrip().startswith("//go:build")
    ]
    assembled = "\n".join(lines)

    for metadata, body, location in blocks:
        marker = marker_for_slot(metadata.get("slot"))
        next_text = substitute(assembled, body, marker)
        if next_text is None:
            return None, f"harness missing marker {marker!r} (required by {location})"
        assembled = next_text

    return assembled, None


def check_compile_group(context, blocks):
    """Assemble all blocks in `blocks` into one harness file and build it.

    Returns (ok: bool, err: Optional[str]). When ok is False, every block
    in the group reports FAIL with the same error message — they share a
    build.

    Per-harness extra requires: if `scripts/snippet-harness/<NAME>.gomod`
    exists, each non-blank line is appended as an additional `require`
    line in the assembled tmpdir go.mod. Lines starting with `#` are
    ignored as comments. Use this for harnesses that exercise external
    packages (e.g. `golang.org/x/time/rate v0.5.0`).
    """
    harness_file = HARNESS_DIR / f"{context}.go"
    if not harness_file.exists():
        return False, f"harness file not found: {harness_file}"

    harness = harness_file.read_text()
    assembled, err = assemble_group(harness, blocks)
    if assembled is None:
        return False, err

    extra_requires = []
    gomod_extra = HARNESS_DIR / f"{context}.gomod"
    if gomod_extra.exists():
        for line in gomod_extra.read_text().splitlines():
            line = line.strip()
            if line and not line.startswith("#"):
                extra_requires.append(f"require {line}")

    with tempfile.TemporaryDirectory(prefix="fluentfp-snippet-") as tmp:
        tmp = Path(tmp)
        (tmp / "snippet.go").write_text(assembled)
        gomod_lines = [
            "module snippet",
            "",
            "go 1.26",
            "",
            "require github.com/binaryphile/fluentfp v0.0.0",
        ]
        gomod_lines.extend(extra_requires)
        gomod_lines.append("")
        gomod_lines.append(f"replace github.com/binaryphile/fluentfp => {REPO_ROOT}")
        gomod_lines.append("")
        (tmp / "go.mod").write_text("\n".join(gomod_lines))
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
                env={**os.environ, "GOFLAGS": "-mod=mod"},
            )
        except FileNotFoundError:
            return False, "go: command not found (run via 'nix develop -c')"

        if proc.returncode == 0:
            return True, None

        msg = (proc.stderr or proc.stdout).strip()
        return False, msg


def main():
    failures = 0
    total = 0
    skipped = 0
    warnings = 0

    # Per-file walk; record each block's metadata + location + body.
    # Buildable blocks are grouped by context for the multi-slot build step.
    block_records = []  # list of (location, metadata, body, kind)
                        # kind ∈ {"warn", "skip", "buildable"}
    blocks_by_context = collections.OrderedDict()

    for md_path in TARGET_FILES:
        content = md_path.read_text()
        for match in FENCED_RE.finditer(content):
            meta_str = match.group(1)
            body = match.group(2)
            metadata = parse_metadata(meta_str)

            line_no = content[: match.start()].count("\n") + 1
            location = f"{md_path.relative_to(REPO_ROOT)}:{line_no}"

            if metadata.get("ignore"):
                block_records.append((location, metadata, body, "skip"))
                skipped += 1
                continue
            if not metadata:
                block_records.append((location, metadata, body, "warn"))
                warnings += 1
                continue
            if not metadata.get("compile"):
                # Future: {run,output=...}. Skip for the prototype.
                block_records.append((location, metadata, body, "skip"))
                skipped += 1
                continue

            context = metadata.get("context")
            if not context:
                # Hard error: {compile} with no context means we can't pick
                # a harness. Report inline so the operator sees the location.
                block_records.append((location, metadata, body, "skip"))
                failures += 1
                total += 1
                print(f"FAIL {location}  [<no-context>]")
                print(f"     block has {{compile}} but no context=NAME")
                continue

            block_records.append((location, metadata, body, "buildable"))
            blocks_by_context.setdefault(context, []).append(
                (metadata, body, location)
            )

    # Run check per context in parallel. Each context's build runs in its
    # own tmpdir with no shared state, so a ThreadPoolExecutor is safe; the
    # work is subprocess-bound (`go build`) so the GIL isn't the bottleneck.
    # Worker count caps at min(8, contexts) — enough to saturate a modern
    # machine without thrashing the Go build cache.
    context_results = {}
    worker_count = min(8, max(1, len(blocks_by_context)))
    with concurrent.futures.ThreadPoolExecutor(max_workers=worker_count) as ex:
        futures = {
            ex.submit(check_compile_group, ctx, blocks): ctx
            for ctx, blocks in blocks_by_context.items()
        }
        for fut in concurrent.futures.as_completed(futures):
            ctx = futures[fut]
            context_results[ctx] = fut.result()

    # Print per-block results in file-order. All blocks in a failing
    # context group report the same error.
    for location, metadata, body, kind in block_records:
        if kind == "warn":
            print(f"WARN {location}  un-annotated go block (add metadata or {{ignore}})")
            continue
        if kind == "skip":
            continue

        # kind == "buildable"
        context = metadata["context"]
        slot = metadata.get("slot")
        slot_suffix = f":{slot}" if slot is not None else ""
        ok, err = context_results[context]
        total += 1
        if ok:
            print(f"OK   {location}  [{context}{slot_suffix}]")
        else:
            print(f"FAIL {location}  [{context}{slot_suffix}]")
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
