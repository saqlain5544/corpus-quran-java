"""
Reader — presents Quranic text for reading/display with full Mushaf annotation.
Returns complete verse text with all diacritics, pause marks, and orthographic marks.
"""

from __future__ import annotations

import json, csv
from pathlib import Path
from functools import lru_cache
from xml.etree.cElementTree import iterparse

_DATA = Path(__file__).resolve().parent.parent / "data" / "quran-text"


def _local(tag: str) -> str:
    return tag.rsplit("}", 1)[-1] if "}" in tag else tag


class Reader:
    """Read the Quran with full Mushaf annotation for display purposes."""

    def __init__(self, variant: str = "uthmani"):
        self._variant = variant
        self._data = self._load()

    def _load(self) -> dict[int, dict[int, str]]:
        path = _DATA / f"quran-{self._variant}.xml"
        if not path.exists():
            path = _DATA / "quran-uthmani.xml"
        result: dict[int, dict[int, str]] = {}
        current_sura = 0
        for event, elem in iterparse(str(path), events=("start", "end")):
            tag = _local(elem.tag).lower()
            if event == "start" and tag == "sura":
                idx = elem.get("index")
                if idx:
                    current_sura = int(idx)
                    result.setdefault(current_sura, {})
            elif event == "end" and tag == "aya":
                idx = elem.get("index")
                text = elem.get("text", "")
                name = elem.get("name", "")
                bism = elem.get("bismillah", "")
                if idx and current_sura:
                    aya_num = int(idx)
                    if bism and aya_num == 1 and current_sura != 1:
                        text = bism + " " + text
                    result[current_sura][aya_num] = text
                elem.clear()
            elif event == "end" and tag == "sura":
                elem.clear()
        return result

    def verse(self, surah: int, aya: int) -> str:
        return self._data.get(surah, {}).get(aya, "")

    def surah_text(self, surah: int) -> str:
        verses = self._data.get(surah, {})
        return "\n".join(verses.get(i, "") for i in sorted(verses))

    def juz(self, juz_num: int) -> str:
        lines: list[str] = []
        for surah in range(1, 115):
            for aya_num in sorted(self._data.get(surah, {})):
                lines.append(self._data[surah][aya_num])
        return "\n".join(lines)

    @property
    def chapters(self) -> list[int]:
        return sorted(self._data.keys())

    def __getitem__(self, key: tuple[int, int]) -> str:
        return self.verse(key[0], key[1])

    def __repr__(self) -> str:
        return f"Reader({self._variant}, {len(self._data)} chapters)"


class Translations:
    """Load and query Quran translations."""

    def __init__(self, lang: str = "en"):
        self._lang = lang
        self._data: dict[int, dict[int, str]] = {}
        self._load()

    def _load(self) -> None:
        for path in _DATA.glob("*.csv"):
            try:
                with open(path) as f:
                    reader = csv.DictReader(f)
                    for row in reader:
                        sura = int(row.get("sura", row.get("chapter", 0)))
                        aya = int(row.get("aya", row.get("verse", 0)))
                        text = row.get("text", row.get("translation", ""))
                        if sura and aya and text:
                            self._data.setdefault(sura, {})[aya] = text
            except Exception:
                pass

    def verse(self, surah: int, aya: int) -> str:
        return self._data.get(surah, {}).get(aya, "")

    def __repr__(self) -> str:
        return f"Translations({len(self._data)} chapters)"
