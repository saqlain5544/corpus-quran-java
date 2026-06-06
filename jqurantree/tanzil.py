"""
Tanzil — download and parse Quran XML/text from the Tanzil project.
"""

from __future__ import annotations

import time, urllib.parse, urllib.request
from pathlib import Path
from xml.etree.cElementTree import iterparse

from .encoding import decode as _unicode_decode
from .text import ArabicTextBuilder, ArabicText
from .model import Document, Chapter, Verse, Token, Location

_TANZIL_URL = "https://tanzil.net/pub/download/index.php"
_CACHE_DIR = Path.home() / ".cache" / "quran"
_BUFFER_SIZE = 65536
_MAX_RETRIES = 3


class DownloadError(Exception):
    pass


def download_tanzil(xml_path: Path | None = None, force: bool = False) -> Path:
    _CACHE_DIR.mkdir(parents=True, exist_ok=True)
    dest = xml_path or (_CACHE_DIR / "quran-uthmani.xml")
    if not force and dest.exists() and dest.stat().st_size > 500000:
        return dest

    params = {"quranType": "uthmani", "outType": "xml", "marks": "false",
              "sajdah": "false", "rub": "false", "tatweel": "true", "stanween": "false"}
    for attempt in range(_MAX_RETRIES):
        tmp = dest.with_suffix(".tmp")
        if tmp.exists(): tmp.unlink()
        try:
            data = urllib.parse.urlencode(params).encode()
            req = urllib.request.Request(_TANZIL_URL, data=data,
                                         headers={"User-Agent": "jqurantree/1.0",
                                                  "Content-Type": "application/x-www-form-urlencoded"})
            with urllib.request.urlopen(req, timeout=120) as resp:
                if resp.status < 200 or resp.status >= 300:
                    raise DownloadError(f"HTTP {resp.status}")
                with open(tmp, "wb") as f:
                    while True:
                        chunk = resp.read(_BUFFER_SIZE)
                        if not chunk: break
                        f.write(chunk)
            if tmp.stat().st_size > 500000:
                if dest.exists(): dest.unlink()
                tmp.rename(dest)
                return dest
        except Exception:
            if tmp.exists(): tmp.unlink()
        time.sleep(2 ** attempt)
    raise DownloadError("Download failed")


def _local(tag: str) -> str:
    return tag.rsplit("}", 1)[-1] if "}" in tag else tag


def parse_xml(file_path: Path) -> Document:
    Document.reset()
    doc = Document.get()
    ch_map: dict[int, Chapter] = {}
    cur_sura: int | None = None
    cur_name = ""
    cur_bism: str | None = None
    cur_verses: list[Verse] = []
    prev_sura: int | None = None

    for event, elem in iterparse(str(file_path), events=("start", "end")):
        tag = _local(elem.tag).lower()
        if event == "start" and tag == "sura":
            idx = elem.get("index")
            if idx is not None:
                cur_sura = int(idx)
                cur_name = elem.get("name", "")
                if prev_sura is not None and cur_verses and prev_sura != cur_sura:
                    nm = _mk_arabic(cur_name) if cur_name else ArabicText.from_unicode("")
                    bm = _mk_arabic(cur_bism) if cur_bism else None
                    ch_map[prev_sura] = Chapter(prev_sura, nm, bm, cur_verses)
                    cur_verses = []; cur_bism = None
                prev_sura = cur_sura

        elif event == "end" and tag == "aya":
            idx = elem.get("index")
            aya_text = elem.get("text", "")
            bism_text = elem.get("bismillah", "")
            if idx is not None and cur_sura is not None:
                an = int(idx)
                if bism_text and not cur_bism:
                    cur_bism = bism_text
                full = aya_text.strip()
                if bism_text and an == 1 and cur_sura != 1:
                    full = bism_text + " " + full if full else bism_text
                tokens: list[Token] = []
                if full:
                    for ti, tt in enumerate(full.split(), start=1):
                        tokens.append(Token.from_characters(
                            _unicode_decode(tt),
                            Location(cur_sura, an, ti)))
                cur_verses.append(Verse(tokens, Location(cur_sura, an, 0)))
            elem.clear()
        elif event == "end" and tag == "sura":
            elem.clear()

    if prev_sura is not None and cur_verses:
        nm = _mk_arabic(cur_name) if cur_name else ArabicText.from_unicode("")
        bm = _mk_arabic(cur_bism) if cur_bism else None
        ch_map[prev_sura] = Chapter(prev_sura, nm, bm, cur_verses)

    chapters = [ch_map.get(cn, Chapter(cn, ArabicText.from_unicode(""),
                None, [])) for cn in range(1, 115)]
    doc.set_chapters(chapters)
    return doc


def _mk_arabic(text: str) -> ArabicText:
    builder = ArabicTextBuilder()
    for ac in _unicode_decode(text):
        builder.add(ac)
    return builder.to_arabic_text()
