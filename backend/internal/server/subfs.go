package server

import (
	"errors"
	"io"
	"io/fs"
)

// subRoot wraps an fs.FS produced by fs.Sub, adding the ReadFile and
// ReadDir methods that fs.Sub strips off. embed.FS provides those
// natively on the parent, so the wrapper forwards back to the parent
// via Open() + reading the returned file.
//
// Production uses embed.FS directly (no wrapping needed). Tests use
// this wrapper to bridge fs.Sub into Options.
type subRoot struct {
	fsys fs.FS
}

func (s subRoot) ReadFile(name string) ([]byte, error) {
	f, err := s.fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readAll(f)
}

func (s subRoot) ReadDir(name string) ([]fs.DirEntry, error) {
	f, err := s.fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	df, ok := f.(fs.ReadDirFile)
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	return df.ReadDir(-1)
}

func (s subRoot) Open(name string) (fs.File, error) { return s.fsys.Open(name) }

// readAll reads the entire file into memory (small files only — used
// for templates, fonts, and CSS).
func readAll(f fs.File) ([]byte, error) {
	var buf []byte
	for {
		var chunk [4096]byte
		n, err := f.Read(chunk[:])
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if err != nil {
			// io.EOF is the normal end-of-file signal from Read;
			// returning whatever bytes we accumulated is the
			// expected behavior. Use errors.Is (not ==) so wrapped
			// EOF errors are still recognized.
			if errors.Is(err, io.EOF) {
				return buf, nil
			}
			return buf, err
		}
	}
}
