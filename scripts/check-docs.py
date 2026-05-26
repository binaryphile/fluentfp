#!/usr/bin/env python3
"""Verify count and anchor consistency across README.md, docs/showcase.md, and methodology.md.

Catches two drift classes:

1. Showcase entry count vs claimed counts in showcase title/intro and README.
2. Markdown cross-document anchor links pointing at headings that don't exist
   (heading reworded, file renamed, slug formula wrong).

Exits 0 if everything matches; non-zero with a summary of mismatches otherwise.
Run from repo root: `python3 scripts/check-docs.py` or `./mk docs-check`.
"""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
README = ROOT / "README.md"
SHOWCASE = ROOT / "docs" / "showcase.md"
METHODOLOGY = ROOT / "methodology.md"
ANALYSIS = ROOT / "analysis.md"

# Files whose anchors we resolve when checking cross-document links.
ANCHORED = {
    "README.md": README,
    "docs/showcase.md": SHOWCASE,
    "methodology.md": METHODOLOGY,
    "analysis.md": ANALYSIS,
}


def gfm_slug(heading: str) -> str:
    """Approximate GitHub's heading-to-anchor slugger.

    Rules: lowercase, strip leading hashes/spaces, remove any character
    that isn't a letter, digit, hyphen, or space, then replace spaces
    with hyphens. Adjacent hyphens (e.g., from em-dash) are preserved.
    """
    h = heading.lstrip("#").strip().lower()
    h = re.sub(r"[^\w\- ]", "", h, flags=re.UNICODE)  # drop punct, em-dash, slash, period
    h = re.sub(r"_", "", h)  # \w includes underscore; GitHub strips it from slugs
    h = h.replace(" ", "-")
    return h


def headings(path: Path) -> set[str]:
    """Extract markdown heading slugs from a file, skipping fenced code blocks."""
    slugs: set[str] = set()
    in_fence = False
    for line in path.read_text(encoding="utf-8").splitlines():
        if line.lstrip().startswith("```"):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        m = re.match(r"^(#{1,6})\s+(.+?)\s*$", line)
        if m:
            slugs.add(gfm_slug(m.group(2)))
    return slugs


def count_entries(path: Path) -> int:
    """Count `### ` headings (entries) in a file, skipping fenced code blocks."""
    count = 0
    in_fence = False
    for line in path.read_text(encoding="utf-8").splitlines():
        if line.lstrip().startswith("```"):
            in_fence = not in_fence
            continue
        if not in_fence and line.startswith("### "):
            count += 1
    return count


# Link form: [anything](file.md#anchor) or [anything](file.md) or [anything](#anchor).
# We only validate cross-document links to files in ANCHORED.
LINK_RE = re.compile(r"\[[^\]]+\]\(([^)\s]+?)(?:#([^)\s]+))?\)")


def cross_doc_links(source: Path):
    """Yield (target_path_str, anchor_or_None) for links from source to ANCHORED files."""
    text = source.read_text(encoding="utf-8")
    source_dir = source.parent
    for url, anchor in LINK_RE.findall(text):
        if url.startswith(("http://", "https://", "mailto:")):
            continue
        if not url.endswith(".md"):
            continue
        # Resolve relative to source file.
        target = (source_dir / url).resolve()
        try:
            rel = target.relative_to(ROOT)
        except ValueError:
            continue
        key = str(rel).replace("\\", "/")
        if key in ANCHORED:
            yield (key, anchor or None, source)


def main() -> int:
    errors: list[str] = []

    # 1. Showcase entry count vs claimed counts.
    actual_entries = count_entries(SHOWCASE)
    showcase_text = SHOWCASE.read_text(encoding="utf-8")
    readme_text = README.read_text(encoding="utf-8")

    # Showcase title: "# From Mechanics to Intent: N Real-World Rewrites"
    m = re.search(r"^#\s+From Mechanics to Intent:\s+(\d+)\s+Real-World Rewrites",
                  showcase_text, flags=re.MULTILINE)
    if m and int(m.group(1)) != actual_entries:
        errors.append(
            f"showcase.md title claims {m.group(1)} rewrites but file has "
            f"{actual_entries} `### ` entries"
        )

    # Showcase intro: "The N examples below..."
    m = re.search(r"The\s+(\d+)\s+examples below", showcase_text)
    if m and int(m.group(1)) != actual_entries:
        errors.append(
            f"showcase.md intro says 'The {m.group(1)} examples below' but file has "
            f"{actual_entries} `### ` entries"
        )

    # README: "for N before/after rewrites"
    m = re.search(r"for\s+(\d+)\s+before/after rewrites", readme_text)
    if m and int(m.group(1)) != actual_entries:
        errors.append(
            f"README.md says 'for {m.group(1)} before/after rewrites' but showcase has "
            f"{actual_entries} entries"
        )

    # README: "N more rewrites" — should equal total minus the 2 shown in README.
    m = re.search(r"has\s+(\d+)\s+more rewrites", readme_text)
    if m:
        claimed_more = int(m.group(1))
        # The README itself contains two before/after table examples.
        readme_examples = 2
        expected_more = actual_entries - readme_examples
        if claimed_more != expected_more:
            errors.append(
                f"README.md says 'has {claimed_more} more rewrites' but expected "
                f"{expected_more} (showcase {actual_entries} − {readme_examples} shown in README)"
            )

    # 2. Cross-document anchor existence.
    for source in (README, SHOWCASE):
        for target_key, anchor, src in cross_doc_links(source):
            target_path = ANCHORED[target_key]
            if not target_path.exists():
                errors.append(
                    f"{src.relative_to(ROOT)} → {target_key}: target file missing"
                )
                continue
            if anchor is None:
                continue
            target_slugs = headings(target_path)
            if anchor not in target_slugs:
                errors.append(
                    f"{src.relative_to(ROOT)} → {target_key}#{anchor}: anchor not found "
                    f"(target file has no heading with slug `{anchor}`)"
                )

    if errors:
        print("docs-check: drift detected", file=sys.stderr)
        for e in errors:
            print(f"  - {e}", file=sys.stderr)
        return 1

    print(f"docs-check: ok ({actual_entries} showcase entries, all anchors resolved)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
