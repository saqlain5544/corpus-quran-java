"""Test variants — loading, diff, cross-variant consistency."""

from jqurantree import VARIANTS, load_variant, diff_variants, diff_variants_lemmas, Align


def test_registry_count():
    assert len(VARIANTS) >= 8

def test_registry_has_key_variants():
    for k in ['uthmani', 'simple', 'simple-clean', 'indopak', 'warsh', 'qaloon']:
        assert k in VARIANTS

def test_load_uthmani():
    data = load_variant('uthmani')
    assert data['variant_key'] == 'uthmani'
    assert len(data['suras']) == 114
    assert sum(len(s['ayas']) for s in data['suras']) == 6236

def test_load_simple():
    data = load_variant('simple')
    assert len(data['suras']) == 114

def test_load_simple_clean():
    data = load_variant('simple-clean')
    assert len(data['suras']) == 114

def test_byte_diff_self():
    d = diff_variants('uthmani', 'uthmani', Align.exact())
    assert d.identical == d.total
    assert d.token_changes == 0

def test_byte_diff_visual():
    d = diff_variants('uthmani', 'simple', Align.visual())
    assert d.total == 6236
    assert d.identical > 1000

def test_lemma_diff_self():
    d = diff_variants_lemmas('uthmani', 'uthmani')
    assert d['same_words'] == d['total_verses']

def test_lemma_diff_simple_clean():
    d = diff_variants_lemmas('simple', 'simple-clean')
    assert d['word_similarity_pct'] == 100.0

def test_lemma_diff_uthmani_simple():
    d = diff_variants_lemmas('uthmani', 'simple')
    assert d['word_similarity_pct'] > 90.0
