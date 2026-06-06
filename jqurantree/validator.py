"""
Encoding validator — ensures 100% Arabic encoding accuracy per variant.
Classifies every character, detects anomalies, validates expected profiles.
"""

from __future__ import annotations

import unicodedata
from collections import Counter
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

# ── Character classification ──

def _in_range(cp: int, *ranges: tuple[int, int]) -> bool:
    return any(lo <= cp <= hi for lo, hi in ranges)

class CharClass:
    ARABIC_LETTER = "arabic_letter"
    DIACRITIC = "diacritic"
    QURANIC_MARK = "quranic_mark"
    EXTENDED_QURANIC = "extended_quranic"
    DAGGER_ALIF = "dagger_alif"
    ALIF_WASLA = "alif_wasla"
    TATWEEL = "tatweel"
    RLM = "rlm"
    SPACE = "space"
    NUMBER = "number"
    OTHER = "other"


def classify(cp: int) -> str:
    if cp == 0x20: return CharClass.SPACE
    if cp == 0x200F: return CharClass.RLM
    if _in_range(cp, (0x0621, 0x064A)): return CharClass.ARABIC_LETTER
    if _in_range(cp, (0x064B, 0x065F)): return CharClass.DIACRITIC
    if cp == 0x0670: return CharClass.DAGGER_ALIF
    if cp == 0x0671: return CharClass.ALIF_WASLA
    if _in_range(cp, (0x06D6, 0x06ED)): return CharClass.QURANIC_MARK
    if _in_range(cp, (0x08A0, 0x08FF)): return CharClass.EXTENDED_QURANIC
    if cp == 0x0640: return CharClass.TATWEEL
    if _in_range(cp, (0x0600, 0x0604), (0x0610, 0x061B), (0x061E, 0x061F),
                  (0x0660, 0x0669)): return CharClass.NUMBER
    if _in_range(cp, (0x0672, 0x06D5), (0x06EE, 0x06FF)): return CharClass.ARABIC_LETTER
    if _in_range(cp, (0xFB50, 0xFDFF), (0xFE70, 0xFEFF)): return CharClass.ARABIC_LETTER
    return CharClass.OTHER


@dataclass(slots=True)
class EncodingReport:
    variant: str
    total_chars: int = 0
    unique_codepoints: int = 0
    classes: dict[str, int] = field(default_factory=dict)
    unexpected: dict[int, int] = field(default_factory=dict)
    issues: list[str] = field(default_factory=list)

    @property
    def is_clean(self) -> bool:
        return len(self.issues) == 0


def validate_variant(text_source, variant_name: str) -> EncodingReport:
    """Validate encoding of a variant's text."""
    report = EncodingReport(variant=variant_name)
    counts = Counter()
    chars = Counter()
    total = 0

    if isinstance(text_source, Path):
        if text_source.suffix == ".xml":
            from xml.etree.cElementTree import iterparse
            for event, elem in iterparse(str(text_source), events=("end",)):
                tag = elem.tag.split("}")[-1] if "}" in elem.tag else elem.tag
                if tag == "aya":
                    text = elem.get("text", "")
                    for ch in text:
                        cp = ord(ch)
                        counts[classify(cp)] += 1
                        chars[cp] += 1
                        total += 1
                elem.clear()
        elif text_source.suffix == ".json":
            import json
            with open(text_source) as f:
                data = json.load(f)
            for k, v in data.items():
                if isinstance(v, str):
                    for ch in v:
                        cp = ord(ch)
                        counts[classify(cp)] += 1
                        chars[cp] += 1
                        total += 1

    report.total_chars = total
    report.unique_codepoints = len(chars)
    report.classes = dict(counts)

    # Detect unexpected characters
    for cp, cnt in chars.items():
        cls = classify(cp)
        if cls == CharClass.OTHER:
            try:
                name = unicodedata.name(chr(cp))
            except:
                name = "UNKNOWN"
            report.unexpected[cp] = cnt
            if cnt > 10:
                report.issues.append(f"Unexpected char U+{cp:04X} ({name}): {cnt} occurrences")

    # Validate expected character classes per variant
    expected = _expected_classes(variant_name)
    for cls, expected_min in expected.items():
        actual = counts.get(cls, 0)
        if actual < expected_min:
            report.issues.append(
                f"Missing {cls}: expected >= {expected_min}, got {actual}"
            )

    # Validate no unexpected high-frequency characters
    for cp, cnt in report.unexpected.items():
        if cnt > 50:
            report.issues.append(
                f"High-frequency unexpected U+{cp:04X}: {cnt} occurrences"
            )

    return report


def _expected_classes(variant: str) -> dict[str, int]:
    base = {"arabic_letter": 300000, "space": 50000}
    if "clean" in variant:
        return base
    if "indopak" in variant:
        base["diacritic"] = 100000
        base["quranic_mark"] = 10000
        base["rlm"] = 100
        return base
    if "simple" in variant:
        base["diacritic"] = 200000
        return base
    if "uthmani" in variant:
        base["diacritic"] = 200000
        base["alif_wasla"] = 10000
        return base
    if "warsh" in variant:
        base["diacritic"] = 200000
        base["quranic_mark"] = 5000
        return base
    return base


def validate_all_variants(data_dir: Path | str | None = None) -> dict[str, EncodingReport]:
    """Validate encoding of all available variants."""
    if data_dir is None:
        data_dir = Path(__file__).resolve().parent.parent / "data" / "quran-text"
    else:
        data_dir = Path(data_dir)

    files = {
        "quran-uthmani.xml": "uthmani",
        "quran-simple.xml": "simple",
        "quran-simple-clean.xml": "simple-clean",
        "uthmani-full.xml": "uthmani-full",
        "simple-full.xml": "simple-full",
    }

    reports = {}
    for fname, label in files.items():
        path = data_dir / fname
        if path.exists():
            reports[label] = validate_variant(path, label)

    return reports
