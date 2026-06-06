"""
Errors — minimal error types.
"""

from __future__ import annotations

from enum import IntEnum


class ErrorCode(IntEnum):
    NONE = 0
    INVALID_CHAPTER = 1
    INVALID_VERSE = 2
    DOCUMENT_UNINITIALIZED = 3
    DOWNLOAD_FAILED = 4
    PARSE_ERROR = 5


class JQuranTreeError(Exception):
    def __init__(self, message: str, code: ErrorCode = ErrorCode.NONE):
        super().__init__(message)
        self.code = code
