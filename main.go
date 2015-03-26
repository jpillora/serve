package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
		fmt.Fprintf(os.Stderr, help)
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

	fmt.Println(
		c("serving ", "grey") +
			c(shorten(dir), "cyan") +
			c(" on port ", "grey") +
			c(port, "cyan"))
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(handle)))
}

func handle(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	//reporting
	stats := &struct{ sent, code int }{}
	reply := func(code int, msg string) {
		w.WriteHeader(code)
		stats.code = code
		if msg != "" {
			w.Write([]byte(msg))
		}
	}
	t0 := time.Now()
	defer func() {
		t := time.Now().Sub(t0)
		size := ""
		if stats.sent != 0 {
			size = " " + tobyte(stats.sent)
		}
		fmt.Println(
			c(r.Method+" "+r.URL.Path, "grey") +
				" " + sc(stats.code) + " " +
				c(fmt.Sprintf("%s%s", t, size), "grey"))
	}()

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
			s := fmt.Sprintf("%s\n", n)
			buff.WriteString(s)
		}
		reply(200, buff.String())
		return
	}

	//check file again
	if _, err := os.Stat(p); err != nil {
		reply(404, "Not found")
		return
	}

	//stream file
	f, err := os.Open(p)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	reply(200, "")
	n, _ := io.Copy(w, f)
	stats.sent = int(n)
}

var scale = []string{"b", "kb", "mb", "gb"}

func tobyte(n int) string {
	for _, s := range scale {
		if n > 1000 {
			n = n / 1000
		} else {
			return strconv.Itoa(n) + s
		}
	}
	return strconv.Itoa(n)
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

func sc(status int) string {
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

// min := s
// for _, e := range os.Environ() {
// 	kv := strings.SplitN(e, "=", 2)
// 	if len(kv) != 2 {
// 		continue
// 	}
// 	k := kv[0]
// 	v := kv[1]
// 	if strings.HasPrefix(s, v) {
// 		m := "$" + k + strings.TrimPrefix(s, v)
// 		if len(m) < len(min) {
// 			min = m
// 		}
// 	}
// }
// return min
