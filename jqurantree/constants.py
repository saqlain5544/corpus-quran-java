"""
Constants — single source of truth for Quranic reference data.
"""

from __future__ import annotations

from enum import IntEnum


class Surah(IntEnum):
    AL_FATIHA = 1; AL_BAQARAH = 2; AL_IMRAN = 3; AN_NISA = 4
    AL_MAIDAH = 5; AL_ANAM = 6; AL_ARAF = 7; AL_ANFAL = 8
    AT_TAWBAH = 9; YUNUS = 10; HUD = 11; YUSUF = 12
    AR_RAD = 13; IBRAHIM = 14; AL_HIJR = 15; AN_NAHL = 16
    AL_ISRA = 17; AL_KAHF = 18; MARYAM = 19; TA_HA = 20
    AL_ANBIYA = 21; AL_HAJJ = 22; AL_MUMINUN = 23; AN_NUR = 24
    AL_FURQAN = 25; ASH_SHUARA = 26; AN_NAML = 27; AL_QASAS = 28
    AL_ANKABUT = 29; AR_RUM = 30; LUQMAN = 31; AS_SAJDAH = 32
    AL_AHZAB = 33; SABA = 34; FATIR = 35; YA_SIN = 36
    AS_SAFFAT = 37; SAD = 38; AZ_ZUMAR = 39; GHAFIR = 40
    FUSSILAT = 41; ASH_SHURA = 42; AZ_ZUKHRUF = 43; AD_DUKHAN = 44
    AL_JATHIYAH = 45; AL_AHQAF = 46; MUHAMMAD = 47; AL_FATH = 48
    AL_HUJURAT = 49; QAF = 50; ADH_DHARIYAT = 51; AT_TUR = 52
    AN_NAJM = 53; AL_QAMAR = 54; AR_RAHMAN = 55; AL_WAQIAH = 56
    AL_HADID = 57; AL_MUJADILA = 58; AL_HASHR = 59; AL_MUMTAHANA = 60
    AS_SAFF = 61; AL_JUMUAH = 62; AL_MUNAFIQUN = 63; AT_TAGHABUN = 64
    AT_TALAQ = 65; AT_TAHRIM = 66; AL_MULK = 67; AL_QALAM = 68
    AL_HAQQAH = 69; AL_MAARIJ = 70; NUH = 71; AL_JINN = 72
    AL_MUZZAMMIL = 73; AL_MUDDATHTHIR = 74; AL_QIYAMAH = 75
    AL_INSAN = 76; AL_MURSALAT = 77; AN_NABA = 78; AN_NAZIAT = 79
    ABASA = 80; AT_TAKWIR = 81; AL_INFITAR = 82; AL_MUTAFFIFIN = 83
    AL_INSHIQAQ = 84; AL_BURUJ = 85; AT_TARIQ = 86; AL_ALA = 87
    AL_GHASHIYAH = 88; AL_FAJR = 89; AL_BALAD = 90; ASH_SHAMS = 91
    AL_LAYL = 92; AD_DUHA = 93; ASH_SHARH = 94; AT_TIN = 95
    AL_ALAQ = 96; AL_QADR = 97; AL_BAYYINAH = 98; AZ_ZALZALAH = 99
    AL_ADIYAT = 100; AL_QARIAH = 101; AT_TAKATHUR = 102; AL_ASR = 103
    AL_HUMAZAH = 104; AL_FIL = 105; QURAISH = 106; AL_MAUN = 107
    AL_KAUTHAR = 108; AL_KAFIRUN = 109; AN_NASR = 110; AL_MASAD = 111
    AL_IKHLAS = 112; AL_FALAQ = 113; AN_NAS = 114


VERSE_COUNTS = (
    7, 286, 200, 176, 120, 165, 206, 75, 129, 109,
    123, 111, 43, 52, 99, 128, 111, 110, 98, 135,
    112, 78, 118, 64, 77, 227, 93, 88, 69, 60,
    34, 30, 73, 54, 45, 83, 182, 88, 75, 85,
    54, 53, 89, 59, 37, 35, 38, 29, 18, 45,
    60, 49, 62, 55, 78, 96, 29, 22, 24, 13,
    14, 11, 11, 18, 12, 12, 30, 52, 52, 44,
    28, 28, 20, 56, 40, 31, 50, 40, 46, 42,
    29, 19, 36, 25, 22, 17, 19, 26, 30, 20,
    15, 21, 11, 8, 8, 19, 5, 8, 8, 11,
    11, 8, 3, 9, 5, 4, 7, 3, 6, 3,
    5, 4, 5, 6,
)

