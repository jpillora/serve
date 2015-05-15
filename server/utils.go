package server

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jpillora/ansi"
	"github.com/jpillora/archiver"
)

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

func c(s string, c string) string {
	var color ansi.Attribute
	switch c {
	case "grey":
		color = ansi.Black
	case "cyan":
		color = ansi.Cyan
	case "yellow":
		color = ansi.Yellow
	case "red":
		color = ansi.Red
	default:
		color = ansi.Green
	}
	return string(ansi.Set(color)) + s + string(ansi.Set(ansi.Reset))
}

func fmtcode(status int) string {
	s := strconv.Itoa(status)
	switch status / 100 {
	case 2:
		return c(s, "green")
	case 3:
		return c(s, "cyan")
	case 4:
		return c(s, "yellow")
	case 5:
		return c(s, "red")
	}
	return s
}

var fmtdurationRe = regexp.MustCompile(`\.\d+`)

func fmtduration(t time.Duration) string {
	return fmtdurationRe.ReplaceAllString(t.String(), "")
}

func archiveRequest(path string) (dir, ext string, ok bool) {
	ext = archiver.Extension(path)
	if ext == "" {
		return
	}
	dir = strings.TrimSuffix(path, ext)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return
	}
	ok = true
	return
}
