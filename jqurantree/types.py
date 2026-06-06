"""
Types — CharacterType, DiacriticTypes, ArabicCharacter.
"""

from __future__ import annotations

from enum import IntEnum, IntFlag


class CharacterType(IntEnum):
    ALIF = 1; BA = 2; TA = 3; THA = 4; JEEM = 5; HHA = 6; KHA = 7
    DAL = 8; THAL = 9; RA = 10; ZAIN = 11; SEEN = 12; SHEEN = 13
    SAD = 14; DAD = 15; TTA = 16; ZZA = 17; AIN = 18; GHAIN = 19
    FA = 20; QAF = 21; KAF = 22; LAM = 23; MEEM = 24; NOON = 25
    HA = 26; WAW = 27; YA = 28; HAMZA = 29; ALIF_HAMZA_ABOVE = 30
    ALIF_HAMZA_BELOW = 31; WAW_HAMZA = 32; YA_HAMZA = 33
    ALIF_MADDA = 34; ALIF_WASLA = 35; ALIF_KHANJAREEYA = 36
    ALIF_MAKSURA = 37; TA_MARBUTA = 38; TATWEEL = 39
    SMALL_WAW = 40; SMALL_YA = 41; SMALL_HIGH_NOON = 42
    SMALL_LOW_MEEM = 43; HIGH_SEEN = 44
    EMPTY_CENTRE_LOW_STOP = 45; EMPTY_CENTRE_HIGH_STOP = 46
    ROUNDED_HIGH_STOP_FILLED = 47; SMALL_HIGH_UPRIGHT_RECT_ZERO = 48
    HIGH_ROUNDED_ZERO = 49


class DiacriticTypes(IntFlag):
    NONE = 0
    FATHA = 1 << 0; DAMMA = 1 << 1; KASRA = 1 << 2
    SUKUN = 1 << 3; SHADDA = 1 << 4
    TANWEEN_FATHA = 1 << 5; TANWEEN_DAMMA = 1 << 6
    TANWEEN_KASRA = 1 << 7


class ArabicCharacter:
    __slots__ = ('_char_type', '_diacritics')

    def __init__(self, char_type: CharacterType, diacritics: DiacriticTypes = DiacriticTypes.NONE):
        self._char_type = char_type
        self._diacritics = diacritics

    @property
    def char_type(self) -> CharacterType: return self._char_type
    @property
    def diacritics(self) -> DiacriticTypes: return self._diacritics

    def __hash__(self): return hash((self._char_type, self._diacritics))
    def __eq__(self, o): return isinstance(o, ArabicCharacter) and self._char_type == o._char_type and self._diacritics == o._diacritics
    def __repr__(self): return f"ArabicCharacter({self._char_type.name}, {self._diacritics!r})"
