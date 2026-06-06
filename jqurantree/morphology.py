"""
Morphology — query the Quranic Arabic Corpus morphological annotation (MASAQ).
157,676 segments, 133 POS tags, 60 syntactic roles, per-word English gloss.
"""

from __future__ import annotations

import csv
from dataclasses import dataclass, field
from functools import lru_cache
from pathlib import Path
from typing import Optional

_DATA_DIR = Path(__file__).resolve().parent.parent / "data" / "morphology"


@dataclass(slots=True)
class Segment:
    id: int
    sura: int
    verse: int
    word: int
    segment: int
    word_text: str
    without_diacritics: str
    segmented_text: str
    pos: str
    morph_type: str
    syntactic_role: str
    possessive: str
    case_mood: str
    case_marker: str
    phrase: str
    phrase_function: str
    gloss: str

    @property
    def is_prefix(self) -> bool: return self.morph_type == "Prefix"
    @property
    def is_stem(self) -> bool: return self.morph_type == "Stem"
    @property
    def is_suffix(self) -> bool: return self.morph_type == "Suffix"

    def __repr__(self) -> str:
        return f"Segment({self.sura}:{self.verse}:{self.word}.{self.segment} {self.pos} {self.segmented_text!r})"


class Morphology:
    __slots__ = ('_segments', '_by_pos', '_by_word')

    def __init__(self, csv_path: Path | str | None = None):
        path = Path(csv_path) if csv_path else _DATA_DIR / "MASAQ.csv"
        self._segments: list[Segment] = []
        self._by_pos: dict[str, list[Segment]] = {}
        self._by_word: dict[tuple[int, int, int], list[Segment]] = {}

        with open(path, newline="", encoding="utf-8") as f:
            for row in csv.DictReader(f):
                seg = Segment(
                    id=int(row["ID"]),
                    sura=int(row["Sura_No"]),
                    verse=int(row["Verse_No"]),
                    word=int(row["Word_No"]),
                    segment=int(row["Segment_No"]),
                    word_text=row["Word"],
                    without_diacritics=row["Without_Diacritics"],
                    segmented_text=row["Segmented_Word"],
                    pos=row["Morph_Tag"],
                    morph_type=row["Morph_Type"],
                    syntactic_role=row["Syntactic_Role"],
                    possessive=row["Possessive_Construct"],
                    case_mood=row["Case_Mood"],
                    case_marker=row["Case_Mood_Marker"],
                    phrase=row["Phrase"],
                    phrase_function=row["Phrasal_Function"],
                    gloss=row["Gloss"],
                )
                self._segments.append(seg)
                self._by_pos.setdefault(seg.pos, []).append(seg)
                key = (seg.sura, seg.verse, seg.word)
                self._by_word.setdefault(key, []).append(seg)

    def __len__(self) -> int: return len(self._segments)

    def word(self, sura: int, verse: int, word_num: int) -> list[Segment]:
        return self._by_word.get((sura, verse, word_num), [])

    def verse(self, sura: int, verse: int) -> list[Segment]:
        return [s for s in self._segments if s.sura == sura and s.verse == verse]

    def surah(self, sura: int) -> list[Segment]:
        return [s for s in self._segments if s.sura == sura]

    def by_pos(self, pos: str) -> list[Segment]:
        return self._by_pos.get(pos, [])

    def gloss(self, sura: int, verse: int, word: int) -> str:
        segs = self.word(sura, verse, word)
        if not segs: return ""
        stems = [s.gloss for s in segs if s.is_stem]
        return " + ".join(stems) if stems else segs[0].gloss

    def search_gloss(self, term: str, case_sensitive: bool = False) -> list[Segment]:
        t = term if case_sensitive else term.lower()
        return [s for s in self._segments
                if (t in s.gloss if case_sensitive else t in s.gloss.lower())]

    def word_count(self) -> int: return len(self._by_word)

    def segment_count(self) -> int: return len(self._segments)

    def pos_distribution(self) -> dict[str, int]:
        from collections import Counter
        return dict(Counter(s.pos for s in self._segments).most_common(20))

    def lemma_map(self) -> dict[tuple[int, int, int], dict]:
        result: dict[tuple[int, int, int], dict] = {}
        for seg in self._segments:
            if seg.is_stem:
                key = (seg.sura, seg.verse, seg.word)
                if key not in result or seg.pos not in ("DET", "PREP", "CONJ", "NEG_PART"):
                    result[key] = {
                        "lemma": seg.without_diacritics,
                        "gloss": seg.gloss,
                        "pos": seg.pos,
                        "segment": seg.segmented_text,
                    }
        return result

    def verse_lemmas(self, sura: int, verse: int) -> list[str]:
        lemmas: list[str] = []
        seen_words: set[int] = set()
        for seg in self._segments:
            if seg.sura == sura and seg.verse == verse:
                if seg.word not in seen_words and seg.is_stem:
                    lemmas.append(seg.gloss)
                    seen_words.add(seg.word)
        return lemmas

    def surah_lemmas(self, sura: int) -> dict[int, list[str]]:
        result: dict[int, list[str]] = {}
        for seg in self._segments:
            if seg.sura == sura and seg.is_stem:
                result.setdefault(seg.verse, [])
                if seg.word not in {s.word for s in self._by_word.get((sura, seg.verse, seg.word), []) if s.is_stem and s.segment < seg.segment}:
                    result[seg.verse].append(seg.gloss)
        return result
