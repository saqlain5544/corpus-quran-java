"""
Alignment — NFC normalization + character equivalence for cross-variant comparison.
"""

from __future__ import annotations

import unicodedata
from dataclasses import dataclass


@dataclass(slots=True, frozen=True)
class Align:
    nfc: bool = False
    ignore_diacritics: bool = False
    wasla_to_alif: bool = False
    dagger_to_alif: bool = False

    @staticmethod
    def exact() -> "Align": return Align()

    @staticmethod
    def canonical() -> "Align": return Align(nfc=True)

    @staticmethod
    def visual() -> "Align": return Align(nfc=True, wasla_to_alif=True, dagger_to_alif=True)

    @staticmethod
    def phonetic() -> "Align": return Align(nfc=True, ignore_diacritics=True,
                                            wasla_to_alif=True, dagger_to_alif=True)

    @staticmethod
    def structural() -> "Align": return Align(nfc=True, wasla_to_alif=True, dagger_to_alif=True)


_DIAC_CP = frozenset({0x064B, 0x064C, 0x064D, 0x064E, 0x064F, 0x0650, 0x0651, 0x0652})


def align(text: str, config: Align) -> str:
    if not text: return ""
    result = text
    if config.nfc:
        result = unicodedata.normalize("NFC", result)
    if config.ignore_diacritics:
        result = "".join(ch for ch in result if ord(ch) not in _DIAC_CP)
    if config.wasla_to_alif:
        result = result.replace("\u0671", "\u0627")
    if config.dagger_to_alif:
        result = result.replace("\u0670", "\u0627")
    return result
