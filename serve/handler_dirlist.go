package serve

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jpillora/serve/serve/static"

	"github.com/jpillora/sizestr"
)

var dirlistHtmlTempl *template.Template

func init() {
	t := template.New("dirlist")
	t = t.Funcs(template.FuncMap{
		"tosize": sizestr.ToString,
		"split":  strings.Split,
		"concat": func(a, b string) string { return a + b },
	})
	listHTML := static.MustAsset("static/list.html")
	var err error
	dirlistHtmlTempl, err = t.Parse(string(listHTML))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type listDir struct {
	Path, Parent      string
	NumFiles, NumDirs int
	TotalSize         int64
	Archive           bool
	Files             []listFile
}

type listFile struct {
	Path, Name string
	Accessible bool
	IsDir      bool
	Size       int64
	Mtime      time.Time
}

type byName []listFile

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (s *Handler) dirlist(w http.ResponseWriter, r *http.Request, dir string) {

	path, _ := filepath.Rel(s.c.Directory, dir)
	parent := ""
	if path != "." {
		parent = "/" + filepath.Join(path, "..")
	}

	list := &listDir{
		Path:    path,
		Parent:  parent,
		Archive: !s.c.NoArchive,
		Files:   []listFile{},
	}

	//readnames and stat separately so a single failed
	//stat doesn't cause the directory listing to fail
	d, err := os.Open(dir)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Cannot open directory: %s", err)
		return
	}
	names, err := d.Readdirnames(-1)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Cannot list directory: %s", err)
		return
	}

	for _, n := range names {
		if n == ".DS_Store" {
			continue //Nope.
		}
		lf := listFile{
			Name: n,
			Path: "/" + filepath.Join(path, n),
		}
		//attempt to stat
		if f, err := os.Stat(filepath.Join(dir, n)); err == nil {
			lf.Accessible = true
			var size int64
			if f.IsDir() {
				n += "/"
				list.NumDirs++
			} else {
				list.NumFiles++
				size = f.Size()
				list.TotalSize += size
			}
			lf.IsDir = f.IsDir()
			lf.Size = size
			lf.Mtime = f.ModTime()
		}

		list.Files = append(list.Files, lf)
	}

	sort.Sort(byName(list.Files))

	accepts := strings.Split(r.Header.Get("Accept"), ",")
	buff := &bytes.Buffer{}
	contype := ""
	for _, accept := range accepts {
		typeencoding := strings.SplitN(accept, "/", 2)
		if len(typeencoding) != 2 {
			continue
		}
		switch typeencoding[1] {
		case "json":
			b, _ := json.MarshalIndent(list, "", "  ")
			buff.Write(b)
		case "xml":
			b, _ := xml.MarshalIndent(list, "", "  ")
			buff.Write(b)
		case "html":
			dirlistHtmlTempl.Execute(buff, list)
		default:
			continue
		}
		contype = accept
		break
	}

	if contype == "" {
		for _, f := range list.Files {
			buff.WriteString(f.Name + "\n")
		}
		contype = "text/plain"
	}

	w.Header().Set("Content-Type", contype)
	w.WriteHeader(200)
	w.Write(buff.Bytes())
}
