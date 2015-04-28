package server

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type listDir struct {
	Path, Parent string
	Files        []listFile
}

type listFile struct {
	Path, Name string
	IsDir      bool
	Size       int64
	Mtime      time.Time
}

func (s *Server) dirlist(w http.ResponseWriter, r *http.Request, dir string) {

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	path, _ := filepath.Rel(s.c.Directory, dir)
	parent := ""
	if path != "." {
		parent = "/" + filepath.Join(path, "..")
	}

	list := &listDir{
		Path:   path,
		Parent: parent,
		Files:  make([]listFile, len(infos)),
	}

	for i, f := range infos {
		n := f.Name()
		if f.IsDir() {
			n += "/"
		}
		list.Files[i] = listFile{
			Name:  n,
			Path:  "/" + filepath.Join(path, n),
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

var dirlistHtmlTempl *template.Template

var dirlistHtml = `
<html>
	<head>
		<title>{{ .Path }}</title>
		<style>
			html,body {
				height:100%;
				width:100%;
				font-family: Courier, monospace;
			}
			a {
				text-decoration: none;
			}
			table {
				width: 300px;
				margin: 5%;
			}
			.path {
				text-style: underline;
			}
			.name {
				text-align: right;
				padding-right: 30px;
			}
			.size {
				text-align: left;
			}
		</style>
	</head>
	<body>
		<table>
			<tr>
				<th class="path" colspan="2">
					<span class="pathstr">
						<a href="/">/</a>
						{{ $path := "" }}
						{{range $i, $p := split .Path "/"}}
							{{ $path := concat $path $p }}
							<a href="/{{ $path }}/">{{ $p }}/</a>
						{{end}}
					</span>
				</th>
			</tr>
			<tr>
				<th class="name">Name</th>
				<th class="size">Size</th>
			</tr>
			{{if ne .Parent ""}}<tr class="file item">
				<td class="name"><a href="{{ .Parent }}">..</a></td>
				<td class="size">-</td>
			</tr>{{end}}
			{{range .Files}}<tr class="file item">
				<td class="name">
					<a href="{{ .Path }}{{if .IsDir}}/{{end}}">{{ .Name }}</a>
				</td>
				<td class="size" alt="{{ .Size }} bytes">
					{{if .IsDir}}-{{else}}{{ tosize .Size }}{{end}}
				</td>
			</tr>{{end}}
			<tr>
				<th class="stats" colspan="2">
					{{ $numfiles := len .Files }}
					{{ $numfiles }} file{{if ne $numfiles 1 }}s{{end}}
				</th>
			</tr>
		</table>
	</body>
</html>
`
