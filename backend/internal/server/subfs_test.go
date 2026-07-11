package server

// subFS is test-only: it bridges fs.Sub (which strips ReadFile/ReadDir)
// back into a value that satisfies fs.ReadFileFS so Options.FS can be
// passed fs.Sub output. embed.FS in production already satisfies
// fs.ReadFileFS natively, so production code never imports this.

import (
	"errors"
	"io"
	"io/fs"
)

// subRootFS is the test-only wrapper around an fs.Sub output.
type subRootFS struct {
	fsys fs.FS
}

func (s subRootFS) Open(name string) (fs.File, error) { return s.fsys.Open(name) }

func (s subRootFS) ReadFile(name string) ([]byte, error) {
	f, err := s.fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readAllFS(f)
}

func (s subRootFS) ReadDir(name string) ([]fs.DirEntry, error) {
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

// readAllFS reads the entire file into memory (test-only — production
// assets are embedded and don't need this path).
func readAllFS(f fs.File) ([]byte, error) {
	var buf []byte
	for {
		var chunk [4096]byte
		n, err := f.Read(chunk[:])
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return buf, nil
			}
			return buf, err
		}
	}
}
