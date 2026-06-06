"""
Encoding — Unicode, Buckwalter, Simple codecs with mapping tables.
"""

from __future__ import annotations

import unicodedata

from .types import CharacterType, DiacriticTypes, ArabicCharacter

# ── Unicode mapping tables ──

CP_TO_CHAR: dict[int, CharacterType] = {
    0x0627: CharacterType.ALIF, 0x0628: CharacterType.BA, 0x062A: CharacterType.TA,
    0x062B: CharacterType.THA, 0x062C: CharacterType.JEEM, 0x062D: CharacterType.HHA,
    0x062E: CharacterType.KHA, 0x062F: CharacterType.DAL, 0x0630: CharacterType.THAL,
    0x0631: CharacterType.RA, 0x0632: CharacterType.ZAIN, 0x0633: CharacterType.SEEN,
    0x0634: CharacterType.SHEEN, 0x0635: CharacterType.SAD, 0x0636: CharacterType.DAD,
    0x0637: CharacterType.TTA, 0x0638: CharacterType.ZZA, 0x0639: CharacterType.AIN,
    0x063A: CharacterType.GHAIN, 0x0641: CharacterType.FA, 0x0642: CharacterType.QAF,
    0x0643: CharacterType.KAF, 0x0644: CharacterType.LAM, 0x0645: CharacterType.MEEM,
    0x0646: CharacterType.NOON, 0x0647: CharacterType.HA, 0x0648: CharacterType.WAW,
    0x0649: CharacterType.ALIF_MAKSURA, 0x064A: CharacterType.YA,
    0x0621: CharacterType.HAMZA, 0x0622: CharacterType.ALIF_MADDA,
    0x0623: CharacterType.ALIF_HAMZA_ABOVE, 0x0624: CharacterType.WAW_HAMZA,
    0x0625: CharacterType.ALIF_HAMZA_BELOW, 0x0626: CharacterType.YA_HAMZA,
    0x0629: CharacterType.TA_MARBUTA, 0x0640: CharacterType.TATWEEL,
    0x0671: CharacterType.ALIF_WASLA,
    0x06E5: CharacterType.SMALL_WAW, 0x06E6: CharacterType.SMALL_YA,
    0x06E8: CharacterType.SMALL_HIGH_NOON, 0x06ED: CharacterType.SMALL_LOW_MEEM,
    0x06DC: CharacterType.HIGH_SEEN, 0x06DF: CharacterType.HIGH_ROUNDED_ZERO,
    0x06E0: CharacterType.SMALL_HIGH_UPRIGHT_RECT_ZERO,
    0x06EA: CharacterType.EMPTY_CENTRE_LOW_STOP,
    0x06EB: CharacterType.EMPTY_CENTRE_HIGH_STOP,
    0x06EC: CharacterType.ROUNDED_HIGH_STOP_FILLED,
}

CP_TO_DIACRITIC: dict[int, DiacriticTypes] = {
    0x064E: DiacriticTypes.FATHA, 0x064F: DiacriticTypes.DAMMA,
    0x0650: DiacriticTypes.KASRA, 0x0651: DiacriticTypes.SHADDA,
    0x0652: DiacriticTypes.SUKUN,
    0x064B: DiacriticTypes.TANWEEN_FATHA, 0x064C: DiacriticTypes.TANWEEN_DAMMA,
    0x064D: DiacriticTypes.TANWEEN_KASRA,
}

CHAR_TO_CP: dict[CharacterType, int] = {v: k for k, v in CP_TO_CHAR.items()}

DIACRITIC_TO_CP: dict[DiacriticTypes, int] = {
    DiacriticTypes.FATHA: 0x064E, DiacriticTypes.DAMMA: 0x064F,
    DiacriticTypes.KASRA: 0x0650, DiacriticTypes.SHADDA: 0x0651,
    DiacriticTypes.SUKUN: 0x0652,
    DiacriticTypes.TANWEEN_FATHA: 0x064B, DiacriticTypes.TANWEEN_DAMMA: 0x064C,
    DiacriticTypes.TANWEEN_KASRA: 0x064D,
}

DIACRITIC_ORDER = [DiacriticTypes.SHADDA, DiacriticTypes.FATHA, DiacriticTypes.DAMMA,
                    DiacriticTypes.KASRA, DiacriticTypes.TANWEEN_FATHA,
                    DiacriticTypes.TANWEEN_DAMMA, DiacriticTypes.TANWEEN_KASRA,
                    DiacriticTypes.SUKUN]

DIACRITIC_CPS = frozenset(CP_TO_DIACRITIC.keys()) | {0x0670}
ALIF_KHANJAREEYA_CP = 0x0670


# ── Decode / Encode ──

def decode(text: str, normalize: bool = True) -> list[ArabicCharacter]:
    if not text: return []
    if normalize: text = unicodedata.normalize("NFC", text)
    result: list[ArabicCharacter] = []
    i, n = 0, len(text)
    while i < n:
        cp = ord(text[i])
        if cp in DIACRITIC_CPS and result:
            if cp == ALIF_KHANJAREEYA_CP:
                result.append(ArabicCharacter(CharacterType.ALIF_KHANJAREEYA))
            elif cp in CP_TO_DIACRITIC:
                prev = result[-1]
                result[-1] = ArabicCharacter(prev.char_type, prev.diacritics | CP_TO_DIACRITIC[cp])
            i += 1
        elif cp in CP_TO_CHAR:
            result.append(ArabicCharacter(CP_TO_CHAR[cp]))
            i += 1
        else:
            i += 1
    return result


