package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jpillora/serve/server"
)

var VERSION string = "0.0.0" //set via ldflags

var help = `
	Usage: serve [options] [directory]

	Serves the files in [directory], where [directory]
	defaults to the current working directory.

	Version: ` + VERSION + `

	Options:

	--host, Host interface (defaults to 0.0.0.0)
	--port, Listening port (defaults to 3000)
	--pushstate, Missing paths (with no extension)
	will return 200 and the root index.html file,
	instead of returning of 404 Not found.
	--nodirlist, Disable directory listing.
	--noindex, Disable use of index.html automatic
	redirection.
	--nologging, Disable logging.
	--open, Automatically runs the 'open' command
	to open the listening page in the default
	browse.
	--fallback, A proxy path to request if a given
	request 404's. This allows you customize one
	file of a live site.
	--help, This help text.

	Read more:
	  https://github.com/jpillora/serve
`

func main() {

	//fill server config
	c := &server.Config{}
	flag.StringVar(&c.Host, "host", "0.0.0.0", "")
	flag.IntVar(&c.Port, "port", 3000, "")
	flag.BoolVar(&c.PushState, "pushstate", false, "")
	flag.BoolVar(&c.NoDirList, "nodirlist", false, "")
	flag.BoolVar(&c.NoIndex, "noindex", false, "")
	flag.BoolVar(&c.NoLogging, "nologging", false, "")
	flag.BoolVar(&c.Open, "open", false, "")
	flag.StringVar(&c.FallbackProxy, "fallback", "", "")

	//meta cli
	h := false
	flag.BoolVar(&h, "h", false, "")
	flag.BoolVar(&h, "help", false, "")
	v := false
	flag.BoolVar(&v, "v", false, "")
	flag.BoolVar(&v, "version", false, "")

	//parse
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, help)
		os.Exit(1)
	}
	if v {
		fmt.Println(VERSION)
		os.Exit(0)
	}
	if h {
		flag.Usage()
	}

	//get directory
	args := flag.Args()
	if len(args) == 0 {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		c.Dir = dir
	} else {
		c.Dir = args[0]
	}

	//ready!
	s, err := server.New(c)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}

}
