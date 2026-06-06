"""Test constants — Surah enum, verse counts, names."""

from jqurantree.constants import (
    Surah, VERSE_COUNTS, CHAPTER_COUNT, TOTAL_VERSES, surah_name, verse_count
)


def test_chapter_count():
    assert CHAPTER_COUNT == 114

def test_total_verses():
    assert TOTAL_VERSES == 6236

def test_verse_counts_length():
    assert len(VERSE_COUNTS) == 114

def test_verse_counts_reference():
    ref = (7, 286, 200, 176, 120, 165, 206, 75, 129, 109,
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
           5, 4, 5, 6)
    assert VERSE_COUNTS == ref

def test_surah_enum():
    assert Surah.AL_FATIHA == 1
    assert Surah.AL_BAQARAH == 2
    assert Surah.AL_IMRAN == 3
    assert Surah.AR_RAHMAN == 55
    assert Surah.AL_IKHLAS == 112
    assert Surah.AN_NAS == 114
    assert len(Surah) == 114

def test_verse_count_function():
    assert verse_count(1) == 7
    assert verse_count(2) == 286
    assert verse_count(114) == 6
    for i in range(1, 115):
        assert verse_count(i) == VERSE_COUNTS[i - 1]

def test_surah_names():
    assert surah_name(1) == "Al-Fatiha"
    assert surah_name(2) == "Al-Baqarah"
    assert surah_name(55) == "Ar-Rahman"
    assert surah_name(114) == "An-Nas"
    assert "Surah" in surah_name(999)
