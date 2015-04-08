package server

import "os"

//spy on file reads
type spyFile struct {
	*os.File
	read int64
}

func (s *spyFile) Read(p []byte) (int, error) {
	n, err := s.File.Read(p)
	s.read += int64(n)
	return n, err
}
