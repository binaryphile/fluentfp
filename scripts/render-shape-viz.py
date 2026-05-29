#!/usr/bin/env python3
# Renders Go source files as code-shape heatmap SVGs for examples/loop-to-chain.
# Each source line is a colored rectangle: width = non-indent content length,
# x-offset = indent depth, color = nesting depth (plasma palette). Function
# declarations and matching close braces are width-capped to median body width
# (floor-guarded) and color-borrowed from the function body so the block reads
# as one unit.
#
# Output: deterministic, side-by-side per pair, written to images/*.svg.

from pathlib import Path
import re
import statistics

# Plasma 6-step palette: perceptually uniform, monotonic lightness, cool-to-hot.
PLASMA = [
    "#0d0887",  # indent 0 — top-level
    "#7e03a8",  # indent 1 — function body
    "#cc4778",  # indent 2 — nested
    "#f0844c",  # indent 3
    "#fcce25",  # indent 4
    "#f0f921",  # indent 5+
]

ROW_HEIGHT = 4
CHAR_UNIT = 2
TAB_UNIT = 8
FLOOR_CHARS = 20
GUTTER = 16
PAD = 8
BG = "#fafafa"

FUNC_DECL_RE = re.compile(r"^func\s")


def color_for_indent(indent):
    return PLASMA[min(max(indent, 0), len(PLASMA) - 1)]


def analyze(text):
    lines = text.split("\n")
    info = []
    for raw in lines:
        if not raw.strip():
            info.append({"indent": 0, "content_len": 0, "raw": raw, "kind": "blank"})
            continue
        indent = 0
        while indent < len(raw) and raw[indent] == "\t":
            indent += 1
        content = raw[indent:].rstrip()
        is_decl = bool(FUNC_DECL_RE.match(content))
        info.append(
            {
                "indent": indent,
                "content_len": len(content),
                "raw": raw,
                "kind": "func_decl" if is_decl else "normal",
            }
        )

    i = 0
    while i < len(info):
        if info[i]["kind"] != "func_decl":
            i += 1
            continue

        decl_indent = info[i]["indent"]
        raw = info[i]["raw"].rstrip()

        # Single-line function: decl + body + close-brace all on one line.
        # No body lines to compute median from, so cap at FLOOR_CHARS and color
        # as if the inline body were at decl_indent + 1.
        if "{" in raw and raw.count("{") == raw.count("}"):
            info[i]["borrowed_indent"] = decl_indent + 1
            info[i]["capped_content_len"] = min(info[i]["content_len"], FLOOR_CHARS)
            i += 1
            continue

        body_widths = []
        body_indent = None
        close_idx = None
        j = i + 1
        while j < len(info):
            row = info[j]
            if row["kind"] == "blank":
                j += 1
                continue
            if row["indent"] == decl_indent and row["raw"].strip() == "}":
                close_idx = j
                break
            if body_indent is None and row["indent"] > decl_indent:
                body_indent = row["indent"]
            body_widths.append(row["content_len"])
            j += 1

        if body_widths and body_indent is not None:
            median = statistics.median(body_widths)
            cap_chars = max(median, FLOOR_CHARS)
            info[i]["borrowed_indent"] = body_indent
            info[i]["capped_content_len"] = min(info[i]["content_len"], cap_chars)
            if close_idx is not None:
                info[close_idx]["kind"] = "close_brace_func"
                info[close_idx]["borrowed_indent"] = body_indent
                info[close_idx]["capped_content_len"] = min(
                    info[close_idx]["content_len"], cap_chars
                )

        i = (close_idx + 1) if close_idx is not None else i + 1

    return info


def column_width(info):
    widths = []
    for row in info:
        if row["kind"] == "blank":
            continue
        content_len = row.get("capped_content_len", row["content_len"])
        widths.append(row["indent"] * TAB_UNIT + content_len * CHAR_UNIT)
    return max(widths) if widths else 0


def render_svg(left_info, right_info):
    n = max(len(left_info), len(right_info))
    height = n * ROW_HEIGHT + 2 * PAD
    left_w = column_width(left_info)
    right_w = column_width(right_info)
    width = PAD + left_w + GUTTER + right_w + PAD

    out = [
        f'<svg xmlns="http://www.w3.org/2000/svg" '
        f'viewBox="0 0 {int(width)} {int(height)}" '
        f'width="{int(width)}" height="{int(height)}" '
        f'preserveAspectRatio="xMinYMin meet">',
        f'<rect width="{int(width)}" height="{int(height)}" fill="{BG}"/>',
    ]

    def emit(info, x_off):
        for idx, row in enumerate(info):
            if row["kind"] == "blank":
                continue
            indent = row["indent"]
            color_indent = row.get("borrowed_indent", indent)
            color = color_for_indent(color_indent)
            content_len = row.get("capped_content_len", row["content_len"])
            x = PAD + x_off + indent * TAB_UNIT
            y = PAD + idx * ROW_HEIGHT
            w = content_len * CHAR_UNIT
            if w <= 0:
                continue
            out.append(
                f'<rect x="{int(x)}" y="{int(y)}" '
                f'width="{int(w)}" height="{ROW_HEIGHT}" fill="{color}"/>'
            )

    emit(left_info, 0)
    emit(right_info, left_w + GUTTER)
    out.append("</svg>")
    return "\n".join(out) + "\n"


def main():
    root = Path(__file__).resolve().parent.parent
    src = root / "examples" / "loop-to-chain"
    img = root / "images"

    pairs = [
        ("conventional.go", "fluentfp.go", "code-shape-comparison.svg"),
        (
            "best-case-conventional.go",
            "best-case-fluentfp.go",
            "best-case-code-shape-comparison.svg",
        ),
    ]

    for left_name, right_name, out_name in pairs:
        left_info = analyze((src / left_name).read_text())
        right_info = analyze((src / right_name).read_text())
        svg = render_svg(left_info, right_info)
        (img / out_name).write_text(svg)
        print(f"wrote images/{out_name}")


if __name__ == "__main__":
    main()
