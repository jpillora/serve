package server

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/jpillora/sizestr"
)

func init() {
	sizestr.ToggleCase()

	t := template.New("name")

	t = t.Funcs(template.FuncMap{
		"tosize": sizestr.ToString,
		"split":  strings.Split,
		"concat": func(a, b string) string { return a + b },
	})

	var err error
	dirlistHtmlTempl, err = t.Parse(dirlistHtml)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
