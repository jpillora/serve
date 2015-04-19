package server

//Config is a server configuration
type Config struct {
	Dir        string `arg:"directory" help:"Root [directory] from which files will be served"`
	Host       string `help:"Host interface"`
	Port       int    `help:"Listening port"`
	LiveReload bool   `help:"Enable LiveReload, a websocket server, which triggers page a refresh after each file change"`
	PushState  bool   `help:"Enable PushState mode, causes missing directory paths will return the root index.html file, instead of returning a 404. This allows correct use of the HTML5 History API"`
	NoList     bool   `help:"Disable directory listing"`
	NoIndex    bool   `help:"Disable use of index.html automatic redirection"`
	NoLogging  bool   `help:"Disable logging"`
	Open       bool   `help:"Automatically runs the 'open' command to open the listening page in the default browse"`
	Fallback   string `help:"A proxy path to request if a given request 404's. This allows you customize one file of a live site"`
}
