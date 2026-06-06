"""Pre-processing encoding validation — 100% accuracy per variant."""

from jqurantree.validator import validate_all_variants, CharClass


def test_all_variants_validate():
    reports = validate_all_variants()
    assert len(reports) >= 5, f"Expected at least 5 variants, got {len(reports)}"
    for name, report in reports.items():
        assert report.is_clean, f"{name}: {report.issues}"


def test_uthmani_no_unexpected():
    reports = validate_all_variants()
    r = reports.get("uthmani")
    assert r is not None
    assert r.classes.get(CharClass.ALIF_WASLA, 0) > 10000, "Uthmani must have Alif Wasla"
    assert r.classes.get(CharClass.DAGGER_ALIF, 0) > 9000, "Uthmani must have Dagger Alif"


def test_simple_clean_no_diacritics():
    reports = validate_all_variants()
    r = reports.get("simple-clean")
    assert r is not None
    assert r.classes.get(CharClass.DIACRITIC, 0) == 0, "Simple-Clean must have zero diacritics"
    assert r.classes.get(CharClass.ARABIC_LETTER, 0) > 300000


def test_simple_has_diacritics():
    reports = validate_all_variants()
    r = reports.get("simple")
    assert r is not None
    assert r.classes.get(CharClass.DIACRITIC, 0) > 200000


def test_all_variants_arabic_letters_dominant():
    reports = validate_all_variants()
    for name, r in reports.items():
        letters = r.classes.get(CharClass.ARABIC_LETTER, 0)
        total = r.total_chars
        assert letters / total > 0.40, f"{name}: letters only {letters/total*100:.1f}%"


def test_indopak_has_rlm():
    reports = validate_all_variants()
    r = reports.get("indopak")
    if r is None:
        # May not be downloaded yet
        return
    assert r.classes.get(CharClass.RLM, 0) > 100, "Indo-Pak must have RLM markers"
    assert r.classes.get(CharClass.EXTENDED_QURANIC, 0) > 1000, "Indo-Pak must have extended Quranic marks"


def test_warsh_has_special_marks():
    reports = validate_all_variants()
    r = reports.get("warsh")
    if r is None:
        return
    assert r.classes.get(CharClass.EXTENDED_QURANIC, 0) > 500, "Warsh must have extended diacritics"
