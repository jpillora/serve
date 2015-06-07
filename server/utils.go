package server

import (
	"net/http"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jpillora/ansi"
)

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

//util functions
func ShortenPath(s string) string {
	usr, err := user.Current()
	if err != nil {
		return s
	}
	if strings.HasPrefix(s, usr.HomeDir) {
		s = strings.Replace(s, usr.HomeDir, "~", 1)
	}
	s = strings.TrimSuffix(s, string(filepath.Separator))
	return s
}

type colors struct {
	Grey, Cyan, Yellow, Red, Reset string
}

var defaultColors = &colors{
	string(ansi.Set(ansi.Black)), string(ansi.Set(ansi.Cyan)), string(ansi.Set(ansi.Yellow)), string(ansi.Set(ansi.Yellow)), string(ansi.Set(ansi.Reset)),
}

func colorcode(status int) string {
	switch status / 100 {
	case 2:
		return string(ansi.Set(ansi.Green))
	case 3:
		return string(ansi.Set(ansi.Cyan))
	case 4:
		return string(ansi.Set(ansi.Yellow))
	case 5:
		return string(ansi.Set(ansi.Red))
	}
	return string(ansi.Set(ansi.Black))
}

var fmtdurationRe = regexp.MustCompile(`\.\d+`)

func fmtduration(t time.Duration) string {
	return fmtdurationRe.ReplaceAllString(t.String(), "")
}
