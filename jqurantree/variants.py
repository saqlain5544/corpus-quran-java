"""
Variants — 6 key Quran text variants: download, load, diff.
"""

from __future__ import annotations

import json, time, urllib.request, difflib, sys
from dataclasses import dataclass, field
from pathlib import Path

from .constants import VERSE_COUNTS, surah_name
from .alignment import Align, align
from .model import Document, Chapter, Verse, Token, Location
from .encoding import decode as _unicode_decode
from .text import ArabicText

_DATA = Path.home() / "corpus-quran-java" / "data" / "quran-text"
_TANZIL_URL = "https://tanzil.net/pub/download/index.php"
_GITHUB_BASE = "https://raw.githubusercontent.com/fawazahmed0/quran-api/refs/heads/1/editions"


@dataclass(slots=True, frozen=True)
class Variant:
    key: str
    name: str
    script: str
    qiraat: str
    filename: str
    source: str  # "tanzil" or "quran_api"
    tanzil_type: str = ""
    api_edition: str = ""


VARIANTS: dict[str, Variant] = {}
def _reg(k, n, s, q, f, src, tt="", ae=""):
    VARIANTS[k] = Variant(k, n, s, q, f, src, tt, ae)

_reg("uthmani", "Uthmani Hafs", "uthmani", "hafs", "quran-uthmani.xml",
     "tanzil", tt="uthmani")
_reg("simple", "Simple (Imla'ei)", "simple", "hafs", "quran-simple.xml",
     "tanzil", tt="simple")
_reg("simple-clean", "Simple Clean (bare consonants)", "simple", "hafs",
     "quran-simple-clean.xml", "tanzil", tt="simple-clean")
_reg("indopak", "Indo-Pak Script", "indopak", "hafs", "ara-quranindopak.json",
     "quran_api", ae="ara_quranindopak")
_reg("warsh", "Warsh 'an Nafi'", "uthmani", "warsh", "ara-quranwarsh.json",
     "quran_api", ae="ara_quranwarsh")
_reg("qaloon", "Qaloon 'an Nafi'", "uthmani", "qaloon", "ara-quranqaloon.json",
     "quran_api", ae="ara_quranqaloon")


@dataclass(slots=True)
class VerseDiff:
    chapter: int; verse: int; a_text: str; b_text: str
    identical: bool = True
    changes: list = field(default_factory=list)


@dataclass(slots=True)
class DiffResult:
    a: str; b: str; config: Align
    total: int = 0; identical: int = 0; token_changes: int = 0
    surahs: list = field(default_factory=list)

    def surah(self, n: int):
        return next((s for s in self.surahs if s["number"] == n), None)


def download_variant(key: str, force: bool = False) -> Path:
    v = VARIANTS.get(key)
    if v is None: raise KeyError(f"Unknown variant: {key}")
    _DATA.mkdir(parents=True, exist_ok=True)
    dest = _DATA / v.filename
    if not force and dest.exists() and dest.stat().st_size > 1000:
        return dest
    if v.source == "tanzil":
        return _dl_tanzil(v, dest)
    return _dl_api(v, dest)


def _dl_tanzil(v: Variant, dest: Path) -> Path:
    params = {"quranType": v.tanzil_type, "outType": "xml", "marks": "false",
              "sajdah": "false", "rub": "false", "tatweel": "true", "stanween": "false"}
    return _http_post(_TANZIL_URL, params, dest)


def _dl_api(v: Variant, dest: Path) -> Path:
    dir_name = v.api_edition.replace("_", "-")
    all_verses: list[str] = []
    for ch in range(1, 115):
        url = f"{_GITHUB_BASE}/{dir_name}/{ch}.json"
        for _ in range(3):
            try:
                req = urllib.request.Request(url, headers={"User-Agent": "jqurantree/1.0"})
                with urllib.request.urlopen(req, timeout=30) as resp:
                    if resp.status == 200:
                        data = json.loads(resp.read().decode())
                        ch_data = data.get("chapter", data)
                        if isinstance(ch_data, list):
                            all_verses.extend(item.get("text", "") for item in ch_data)
                        break
            except Exception:
                time.sleep(1)
    data_out = {str(i + 1): verse for i, verse in enumerate(all_verses)}
    with open(dest, "w", encoding="utf-8") as f:
        json.dump(data_out, f, ensure_ascii=False)
    return dest


def _http_post(url, params, dest) -> Path:
    for attempt in range(3):
        tmp = dest.with_suffix(".tmp")
        if tmp.exists(): tmp.unlink()
        try:
            data = urllib.parse.urlencode(params).encode()
            req = urllib.request.Request(url, data=data,
                                         headers={"User-Agent": "jqurantree/1.0",
                                                  "Content-Type": "application/x-www-form-urlencoded"})
            with urllib.request.urlopen(req, timeout=120) as resp:
                with open(tmp, "wb") as f:
                    while True:
                        chunk = resp.read(65536)
                        if not chunk: break
                        f.write(chunk)
            if tmp.stat().st_size > 1000:
                if dest.exists(): dest.unlink()
                tmp.rename(dest)
                return dest
        except Exception:
            if tmp.exists(): tmp.unlink()
        time.sleep(2 ** attempt)
    raise RuntimeError(f"Download failed: {dest}")


