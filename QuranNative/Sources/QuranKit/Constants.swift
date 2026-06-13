public enum Surah: Int, CaseIterable, Sendable {
    case AL_FATIHA = 1
    case AL_BAQARAH = 2
    case AL_IMRAN = 3
    case AN_NISA = 4
    case AL_MAIDAH = 5
    case AL_ANAM = 6
    case AL_ARAF = 7
    case AL_ANFAL = 8
    case AT_TAWBAH = 9
    case YUNUS = 10
    case HUD = 11
    case YUSUF = 12
    case AR_RAD = 13
    case IBRAHIM = 14
    case AL_HIJR = 15
    case AN_NAHL = 16
    case AL_ISRA = 17
    case AL_KAHF = 18
    case MARYAM = 19
    case TA_HA = 20
    case AL_ANBIYA = 21
    case AL_HAJJ = 22
    case AL_MUMINUN = 23
    case AN_NUR = 24
    case AL_FURQAN = 25
    case ASH_SHUARA = 26
    case AN_NAML = 27
    case AL_QASAS = 28
    case AL_ANKABUT = 29
    case AR_RUM = 30
    case LUQMAN = 31
    case AS_SAJDAH = 32
    case AL_AHZAB = 33
    case SABA = 34
    case FATIR = 35
    case YA_SIN = 36
    case AS_SAFFAT = 37
    case SAD = 38
    case AZ_ZUMAR = 39
    case GHAFIR = 40
    case FUSSILAT = 41
    case ASH_SHURA = 42
    case AZ_ZUKHRUF = 43
    case AD_DUKHAN = 44
    case AL_JATHIYAH = 45
    case AL_AHQAF = 46
    case MUHAMMAD = 47
    case AL_FATH = 48
    case AL_HUJURAT = 49
    case QAF = 50
    case ADH_DHARIYAT = 51
    case AT_TUR = 52
    case AN_NAJM = 53
    case AL_QAMAR = 54
    case AR_RAHMAN = 55
    case AL_WAQIAH = 56
    case AL_HADID = 57
    case AL_MUJADILAH = 58
    case AL_HASHR = 59
    case AL_MUMTAHANAH = 60
    case AS_SAFF = 61
    case AL_JUMUAH = 62
    case AL_MUNAFIQUN = 63
    case AT_TAGHABUN = 64
    case AT_TALAQ = 65
    case AT_TAHRIM = 66
    case AL_MULK = 67
    case AL_QALAM = 68
    case AL_HAQQAH = 69
    case AL_MAARIJ = 70
    case NUH = 71
    case AL_JINN = 72
    case AL_MUZZAMMIL = 73
    case AL_MUDDATHTHIR = 74
    case AL_QIYAMAH = 75
    case AL_INSAN = 76
    case AL_MURSALAT = 77
    case AN_NABA = 78
    case AN_NAZIAT = 79
    case ABASA = 80
    case AT_TAKWIR = 81
    case AL_INFITAR = 82
    case AL_MUTAFFIFIN = 83
    case AL_INSHIQAQ = 84
    case AL_BURUJ = 85
    case AT_TARIQ = 86
    case AL_ALA = 87
    case AL_GHASHIYAH = 88
    case AL_FAJR = 89
    case AL_BALAD = 90
    case ASH_SHAMS = 91
    case AL_LAYL = 92
    case AD_DUHA = 93
    case ASH_SHARH = 94
    case AT_TIN = 95
    case AL_ALAQ = 96
    case AL_QADR = 97
    case AL_BAYYINAH = 98
    case AZ_ZALZALAH = 99
    case AL_ADIYAT = 100
    case AL_QARIAH = 101
    case AT_TAKATHUR = 102
    case AL_ASR = 103
    case AL_HUMAZAH = 104
    case AL_FIL = 105
    case QURAYSH = 106
    case AL_MAUN = 107
    case AL_KAWTHAR = 108
    case AL_KAFIRUN = 109
    case AN_NASR = 110
    case AL_MASAD = 111
    case AL_IKHLAS = 112
    case AL_FALAQ = 113
    case AN_NAS = 114
}

public let VERSE_COUNTS: [Int] = [
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
]

