"""Test encoding — Unicode round-trip, Buckwalter, Simple."""

from jqurantree.encoding import decode, encode, encode_buckwalter, decode_buckwalter, encode_simple
from jqurantree.types import CharacterType, DiacriticTypes


def test_unicode_round_trip():
    for text in ['بِسْمِ', 'ٱلْحَمْدُ', 'شَمْس', 'قَمَر']:
        chars = decode(text, normalize=False)
        reencoded = encode(chars)
        assert text == reencoded, f"Round-trip failed: {repr(text)} != {repr(reencoded)}"

def test_decode_with_spaces():
    chars = decode('بِسْمِ ٱللَّهِ', normalize=False)
    assert len(chars) > 3

def test_unicode_nfc_normalization():
    text = 'رَبِّكَ'
    chars = decode(text, normalize=True)
    result = encode(chars)
    assert len(result) == len(text)

def test_decode_preserves_chars():
    chars = decode('بِسْمِ', normalize=False)
    assert len(chars) == 3
    assert chars[0].char_type == CharacterType.BA
    assert chars[0].diacritics == DiacriticTypes.KASRA
    assert chars[1].char_type == CharacterType.SEEN
    assert chars[1].diacritics == DiacriticTypes.SUKUN

def test_buckwalter_output():
    chars = decode('بِسْمِ', normalize=False)
    bw = encode_buckwalter(list(chars))
    assert len(bw) > 0

def test_buckwalter_round_trip():
    original = 'بِسْمِ'
    chars = decode(original, normalize=False)
    bw = encode_buckwalter(list(chars))
    decoded_bw = decode_buckwalter(bw)
    result = encode(decoded_bw)
    assert result == original

def test_simple_encoder():
    chars = decode('بِسْمِ ٱللَّهِ', normalize=False)
    simple = encode_simple(list(chars))
    assert len(simple) > 0
    assert all(c.diacritics == DiacriticTypes.NONE for c in simple)
