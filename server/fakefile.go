package server

import "os"

//spy on file reads
type fakeFile struct {
	*os.File
	read int64
}

func (s *fakeFile) Read(p []byte) (int, error) {
	n, err := s.File.Read(p)
	s.read += int64(n)
	return n, err
}
