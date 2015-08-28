package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/jpillora/ansi"
	"github.com/jpillora/opts"
	"github.com/jpillora/serve/handler"
)

var VERSION string = "0.0.0"

type Config struct {
	Host           string `help:"Host interface"`
	Port           int    `help:"Listening port"`
	Open           bool   `help:"On server startup, open the root in the default browser (uses the 'open' command)"`
	handler.Config `type:"embedded"`
}

func main() {

	//defaults
	c := Config{
		Host: "0.0.0.0",
		Port: 3000,
		Config: handler.Config{
			Directory: ".",
			TimeFmt:   "[2006-01-02 15:04:05.000]",
		},
	}

	//parse
	opts.New(&c).
		Name("serve").
		Version(VERSION).
		Repo("github.com/jpillora/serve").
		Parse()

	//ready!
	h, err := handler.New(c.Config)
	if err != nil {
		log.Fatal(err)
	}

	port := strconv.Itoa(c.Port)

	if c.Open {
		go func() {
			time.Sleep(500 * time.Millisecond)
			cmd := exec.Command("open", "http://localhost:"+port)
			cmd.Run()
		}()
	}

	fmt.Printf("%sserving %s%s %son port %s%d%s\n",
		ansi.BlackBytes,
		ansi.CyanBytes, c.Config.Directory,
		ansi.BlackBytes,
		ansi.CyanBytes, c.Port,
		ansi.ResetBytes,
	)

	log.Fatal(http.ListenAndServe(":"+port, h))
}
