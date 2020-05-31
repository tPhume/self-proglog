package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())
	return &store{
		File: f,
		buf:  bufio.NewWriter(f),
		size: size,
	}, nil
}

// Add contents to file then return number of bytes written and position (offset)
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Gets size of content to be written then write size to file
	pos = s.size
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// Write the contents to file
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	// Add 8 bytes (length of uint64 number) to the length of contents written
	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

// Read content given the position (offset) then return the content as slice of bytes
func (s *store) ReadAt(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write any unwritten contents in buffer first
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	// Get size of content by getting the first 8 bytes starting from the offset
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	// Make a variable to store the contents and read the contents only
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return b, nil
}
