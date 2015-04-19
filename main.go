package main

import (
	"log"

	"github.com/jpillora/opts"
	"github.com/jpillora/serve/server"
)

var VERSION string = "0.0.0"

func main() {

	//defaults
	c := &server.Config{
		Dir:  "./",
		Host: "0.0.0.0",
		Port: 3000,
	}

	//parse
	opts.AutoNew(c).
		Version(VERSION).
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
