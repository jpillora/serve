package server

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jaschaephraim/lrserver"
	"github.com/jpillora/sizestr"
	"gopkg.in/fsnotify.v1"
)

func init() {
	sizestr.ToggleCase()
}

//Server is custom file server
type Server struct {
	c        *Config
	addr     string
	port     string
	root     string
	hasIndex bool
	fallback *httputil.ReverseProxy
	watcher  *fsnotify.Watcher
	watching map[string]bool
	lr       *lrserver.Server
}

//NewServer creates a new Server
func New(c *Config) (*Server, error) {

	port := strconv.Itoa(c.Port)
	s := &Server{
		c:    c,
		port: port,
		addr: c.Host + ":" + port,
	}

	_, err := os.Stat(c.Dir)
	if c.Dir == "" || err != nil {
		return nil, fmt.Errorf("Missing directory: %s", c.Dir)
	}

	if c.PushState {
		s.root = filepath.Join(c.Dir, "index.html")
		if _, err := os.Stat(s.root); err != nil {
			return nil, fmt.Errorf("'%s' is required for pushstate", s.root)

		}
		s.hasIndex = true
	}

	if c.Fallback != "" {
		u, err := url.Parse(c.Fallback)

		if !strings.HasPrefix(u.Scheme, "http") {
			return nil, fmt.Errorf("Invalid fallback protocol scheme")
		}

		if err != nil {
			return nil, err
		}
		s.fallback = httputil.NewSingleHostReverseProxy(u)
	}

	if c.LiveReload {
		s.lr, _ = lrserver.New("serve-lr", lrserver.DefaultPort)
		discard := log.New(ioutil.Discard, "", 0)
		s.lr.SetErrorLog(discard)
		s.lr.SetStatusLog(discard)
		s.watching = map[string]bool{}
		s.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) Start() error {

	if s.c.Open {
		go func() {
			time.Sleep(500 * time.Millisecond)
			cmd := exec.Command("open", "http://localhost:"+s.port)
			cmd.Run()
		}()
	}

	if s.c.LiveReload {
		go func() {
			if err := s.lr.ListenAndServe(); err != nil {
				fmt.Printf("LiveReload server closed: %s", err)
			}
		}()
		go func() {
			for {
				event := <-s.watcher.Events
				s.lr.Reload(event.Name)
			}
		}()
	}

	h := http.Handler(http.HandlerFunc(s.handle))
	//logging is enabled
	if !s.c.NoLogging {
		fmt.Println(c("serving ", "grey") +
			c(ShortenPath(s.c.Dir), "cyan") +
			c(" on port ", "grey") +
			c(s.port, "cyan"))
		//insert measurement middleware
		h = s.measure(h)
	}
	//listen
	return http.ListenAndServe(s.addr, h)
}

func (s *Server) measure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//capture response
		dummyw := httptest.NewRecorder()
		//track timing
		t0 := time.Now()
		//perform response
		next.ServeHTTP(dummyw, r)
		//log result
		t := time.Now().Sub(t0)

		//write real response
		for name, _ := range dummyw.HeaderMap {
			w.Header().Set(name, dummyw.Header().Get(name))
		}
		w.WriteHeader(dummyw.Code)
		w.Write(dummyw.Body.Bytes())

		//log result
		fmt.Println(c(r.Method+" "+r.URL.Path, "grey") + " " +
			fmtcode(dummyw.Code) + " " +
			c(fmtduration(t)+" "+
				sizestr.ToString(int64(dummyw.Body.Len())), "grey"))
	})
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	//shorthand
	reply := func(c int, msg string) {
		w.WriteHeader(c)
		if msg != "" {
			w.Write([]byte(msg))
		}
	}
	//requested file
	p := filepath.Join(s.c.Dir, path)
	//check file or dir
	isdir := false
	if info, err := os.Stat(p); err == nil {
		//found! -> is file or dir?
		isdir = info.IsDir()
	} else if err != nil {
		//missing! -> no-pushstate or has extension?
		if !s.c.PushState || filepath.Ext(p) != "" {
			//not found! handle with proxy?
			if s.fallback != nil {
				s.fallback.ServeHTTP(w, r)
				return
			}
			//not found!!
			reply(404, "Not found")
			return
		}
		p = s.root //known to exist
		isdir = false
	}

	//optionally use index instead of directory list
	if isdir && !s.c.NoIndex {
		dirindex := filepath.Join(p, "index.html")
		if _, err := os.Stat(dirindex); err == nil {
			p = dirindex
			isdir = false
		}
	}

	//directory list
	if isdir {
		if s.c.NoList {
			reply(403, "Listing not allowed")
			return
		}
		s.dirlist(w, r, p)
		return
	}

	//check file again
	info, err := os.Stat(p)
	if err != nil {
		reply(404, "Not found")
		return
	}

	//stream file
	f, err := os.Open(p)
	if err != nil {
		reply(500, err.Error())
		return
	}

	//add all served file's parent dirs to the watcher
	if s.c.LiveReload {
		dir, _ := filepath.Split(p)
		if _, ok := s.watching[dir]; !ok {
			s.watcher.Add(dir)
			s.watching[dir] = true
		}
	}

	//by default, caching is disabled
	if !s.c.Caching {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}

	//http.ServeContent handles caching and range requests
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

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

	//create encodable
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

	//return in acceptable format
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
				s := fmt.Sprintf("<a href=\"%s\">\n\t%s\n</a><br>\n",
					filepath.Join(list.Path, f.Name), f.Name)
				buff.WriteString(s)
			}
		default:
			continue
		}
		contype = accept
		break
	}
	//no match
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
