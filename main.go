package main

import (
	"log"

	"github.com/jpillora/opts"
	"github.com/jpillora/serve/server"
)

var VERSION string = "0.0.0"

func main() {

	//defaults
	c := server.Config{
		Directory: ".",
		Host:      "0.0.0.0",
		Port:      3000,
		TimeFmt:   "[2006-01-02 15:04:05.000]",
	}

	//parse
	opts.New(&c).
		Version(VERSION).
		PkgRepo().
		Parse()

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