def encode(characters: list[ArabicCharacter]) -> str:
    parts: list[str] = []
    for ac in characters:
        ct = ac.char_type
        if ct == CharacterType.ALIF_KHANJAREEYA:
            parts.append(chr(ALIF_KHANJAREEYA_CP))
            continue
        base = CHAR_TO_CP.get(ct)
        if base is not None: parts.append(chr(base))
        for dt in DIACRITIC_ORDER:
            if ac.diacritics & dt:
                dc = DIACRITIC_TO_CP.get(dt)
                if dc is not None: parts.append(chr(dc))
    return "".join(parts)


# ── Buckwalter ──

_CHAR_TO_BW: dict[CharacterType, str] = {
    CharacterType.HAMZA: "'", CharacterType.ALIF_MADDA: "|",
    CharacterType.ALIF_HAMZA_ABOVE: ">", CharacterType.WAW_HAMZA: "&",
    CharacterType.ALIF_HAMZA_BELOW: "<", CharacterType.YA_HAMZA: "}",
    CharacterType.ALIF: "A", CharacterType.BA: "b", CharacterType.TA_MARBUTA: "p",
    CharacterType.TA: "t", CharacterType.THA: "v", CharacterType.JEEM: "j",
    CharacterType.HHA: "H", CharacterType.KHA: "x", CharacterType.DAL: "d",
    CharacterType.THAL: "*", CharacterType.RA: "r", CharacterType.ZAIN: "z",
    CharacterType.SEEN: "s", CharacterType.SHEEN: "$", CharacterType.SAD: "S",
    CharacterType.DAD: "D", CharacterType.TTA: "T", CharacterType.ZZA: "Z",
    CharacterType.AIN: "E", CharacterType.GHAIN: "g", CharacterType.FA: "f",
    CharacterType.QAF: "q", CharacterType.KAF: "k", CharacterType.LAM: "l",
    CharacterType.MEEM: "m", CharacterType.NOON: "n", CharacterType.HA: "h",
    CharacterType.WAW: "w", CharacterType.ALIF_MAKSURA: "Y", CharacterType.YA: "y",
    CharacterType.ALIF_WASLA: "{", CharacterType.TATWEEL: "_",
    CharacterType.ALIF_KHANJAREEYA: "`",
}

_DIACRITIC_TO_BW: dict[DiacriticTypes, str] = {
    DiacriticTypes.FATHA: "a", DiacriticTypes.DAMMA: "u",
    DiacriticTypes.KASRA: "i", DiacriticTypes.SUKUN: "o",
    DiacriticTypes.SHADDA: "~", DiacriticTypes.TANWEEN_FATHA: "F",
    DiacriticTypes.TANWEEN_DAMMA: "N", DiacriticTypes.TANWEEN_KASRA: "K",
}

_BW_REVERSE: dict[str, CharacterType] = {v: k for k, v in _CHAR_TO_BW.items()}
_BW_DIAC_REVERSE: dict[str, DiacriticTypes] = {v: k for k, v in _DIACRITIC_TO_BW.items()}


def encode_buckwalter(characters: list[ArabicCharacter]) -> str:
    parts: list[str] = []
    for ac in characters:
        bw = _CHAR_TO_BW.get(ac.char_type, "")
        parts.append(bw)
        for dt in DIACRITIC_ORDER:
            if ac.diacritics & dt:
                bd = _DIACRITIC_TO_BW.get(dt, "")
                if bd: parts.append(bd)
    return "".join(parts)


def decode_buckwalter(text: str) -> list[ArabicCharacter]:
    result: list[ArabicCharacter] = []
    i = 0
    while i < len(text):
        ch = text[i]
        if ch in _BW_REVERSE:
            result.append(ArabicCharacter(_BW_REVERSE[ch]))
        elif result and ch in _BW_DIAC_REVERSE:
            prev = result[-1]
            result[-1] = ArabicCharacter(prev.char_type, prev.diacritics | _BW_DIAC_REVERSE[ch])
        i += 1
    return result


# ── Simple encoder ──

_STRIP_MAP: dict[CharacterType, CharacterType] = {
    CharacterType.ALIF_HAMZA_ABOVE: CharacterType.ALIF,
    CharacterType.ALIF_HAMZA_BELOW: CharacterType.ALIF,
    CharacterType.ALIF_WASLA: CharacterType.ALIF,
    CharacterType.ALIF_MADDA: CharacterType.ALIF,
    CharacterType.ALIF_MAKSURA: CharacterType.YA,
    CharacterType.WAW_HAMZA: CharacterType.WAW,
    CharacterType.YA_HAMZA: CharacterType.YA,
    CharacterType.TA_MARBUTA: CharacterType.HA,
}

_STRIP_SKIP = frozenset({
    CharacterType.ALIF_KHANJAREEYA, CharacterType.TATWEEL,
    CharacterType.SMALL_WAW, CharacterType.SMALL_YA,
    CharacterType.SMALL_HIGH_NOON, CharacterType.SMALL_LOW_MEEM,
    CharacterType.HIGH_SEEN, CharacterType.HIGH_ROUNDED_ZERO,
    CharacterType.SMALL_HIGH_UPRIGHT_RECT_ZERO,
    CharacterType.EMPTY_CENTRE_LOW_STOP, CharacterType.EMPTY_CENTRE_HIGH_STOP,
    CharacterType.ROUNDED_HIGH_STOP_FILLED,
})


def encode_simple(characters: list[ArabicCharacter]) -> list[ArabicCharacter]:
    result: list[ArabicCharacter] = []
    for ac in characters:
        ct = ac.char_type
        if ct in _STRIP_SKIP: continue
        stripped = _STRIP_MAP.get(ct, ct)
        result.append(ArabicCharacter(stripped, DiacriticTypes.NONE))
    return result
