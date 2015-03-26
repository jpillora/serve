package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jpillora/ansi"
)

var (
	host      = flag.String("host", "0.0.0.0", "")
	port      = flag.Int("port", 3000, "")
	pushstate = flag.Bool("pushstate", false, "")
	nodirlist = flag.Bool("nodirlist", false, "")
	noindex   = flag.Bool("noindex", false, "")
	open      = flag.Bool("open", false, "")
	dir       = ""
)

const help = `
	Usage: serve [options] [directory]

	Serves the files in [directory], where [directory]
	defaults to the current working directory.

	Options:

	--host, Host interface (defaults to 0.0.0.0)
	--port, Listening port (defaults to 3000)
	--pushstate, Missing paths (with no extension)
	will return 200 and the root index.html file,
	instead of returning of 404 Not found.
	--nodirlist, Disable directory listing.
	--noindex, Disable use of index.html automatic
	redirection.
	--open, Automatically runs the 'open' command
	to open the listening page in the default
	browse.
	--help, This help text.

	Read more:
	  https://github.com/jpillora/serve
`

var root = ""
var hasIndex = false

func main() {
	he := flag.Bool("help", false, "")
	h := flag.Bool("h", false, "")

	flag.Parse()
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, help)
		os.Exit(1)
	}
	if *he || *h {
		flag.Usage()
	}

	port := strconv.Itoa(*port)
	addr := (*host) + ":" + port

	args := flag.Args()
	if len(args) == 0 {
		dir, _ = os.Getwd()
	} else {
		dir = args[0]
	}

	_, err := os.Stat(dir)
	if dir == "" || err != nil {
		fmt.Printf("\n\tMissing directory: %s\n", dir)
		flag.Usage()
	}

	if *pushstate {
		root = filepath.Join(dir, "index.html")
		if _, err := os.Stat(root); err != nil {
			fmt.Printf("\n\t'%s' is required for pushstate\n", root)
			flag.Usage()
		}
		hasIndex = true
	}

	if *open {
		go func() {
			time.Sleep(500 * time.Millisecond)
			cmd := exec.Command("open", "http://localhost:"+port)
			cmd.Run()
		}()
	}

	fmt.Println(c("serving ", "grey") +
		c(shorten(dir), "cyan") +
		c(" on port ", "grey") +
		c(port, "cyan"))
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(handle)))
}

func handle(w http.ResponseWriter, r *http.Request) {

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
		fmt.Println(c(r.Method+" "+r.URL.Path, "grey") +
			" " + fmtcode(code) + " " +
			c(fmtduration(t)+size, "grey"))
	}()

	//requested file
	p := filepath.Join(dir, path)

	//check file or dir
	isdir := false
	if info, err := os.Stat(p); err == nil {
		//is file or dir
		isdir = info.IsDir()
	} else if err != nil {
		//no-pushstate or has extension
		if !*pushstate || filepath.Ext(p) != "" {
			reply(404, "Not found")
			return
		}
		p = root //known to exist
		isdir = false
	}

	//optionally use index instead of directory list
	if isdir && !*noindex {
		dirindex := filepath.Join(p, "index.html")
		if _, err := os.Stat(dirindex); err == nil {
			p = dirindex
			isdir = false
		}
	}

	//directory list
	if isdir {
		if *nodirlist {
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

type spyFile struct {
	*os.File
	read int64
}

func (s *spyFile) Read(p []byte) (int, error) {
	n, err := s.File.Read(p)
	s.read += int64(n)
	return n, err
}

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
