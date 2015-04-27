package server

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type listDir struct {
	Path  string
	Files []listFile
}

type listFile struct {
	Name  string
	IsDir bool
	Size  int64
	Mtime time.Time
}

func (s *Server) dirlist(w http.ResponseWriter, r *http.Request, dir string) {

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	list := &listDir{
		Path:  dir,
		Files: make([]listFile, len(infos)),
	}

	for i, f := range infos {
		n := f.Name()
		if f.IsDir() {
			n += "/"
		}
		list.Files[i] = listFile{
			Name:  n,
			IsDir: f.IsDir(),
			Size:  f.Size(),
			Mtime: f.ModTime(),
		}
	}

	accepts := strings.Split(r.Header.Get("Accept"), ",")
	buff := &bytes.Buffer{}
	contype := ""
	for _, accept := range accepts {
		tenc := strings.SplitN(accept, "/", 2)
		if len(tenc) != 2 {
			continue
		}
		switch tenc[1] {
		case "json":
			b, _ := json.Marshal(list)
			buff.Write(b)
		case "xml":
			b, _ := xml.Marshal(list)
			buff.Write(b)
		case "html":
			for _, f := range list.Files {

				s := fmt.Sprintf("<a href=\"%s\">\n\t%s\n</a><br>\n", f.Name, f.Name)
				buff.WriteString(s)
			}
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
