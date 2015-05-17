package server

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/jpillora/sizestr"
)

var introFormat = `{{ .Grey }}serving {{ .Cyan }}{{ .Dir }} {{ .Grey }}on port {{ .Cyan }}{{ .Port }}{{ .Reset }}` + "\n"
var logFormat = `{{ .Grey }}{{ if .Timestamp }}{{ .Timestamp }} {{end}}` +
	`{{ .Method }} {{ .Path }} {{ .CodeColor }}{{ .Code }}{{ .Grey }} ` +
	`{{ .Duration }} {{ .Size }}{{ if .IP }} ({{ .IP }}){{end}}{{ .Reset }}` + "\n"

var introTemplate, logTemplate *template.Template

func init() {
	sizestr.ToggleCase()

	var err error
	t := template.New("intro")
	introTemplate, err = t.Parse(introFormat)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t = template.New("log")
	logTemplate, err = t.Parse(logFormat)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t = template.New("dirlist")
	t = t.Funcs(template.FuncMap{
		"tosize": sizestr.ToString,
		"split":  strings.Split,
		"concat": func(a, b string) string { return a + b },
	})
	dirlistHtmlTempl, err = t.Parse(dirlistHtml)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