def load_variant(key: str) -> dict:
    v = VARIANTS.get(key)
    if v is None: raise KeyError(f"Unknown variant: {key}")
    path = download_variant(key)
    if v.source == "tanzil":
        from .tanzil import parse_xml
        doc = parse_xml(path)
        return _doc_to_dict(doc, key)
    else:
        with open(path, encoding="utf-8") as f:
            data = json.load(f)
        return _json_to_dict(data, key)


def _doc_to_dict(doc, key) -> dict:
    suras = []
    for ch in doc.chapters:
        ayas = []
        for verse in ch:
            text = " ".join(str(t) for t in verse)
            ayas.append({"number": verse.verse_number, "text": text})
        suras.append({"number": ch.chapter_number, "ayas": ayas})
    return {"variant_key": key, "suras": suras}


def _json_to_dict(data: dict, key) -> dict:
    suras = []
    vi = 1
    for cn in range(1, 115):
        vc = VERSE_COUNTS[cn - 1]
        ayas = []
        for an in range(1, vc + 1):
            text = data.get(str(vi + an - 1), "")
            ayas.append({"number": an, "text": text})
        vi += vc
        suras.append({"number": cn, "ayas": ayas})
    return {"variant_key": key, "suras": suras}


def diff_variants(key_a: str, key_b: str, cfg: Align | None = None) -> DiffResult:
    if cfg is None: cfg = Align.exact()
    data_a = load_variant(key_a)
    data_b = load_variant(key_b)
    sura_map_a = {s["number"]: s for s in data_a["suras"]}
    sura_map_b = {s["number"]: s for s in data_b["suras"]}

    result = DiffResult(a=key_a, b=key_b, config=cfg)
    for cn in range(1, 115):
        sa = sura_map_a[cn]
        sb = sura_map_b[cn]
        aa = {a["number"]: a["text"] for a in sa["ayas"]}
        ba = {a["number"]: a["text"] for a in sb["ayas"]}
        verses = []
        for an in sorted(set(aa.keys()) | set(ba.keys())):
            at = align(aa.get(an, ""), cfg)
            bt = align(ba.get(an, ""), cfg)
            result.total += 1
            if at == bt:
                result.identical += 1
                verses.append({"verse": an, "identical": True})
            else:
                matcher = difflib.SequenceMatcher(None, at.split(), bt.split())
                changes = list(matcher.get_opcodes())
                token_ch = sum(1 for t, *_ in changes if t != "equal")
                result.token_changes += token_ch
                verses.append({"verse": an, "identical": False, "changes": token_ch})
        result.surahs.append({
            "number": cn, "name": surah_name(cn),
            "verses": verses,
            "identical": sum(1 for v in verses if v.get("identical")),
            "different": sum(1 for v in verses if not v.get("identical")),
        })
    return result


def diff_variants_lemmas(key_a: str, key_b: str) -> dict:
    """Compare variants at the lemma level — unique word sets per verse.
    Returns: how many verses share the same word set (after diacritic removal + normalization)."""
    from .morphology import Morphology
    import unicodedata

    data_a = load_variant(key_a)
    data_b = load_variant(key_b)

    sura_map_a = {s["number"]: s for s in data_a["suras"]}
    sura_map_b = {s["number"]: s for s in data_b["suras"]}

    total_verses = 0
    same_words = 0
    diff_words = 0
    examples: list[dict] = []

    _diacritics = frozenset({
        0x064B, 0x064C, 0x064D, 0x064E, 0x064F, 0x0650, 0x0651, 0x0652,
        0x0670, 0x06DC, 0x06DF, 0x06E0, 0x06E1, 0x06E2, 0x06E3,
        0x06E5, 0x06E6, 0x06E8, 0x06EA, 0x06EB, 0x06EC, 0x06ED,
        0x0615, 0x06D6, 0x06D7, 0x06D8, 0x06D9, 0x06DA, 0x06DB,
    })

    def _normalize_word(w: str) -> str:
        w = "".join(ch for ch in w if ord(ch) not in _diacritics)
        w = unicodedata.normalize("NFC", w)
        w = w.replace("\u0671", "\u0627").replace("\u0670", "\u0627")
        w = w.replace("\u0649", "\u064A")
        return w

    for cn in range(1, 115):
        sa = sura_map_a[cn]
        sb = sura_map_b[cn]
        for an in range(1, max(len(sa["ayas"]), len(sb["ayas"])) + 1):
            total_verses += 1

            a_text = sa["ayas"][an - 1]["text"] if an <= len(sa["ayas"]) else ""
            b_text = sb["ayas"][an - 1]["text"] if an <= len(sb["ayas"]) else ""

            a_words = set(_normalize_word(w) for w in a_text.split() if len(_normalize_word(w)) > 1)
            b_words = set(_normalize_word(w) for w in b_text.split() if len(_normalize_word(w)) > 1)

            only_a = a_words - b_words
            only_b = b_words - a_words

            if not only_a and not only_b:
                same_words += 1
            else:
                diff_words += 1
                if len(examples) < 5 and (only_a or only_b):
                    examples.append({
                        "chapter": cn, "verse": an,
                        "only_in_a": sorted(only_a)[:5],
                        "only_in_b": sorted(only_b)[:5],
                        "a_text": a_text[:100],
                        "b_text": b_text[:100],
                    })

    return {
        "variant_a": key_a, "variant_b": key_b,
        "level": "word-set (diacritic-stripped + NFC + equivalence)",
        "total_verses": total_verses,
        "same_words": same_words,
        "different_words": diff_words,
        "word_similarity_pct": round(same_words / total_verses * 100, 1) if total_verses else 0,
        "examples": examples,
    }
