package server

import (
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jpillora/ansi"
)

var scale = []string{"b", "kb", "mb", "gb"}

func tobyte(n int64) string {
	for _, s := range scale {
		if n > 1024 {
			n = n / 1024
		} else {
			return strconv.FormatInt(n, 10) + s
		}
	}
	return strconv.FormatInt(n, 10)
}

func shorten(s string) string {
	usr, err := user.Current()
	if err != nil {
		return s
	}
	if strings.HasPrefix(s, usr.HomeDir) {
		s = strings.Replace(s, usr.HomeDir, "~", 1)
	}
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
