package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//Server is custom file server
type Server struct {
	c        *Config
	addr     string
	port     string
	root     string
	hasIndex bool
	fallback *httputil.ReverseProxy
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

	if c.FallbackProxy != "" {
		u, err := url.Parse(c.FallbackProxy)

		if !strings.HasPrefix(u.Scheme, "http") {
			return nil, fmt.Errorf("Invalid fallback protocol scheme")
		}

		if err != nil {
			return nil, err
		}
		s.fallback = httputil.NewSingleHostReverseProxy(u)
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

	if !s.c.NoLogging {
		fmt.Println(c("serving ", "grey") +
			c(shorten(s.c.Dir), "cyan") +
			c(" on port ", "grey") +
			c(s.port, "cyan"))
	}
	return http.ListenAndServe(s.addr, http.HandlerFunc(s.handle))
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	//reporting
	var sf *spyFile
	var code int
	reply := func(c int, msg string) {
		w.WriteHeader(c)
		code = c
		if msg != "" {
			w.Write([]byte(msg))
		}
	}
	t0 := time.Now()
	defer func() {
		t := time.Now().Sub(t0)
		size := ""
		if sf != nil {
			size = " " + tobyte(sf.read)
		}
		if !s.c.NoLogging {
			fmt.Println(c(r.Method+" "+r.URL.Path, "grey") +
				" " + fmtcode(code) + " " +
				c(fmtduration(t)+size, "grey"))
		}
	}()

	//requested file
	p := filepath.Join(s.c.Dir, path)

	//check file or dir
	isdir := false
	if info, err := os.Stat(p); err == nil {
		//is file or dir
		isdir = info.IsDir()
	} else if err != nil {
		//no-pushstate or has extension
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
		if s.c.NoDirList {
			reply(403, "Listing not allowed")
			return
		}
		files, err := ioutil.ReadDir(p)
		if err != nil {
			reply(500, err.Error())
			return
		}
		buff := &bytes.Buffer{}
		for _, f := range files {
			n := f.Name()
			if f.IsDir() {
				n += "/"
			}
			s := fmt.Sprintf("<a href=\"%s\">\n\t%s\n</a><br>\n", filepath.Join(path, n), n)
			buff.WriteString(s)
		}
		w.Header().Set("Content-Type", "text/html")
		reply(200, buff.String())
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

	//http.ServeContent handles caching and range requests
	code = 200
	sf = &spyFile{File: f}
	http.ServeContent(w, r, info.Name(), info.ModTime(), sf)
}
