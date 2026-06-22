-- CreateTable
CREATE TABLE "surah" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL,
    "englishName" TEXT NOT NULL,
    "englishTranslation" TEXT NOT NULL,
    "revelationType" TEXT NOT NULL,
    "verseCount" INTEGER NOT NULL
);

-- CreateTable
CREATE TABLE "verse" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "surahId" INTEGER NOT NULL,
    "number" INTEGER NOT NULL,
    "text" TEXT NOT NULL,
    CONSTRAINT "verse_surahId_fkey" FOREIGN KEY ("surahId") REFERENCES "surah" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "word" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "surahId" INTEGER NOT NULL,
    "verseId" INTEGER NOT NULL,
    "number" INTEGER NOT NULL,
    "token" TEXT NOT NULL,
    "withoutDiacritics" TEXT NOT NULL,
    "translation" TEXT,
    "punctuationMark" TEXT,
    CONSTRAINT "word_surahId_fkey" FOREIGN KEY ("surahId") REFERENCES "surah" ("id") ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT "word_verseId_fkey" FOREIGN KEY ("verseId") REFERENCES "verse" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "segment" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "wordId" INTEGER NOT NULL,
    "number" INTEGER NOT NULL,
    "text" TEXT NOT NULL,
    "partOfSpeech" TEXT NOT NULL,
    "morphType" TEXT NOT NULL,
    "lemma" TEXT,
    "root" TEXT,
    "gender" TEXT,
    "case_field" TEXT,
    "syntacticRole" TEXT,
    "gloss" TEXT,
    "possessiveConstruct" TEXT,
    "caseMoodMarker" TEXT,
    "invariableDeclinable" TEXT,
    "phrase" TEXT,
    "phrasalFunction" TEXT,
    "punctuationMark" TEXT,
    CONSTRAINT "segment_wordId_fkey" FOREIGN KEY ("wordId") REFERENCES "word" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "root_reference" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "root" TEXT NOT NULL,
    "surahId" INTEGER NOT NULL,
    "verseNumber" INTEGER NOT NULL,
    "wordNumber" INTEGER NOT NULL,
    CONSTRAINT "root_reference_surahId_fkey" FOREIGN KEY ("surahId") REFERENCES "surah" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "root_meaning" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "root" TEXT NOT NULL,
    "rootArabic" TEXT NOT NULL,
    "generalMeaning" TEXT NOT NULL,
    "shadesOfMeaning" TEXT NOT NULL,
    "hadithExamples" TEXT NOT NULL,
    "classicalSources" TEXT NOT NULL,
    "idiomsCustoms" TEXT NOT NULL,
    "wordByWordShades" TEXT NOT NULL,
    "pos" TEXT,
    "modelUsed" TEXT
);

-- CreateIndex
CREATE INDEX "verse_surahId_idx" ON "verse"("surahId");

-- CreateIndex
CREATE UNIQUE INDEX "verse_surahId_number_key" ON "verse"("surahId", "number");

-- CreateIndex
CREATE INDEX "word_surahId_idx" ON "word"("surahId");

-- CreateIndex
CREATE INDEX "word_verseId_idx" ON "word"("verseId");

-- CreateIndex
CREATE UNIQUE INDEX "word_verseId_number_key" ON "word"("verseId", "number");

-- CreateIndex
CREATE INDEX "segment_wordId_idx" ON "segment"("wordId");

-- CreateIndex
CREATE INDEX "segment_root_idx" ON "segment"("root");

-- CreateIndex
CREATE UNIQUE INDEX "segment_wordId_number_key" ON "segment"("wordId", "number");

-- CreateIndex
CREATE INDEX "root_reference_root_idx" ON "root_reference"("root");

-- CreateIndex
CREATE UNIQUE INDEX "root_meaning_root_key" ON "root_meaning"("root");

-- CreateIndex
CREATE INDEX "root_meaning_root_idx" ON "root_meaning"("root");
