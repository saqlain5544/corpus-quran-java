import json
import re
from html import escape
from pathlib import Path

from masaq import Masaq
from quran import parse_quran

QURAN_PATH = "../data/quran/quran-uthmani.xml"
MASAQ_PATH = "../data/morphology/MASAQ.csv"

OUTPUT_DIR = Path("../surahs-html")


quran = parse_quran(QURAN_PATH)
masaq = Masaq(MASAQ_PATH)


# ==========================================================
# Quran marks
# ==========================================================

QURAN_MARKS = {
    "\u06d6": {
        "name": "قلى",
        "type": "waqf",
        "meaning": "Stop preferred",
        "reader_action": "Stopping is better, continuing is allowed.",
    },
    "\u06d7": {
        "name": "صلى",
        "type": "waqf",
        "meaning": "Continue preferred",
        "reader_action": "Continuing is better.",
    },
    "\u06d8": {
        "name": "م",
        "type": "waqf",
        "meaning": "Mandatory stop",
        "reader_action": "Stop here.",
    },
    "\u06d9": {
        "name": "لا",
        "type": "waqf",
        "meaning": "Do not stop",
        "reader_action": "Avoid stopping.",
    },
    "\u06da": {
        "name": "ج",
        "type": "waqf",
        "meaning": "Permissible stop",
        "reader_action": "Stop or continue.",
    },
    "\u06db": {
        "name": "وقف معانقة",
        "type": "waqf",
        "meaning": "Paired stopping sign",
        "reader_action": "Do not stop at both places.",
    },
    "\u06dd": {
        "name": "۝",
        "type": "ayah-marker",
        "meaning": "Ayah boundary",
        "reader_action": "Verse ending.",
    },
    "\u06de": {
        "name": "۞",
        "type": "division-marker",
        "meaning": "Rub al-Hizb",
        "reader_action": "Section marker.",
    },
    "\u06e9": {
        "name": "۩",
        "type": "sajdah",
        "meaning": "Sajdah marker",
        "reader_action": "Perform sajdah tilawah when applicable.",
    },
}


TRAILING_MARKS = set(QURAN_MARKS)


# ==========================================================
# Tokenizer
# ==========================================================


def split_mark(token):

    marks = []

    while token and token[-1] in TRAILING_MARKS:
        marks.append(token[-1])
        token = token[:-1]

    return token, marks[::-1]


def tokenize(text):

    result = []

    for token in text.split():
        word, marks = split_mark(token)

        if word:
            result.append(("word", word))

        for mark in marks:
            result.append(("mark", mark))

    return result


# ==========================================================
# HTML helpers
# ==========================================================


def slugify(text):

    text = re.sub(r"[^\w]+", "-", text)

    return text.strip("-")


def render_mark(symbol):

    info = QURAN_MARKS[symbol]

    description = f"{info['name']}: {info['meaning']}. {info['reader_action']}"

    return f"""
<span
    class="quran-mark {info["type"]}"
    role="note"
    aria-label="{escape(description)}"
    data-symbol="{escape(symbol)}"
    data-meaning="{escape(info["meaning"])}"
    data-action="{escape(info["reader_action"])}"
>
    {symbol}
</span>
"""


def render_word(word, surah, ayah, word_no, index):

    morph = masaq.word(surah, ayah, word_no)

    return f"""
<span
    class="quran-word"

    data-surah="{surah}"
    data-ayah="{ayah}"
    data-word="{word_no}"
    data-index="{index}"

    data-morphology='{escape(json.dumps(morph, ensure_ascii=False))}'
>
    {escape(word)}
</span>
"""


# ==========================================================
# Surah HTML
# ==========================================================


def render_surah(number):

    surah = quran.surah(number)

    html = []

    html.append(
        """<!doctype html>
<html lang="ar" dir="rtl">

<head>
<meta charset="utf-8">

<title>
"""
        + escape(surah.name)
        + """
</title>

</head>

<body>

<main
    class="quran"
    data-surah-number="
"""
        + str(number)
        + """
">

<header class="surah-header">

<h1 class="surah-name">
"""
        + escape(surah.name)
        + """
</h1>

</header>
"""
    )

    global_index = 1

    for verse in surah:
        html.append(
            f"""
<section
    class="ayah"
    id="ayah-{verse.number}"
    data-ayah="{verse.number}"
>


<h2 class="ayah-number">
{verse.number}
</h2>


<p class="ayah-text">
"""
        )

        word_no = 1

        for kind, token in tokenize(verse.text):
            if kind == "mark":
                html.append(render_mark(token))

                continue

            html.append(render_word(token, number, verse.number, word_no, global_index))

            word_no += 1
            global_index += 1

        html.append(
            """
</p>

</section>
"""
        )

    html.append(
        """
</main>

</body>

</html>
"""
    )

    return "\n".join(html)


# ==========================================================
# Generate all surahs
# ==========================================================


def generate():

    OUTPUT_DIR.mkdir(exist_ok=True)

    for number, surah in quran.surahs.items():
        filename = f"{number:03d}-{slugify(surah.name)}.html"

        path = OUTPUT_DIR / filename

        path.write_text(render_surah(number), encoding="utf-8")

        print("generated", path)


if __name__ == "__main__":
    generate()
