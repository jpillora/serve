package serve

import (
	"os/user"
	"path/filepath"
	"strings"
)

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
