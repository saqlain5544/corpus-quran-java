#!/usr/bin/env python3
"""
extract-font-cmaps.py — One-time helper that reads hafs.woff2 and
AmiriQuran-Regular.woff2 and dumps each font's cmap (codepoint →
glyph name) as JSON to backend/internal/server/testdata/fontcmaps.json.

Run after font files change. The Go server reads the JSON at startup
via `go:embed`.

Why a Python script: woff2 is brotli-compressed SFNT, and Go's
standard library doesn't include a brotli decoder. Pulling in
`github.com/andybalholm/brotli` would add a dependency for what's
fundamentally a build-time tooling concern. Python has fontTools +
brotli already available (we use both for the data pipeline).
"""

import json
import sys
from pathlib import Path

from fontTools.ttLib import TTFont


def cmap_from(woff2_path: Path) -> dict[str, str]:
    """Return {hex_codepoint: glyph_name} for the woff2 file's cmap."""
    font = TTFont(str(woff2_path))
    cmap = font.getBestCmap()
    return {f"{cp:04X}": name for cp, name in cmap.items()}


def main() -> int:
    base = Path(__file__).resolve().parent.parent
    hafs = base / "backend" / "fonts" / "hafs.woff2"
    amiri = base / "backend" / "fonts" / "AmiriQuran-Regular.woff2"
    out = base / "backend" / "internal" / "server" / "testdata" / "fontcmaps.json"

    if not hafs.exists():
        print(f"missing {hafs}", file=sys.stderr)
        return 1
    if not amiri.exists():
        print(f"missing {amiri}", file=sys.stderr)
        return 1

    out.parent.mkdir(parents=True, exist_ok=True)
    data = {
        "hafs": cmap_from(hafs),
        "amiriquran": cmap_from(amiri),
    }
    out.write_text(json.dumps(data, indent=2, ensure_ascii=False, sort_keys=True))
    print(
        f"wrote {out}  ({len(data['hafs'])} hafs, {len(data['amiriquran'])} amiriquran)"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
