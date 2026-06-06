"""
Unified Quran query API — morphology, roots, lemmas, glosses
seamlessly work across any variant via word-to-word alignment.

Usage:
  from jqurantree import Quran

  q = Quran()                          # default: Uthmani Hafs
  q = Quran("indopak")                 # query against Indo-Pak variant
  
  q.roots(2, 255)                      # roots of each word in Sura 2:255
  q.lemmas(3, 30)                      # lemmas
  q.gloss(1, 1)                        # English glosses
  q.pos(112, 1)                        # POS tags
  q.word_count(2)                      # word count in Sura 2
"""

from __future__ import annotations

from dataclasses import dataclass, field
from functools import cached_property
from .morphology import Morphology, Segment
from .alignment_index import Alignment
from .reader import Reader
from .variants import VARIANTS


@dataclass(slots=True)
class WordAnalysis:
    word_index: int
    text: str
    segments: list[Segment] = field(default_factory=list)

    @property
    def gloss(self) -> str:
        stems = [s.gloss for s in self.segments if s.is_stem]
        return " + ".join(stems) if stems else ""

    @property
    def pos_tags(self) -> list[str]:
        return [s.pos for s in self.segments]

    @property
    def lemmas(self) -> list[str]:
        return [s.without_diacritics for s in self.segments if s.is_stem]

    def __repr__(self) -> str:
        return f"WordAnalysis({self.word_index}: {self.text!r} → {self.gloss})"


class Quran:
    def __init__(self, variant: str = "uthmani"):
        if variant not in VARIANTS:
            raise KeyError(f"Unknown variant: {variant}. Available: {list(VARIANTS.keys())}")
        self._variant = variant

    @cached_property
    def _morphology(self) -> Morphology:
        return Morphology()

    @cached_property
    def _alignment(self) -> Alignment:
        return Alignment(self._variant)

    @cached_property
    def _reader(self) -> Reader:
        script = VARIANTS[self._variant].script
        return Reader(script)

    @property
    def variant(self) -> str:
        return self._variant

    def verse_text(self, sura: int, aya: int) -> str:
        return self._reader.verse(sura, aya)

    def words(self, sura: int, aya: int) -> list[WordAnalysis]:
        text = self.verse_text(sura, aya)
        if not text:
            return []
        tokens = text.split()
        alignment = self._alignment.map(sura, aya)
        results: list[WordAnalysis] = []
        for vi, token in enumerate(tokens, start=1):
            ref_wi = alignment.variant_to_ref.get(vi, vi)
            segments = self._morphology.word(sura, aya, ref_wi)
            results.append(WordAnalysis(
                word_index=vi,
                text=token,
                segments=list(segments),
            ))
        return results

    def gloss(self, sura: int, aya: int) -> list[str]:
        return [w.gloss for w in self.words(sura, aya)]

    def lemmas(self, sura: int, aya: int) -> list[str]:
        result: list[str] = []
        for w in self.words(sura, aya):
            result.extend(w.lemmas)
        return result

    def roots(self, sura: int, aya: int) -> list[str]:
        result: list[str] = []
        for w in self.words(sura, aya):
            ref_wi = self._alignment.map(sura, aya).variant_to_ref.get(w.word_index, w.word_index)
            root = self._morphology.root_for(sura, aya, ref_wi)
            if root:
                result.append(root)
        return result

    def pos(self, sura: int, aya: int) -> list[str]:
        result: list[str] = []
        for w in self.words(sura, aya):
            result.extend(w.pos_tags)
        return result

    def analysis(self, sura: int, aya: int) -> list[WordAnalysis]:
        return self.words(sura, aya)

    def word_count(self, sura: int) -> int:
        from .constants import VERSE_COUNTS
        total = 0
        for an in range(1, VERSE_COUNTS[sura - 1] + 1):
            text = self.verse_text(sura, an)
            total += len(text.split())
        return total

    def __repr__(self) -> str:
        return f"Quran({self._variant})"
