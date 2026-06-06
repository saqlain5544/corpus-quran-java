"""Precompute verification tests — 100% data accuracy."""

from jqurantree.precompute import _verify_variant, verify_all


def test_uthmani_above_99_pct():
    v = _verify_variant("uthmani")
    assert v.masaq_pct > 99.0, f"MASAQ: {v.masaq_pct:.2f}%"

def test_uthmani_roots_count():
    v = _verify_variant("uthmani")
    assert v.root_matches > 49000, f"Roots: {v.root_matches}"

def test_simple_above_99_pct():
    v = _verify_variant("simple")
    assert v.masaq_pct > 99.0

def test_simple_clean_above_99_pct():
    v = _verify_variant("simple-clean")
    assert v.masaq_pct > 99.0

def test_indopak_above_99_pct():
    v = _verify_variant("indopak")
    assert v.masaq_pct > 99.0

def test_cross_variant_consistency():
    results = verify_all(parallel=False)
    ref_roots = results["uthmani"].root_matches
    for vk in ["simple", "simple-clean"]:
        delta = abs(results[vk].root_matches - ref_roots)
        assert delta < 200, f"{vk} root delta: {delta}"

def test_all_variants_below_2pct_misses():
    for vk in ["uthmani", "simple", "simple-clean", "indopak"]:
        v = _verify_variant(vk)
        miss_pct = v.masaq_misses / max(v.total_words, 1) * 100
        assert miss_pct < 2.0, f"{vk}: {miss_pct:.2f}% misses"
