package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/jpillora/opts"
	"github.com/jpillora/requestlog"
	"github.com/jpillora/serve/serve"
)

var version string = "0.0.0"

type Config struct {
	Host         string `help:"Host interface"`
	Port         int    `help:"Listening port"`
	Open         bool   `help:"On server startup, open the webroot in the default browser (uses the 'open' command)"`
	KeyPath      string `help:"TLS Key file path"`
	CertPath     string `help:"TLS Certicate file path"`
	serve.Config `type:"embedded"`
}

func main() {

	//defaults
	c := Config{
		Host: "0.0.0.0",
		Port: 3000,
		Config: serve.Config{
			Directory: ".",
		},
	}

	//parse
	opts.New(&c).
		Name("serve").
		Version(version).
		Repo("github.com/jpillora/serve").
		Parse()

	isTLS := c.CertPath != "" || c.KeyPath != "" //poor man's XOR
	if isTLS && (c.CertPath == "" || c.KeyPath == "") {
		log.Fatal("You must provide both key and cert paths")
	}

	//ready!
	h, err := serve.NewHandler(c.Config)
	if err != nil {
		log.Fatal(err)
	}

	port := strconv.Itoa(c.Port)

	if c.Open {
		go func() {
			proto := "http"
			if isTLS {
				proto += "s"
			}
			host := c.Host
			if host == "0.0.0.0" {
				host = "localhost"
			}
			time.Sleep(500 * time.Millisecond)
			cmd := exec.Command("open", proto+"://"+host+":"+port)
			cmd.Run()
		}()
	}

	fmt.Printf("%sserving %s%s %son port %s%d%s\n",
		requestlog.DefaultOptions.Colors.Grey,
		requestlog.DefaultOptions.Colors.Cyan, c.Config.Directory,
		requestlog.DefaultOptions.Colors.Grey,
		requestlog.DefaultOptions.Colors.Cyan, c.Port,
		requestlog.DefaultOptions.Colors.Reset,
	)

	server := http.Server{
		Addr:    c.Host + ":" + port,
		Handler: h,
	}

	if isTLS {
		log.Fatal(server.ListenAndServeTLS(c.CertPath, c.KeyPath))
	}
	log.Fatal(server.ListenAndServe())
}
