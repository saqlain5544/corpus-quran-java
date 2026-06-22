"use client";

import { useState, useEffect, useCallback } from "react";

export interface Bookmark {
  surahNumber: number;
  verseNumber: number;
  surahName: string;
  timestamp: number;
  note?: string;
}

const STORAGE_KEY = "quran-bookmarks";

function loadBookmarks(): Bookmark[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveBookmarks(bookmarks: Bookmark[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(bookmarks));
}

export function useBookmarks() {
  const [bookmarks, setBookmarks] = useState<Bookmark[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    setBookmarks(loadBookmarks());
    setLoaded(true);
  }, []);

  const addBookmark = useCallback((bookmark: Bookmark) => {
    setBookmarks((prev) => {
      const exists = prev.find(
        (b) => b.surahNumber === bookmark.surahNumber && b.verseNumber === bookmark.verseNumber
      );
      if (exists) return prev;
      const updated = [bookmark, ...prev];
      saveBookmarks(updated);
      return updated;
    });
  }, []);

  const removeBookmark = useCallback((surahNumber: number, verseNumber: number) => {
    setBookmarks((prev) => {
      const updated = prev.filter(
        (b) => !(b.surahNumber === surahNumber && b.verseNumber === verseNumber)
      );
      saveBookmarks(updated);
      return updated;
    });
  }, []);

  const toggleBookmark = useCallback(
    (bookmark: Bookmark) => {
      const exists = bookmarks.find(
        (b) => b.surahNumber === bookmark.surahNumber && b.verseNumber === bookmark.verseNumber
      );
      if (exists) {
        removeBookmark(bookmark.surahNumber, bookmark.verseNumber);
      } else {
        addBookmark(bookmark);
      }
    },
    [bookmarks, addBookmark, removeBookmark]
  );

  const isBookmarked = useCallback(
    (surahNumber: number, verseNumber: number) => {
      return bookmarks.some(
        (b) => b.surahNumber === surahNumber && b.verseNumber === verseNumber
      );
    },
    [bookmarks]
  );

  return { bookmarks, loaded, addBookmark, removeBookmark, toggleBookmark, isBookmarked };
}