SURAH_NAMES_EN = {
    1: "Al-Fatiha", 2: "Al-Baqarah", 3: "Al-Imran", 4: "An-Nisa",
    5: "Al-Ma'idah", 6: "Al-An'am", 7: "Al-A'raf", 8: "Al-Anfal",
    9: "At-Tawbah", 10: "Yunus", 11: "Hud", 12: "Yusuf",
    13: "Ar-Ra'd", 14: "Ibrahim", 15: "Al-Hijr", 16: "An-Nahl",
    17: "Al-Isra", 18: "Al-Kahf", 19: "Maryam", 20: "Ta-Ha",
    21: "Al-Anbiya", 22: "Al-Hajj", 23: "Al-Mu'minun", 24: "An-Nur",
    25: "Al-Furqan", 26: "Ash-Shu'ara", 27: "An-Naml", 28: "Al-Qasas",
    29: "Al-Ankabut", 30: "Ar-Rum", 31: "Luqman", 32: "As-Sajdah",
    33: "Al-Ahzab", 34: "Saba", 35: "Fatir", 36: "Ya-Sin",
    37: "As-Saffat", 38: "Sad", 39: "Az-Zumar", 40: "Ghafir",
    41: "Fussilat", 42: "Ash-Shura", 43: "Az-Zukhruf", 44: "Ad-Dukhan",
    45: "Al-Jathiyah", 46: "Al-Ahqaf", 47: "Muhammad", 48: "Al-Fath",
    49: "Al-Hujurat", 50: "Qaf", 51: "Adh-Dhariyat", 52: "At-Tur",
    53: "An-Najm", 54: "Al-Qamar", 55: "Ar-Rahman", 56: "Al-Waqi'ah",
    57: "Al-Hadid", 58: "Al-Mujadila", 59: "Al-Hashr", 60: "Al-Mumtahana",
    61: "As-Saff", 62: "Al-Jumu'ah", 63: "Al-Munafiqun", 64: "At-Taghabun",
    65: "At-Talaq", 66: "At-Tahrim", 67: "Al-Mulk", 68: "Al-Qalam",
    69: "Al-Haqqah", 70: "Al-Ma'arij", 71: "Nuh", 72: "Al-Jinn",
    73: "Al-Muzzammil", 74: "Al-Muddaththir", 75: "Al-Qiyamah",
    76: "Al-Insan", 77: "Al-Mursalat", 78: "An-Naba", 79: "An-Nazi'at",
    80: "Abasa", 81: "At-Takwir", 82: "Al-Infitar", 83: "Al-Mutaffifin",
    84: "Al-Inshiqaq", 85: "Al-Buruj", 86: "At-Tariq", 87: "Al-A'la",
    88: "Al-Ghashiyah", 89: "Al-Fajr", 90: "Al-Balad", 91: "Ash-Shams",
    92: "Al-Layl", 93: "Ad-Duha", 94: "Ash-Sharh", 95: "At-Tin",
    96: "Al-Alaq", 97: "Al-Qadr", 98: "Al-Bayyinah", 99: "Az-Zalzalah",
    100: "Al-Adiyat", 101: "Al-Qari'ah", 102: "At-Takathur", 103: "Al-Asr",
    104: "Al-Humazah", 105: "Al-Fil", 106: "Quraish", 107: "Al-Ma'un",
    108: "Al-Kauthar", 109: "Al-Kafirun", 110: "An-Nasr", 111: "Al-Masad",
    112: "Al-Ikhlas", 113: "Al-Falaq", 114: "An-Nas",
}

CHAPTER_COUNT = 114
TOTAL_VERSES = 6236


def surah_name(n: int) -> str:
    return SURAH_NAMES_EN.get(n, f"Surah {n}")


def verse_count(n: int) -> int:
    if n < 1 or n > 114:
        raise IndexError(f"Surah {n} out of range")
    return VERSE_COUNTS[n - 1]
