package server

import "net/http"

//inpsectable ResponseWriter
type ServeWriter struct {
	w http.ResponseWriter
	//stats
	Code int
	Size int64
}

func (s *ServeWriter) Header() http.Header {
	return s.w.Header()
}

func (s *ServeWriter) Write(p []byte) (int, error) {
	s.Size += int64(len(p))
	return s.w.Write(p)
}

func (s *ServeWriter) WriteHeader(c int) {
	s.Code = c
	s.w.WriteHeader(c)
}