public let SURAH_NAMES_EN: [Int: String] = [
    1: "Al-Fatiḥah", 2: "Al-Baqarah", 3: "Āl ʿImrān", 4: "An-Nisāʾ",
    5: "Al-Māʾidah", 6: "Al-Anʿām", 7: "Al-Aʿrāf", 8: "Al-Anfāl",
    9: "At-Tawbah", 10: "Yūnus", 11: "Hūd", 12: "Yūsuf",
    13: "Ar-Raʿd", 14: "Ibrāhīm", 15: "Al-Ḥijr", 16: "An-Naḥl",
    17: "Al-Isrāʾ", 18: "Al-Kahf", 19: "Maryam", 20: "Ṭā Hā",
    21: "Al-Anbiyāʾ", 22: "Al-Ḥajj", 23: "Al-Muʾminūn", 24: "An-Nūr",
    25: "Al-Furqān", 26: "Ash-Shuʿarāʾ", 27: "An-Naml", 28: "Al-Qaṣaṣ",
    29: "Al-ʿAnkabūt", 30: "Ar-Rūm", 31: "Luqmān", 32: "As-Sajdah",
    33: "Al-Aḥzāb", 34: "Sabaʾ", 35: "Fāṭir", 36: "Yā Sīn",
    37: "Aṣ-Ṣāffāt", 38: "Ṣād", 39: "Az-Zumar", 40: "Ghāfir",
    41: "Fuṣṣilat", 42: "Ash-Shūrā", 43: "Az-Zukhruf", 44: "Ad-Dukhān",
    45: "Al-Jāthiyah", 46: "Al-Aḥqāf", 47: "Muḥammad", 48: "Al-Fatḥ",
    49: "Al-Ḥujurāt", 50: "Qāf", 51: "Adh-Dhāriyāt", 52: "Aṭ-Ṭūr",
    53: "An-Najm", 54: "Al-Qamar", 55: "Ar-Raḥmān", 56: "Al-Wāqiʿah",
    57: "Al-Ḥadīd", 58: "Al-Mujādilah", 59: "Al-Ḥashr", 60: "Al-Mumtaḥanah",
    61: "Aṣ-Ṣaff", 62: "Al-Jumuʿah", 63: "Al-Munāfiqūn", 64: "At-Taghābun",
    65: "Aṭ-Ṭalāq", 66: "At-Taḥrīm", 67: "Al-Mulk", 68: "Al-Qalam",
    69: "Al-Ḥāqqah", 70: "Al-Maʿārij", 71: "Nūḥ", 72: "Al-Jinn",
    73: "Al-Muzzammil", 74: "Al-Muddaththir", 75: "Al-Qiyāmah", 76: "Al-Insān",
    77: "Al-Mursalāt", 78: "An-Nabaʾ", 79: "An-Nāziʿāt", 80: "ʿAbasa",
    81: "At-Takwīr", 82: "Al-Infiṭār", 83: "Al-Muṭaffifīn", 84: "Al-Inshiqāq",
    85: "Al-Burūj", 86: "Aṭ-Ṭāriq", 87: "Al-Aʿlā", 88: "Al-Ghāshiyah",
    89: "Al-Fajr", 90: "Al-Balad", 91: "Ash-Shams", 92: "Al-Layl",
    93: "Aḍ-Ḍuḥā", 94: "Ash-Sharḥ", 95: "At-Tīn", 96: "Al-ʿAlaq",
    97: "Al-Qadr", 98: "Al-Bayyinah", 99: "Az-Zalzalah", 100: "Al-ʿĀdiyāt",
    101: "Al-Qāriʿah", 102: "At-Takāthur", 103: "Al-ʿAṣr", 104: "Al-Humazah",
    105: "Al-Fīl", 106: "Quraysh", 107: "Al-Māʿūn", 108: "Al-Kawthar",
    109: "Al-Kāfirūn", 110: "An-Naṣr", 111: "Al-Masad", 112: "Al-Ikhlāṣ",
    113: "Al-Falaq", 114: "An-Nās",
]

public func surahName(_ n: Int) -> String {
    SURAH_NAMES_EN[n] ?? "Surah \(n)"
}

public func verseCount(_ n: Int) -> Int {
    guard n >= 1, n <= 114 else { return 0 }
    return VERSE_COUNTS[n - 1]
}

public func isValid(sura: Int, verse: Int) -> Bool {
    guard sura >= 1, sura <= 114 else { return false }
    return verse >= 1 && verse <= VERSE_COUNTS[sura - 1]
}
