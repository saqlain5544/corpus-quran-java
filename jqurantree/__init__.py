"""
jqurantree — Python Quran NLP library.

Core: constants, types, text, encoding, alignment, model, tanzil, search, variants.
~1,900 LOC total. Single-library, zero-ceremony.

Usage:
  from jqurantree import parse_xml, Document, TokenSearch, AnalysisTable, Align, diff_variants

  doc = parse_xml(xml_path)
  ts = TokenSearch()
  ts.find_token("الله").get_results()
  diff = diff_variants("uthmani", "warsh", Align.canonical())
"""

from .constants import Surah, VERSE_COUNTS, surah_name, verse_count
from .types import CharacterType, DiacriticTypes, ArabicCharacter
from .text import ArabicText, ArabicTextBuilder, ByteFormat
from .encoding import decode, encode, encode_buckwalter, decode_buckwalter, encode_simple
from .alignment import Align, align
from .model import Document, Chapter, Verse, Token, Location
from .tanzil import parse_xml, download_tanzil, DownloadError
from .search import TokenSearch, AnalysisTable
from .variants import VARIANTS, Variant, download_variant, load_variant, diff_variants, diff_variants_lemmas, DiffResult, VerseDiff
from .morphology import Morphology, Segment
from .errors import JQuranTreeError, ErrorCode

__version__ = "4.0.0"

__all__ = [
    "Surah", "VERSE_COUNTS", "surah_name", "verse_count",
    "CharacterType", "DiacriticTypes", "ArabicCharacter",
    "ArabicText", "ArabicTextBuilder", "ByteFormat",
    "decode", "encode", "encode_buckwalter", "decode_buckwalter", "encode_simple",
    "Align", "align",
    "Document", "Chapter", "Verse", "Token", "Location",
    "parse_xml", "download_tanzil", "DownloadError",
    "TokenSearch", "AnalysisTable",
    "VARIANTS", "Variant", "download_variant", "load_variant", "diff_variants",
    "diff_variants_lemmas", "DiffResult", "VerseDiff",
    "Morphology", "Segment",
    "JQuranTreeError", "ErrorCode",
]
