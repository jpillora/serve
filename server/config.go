package server

//Config is a server configuration
type Config struct {
	Directory  string `type:"arg" help:"[directory] from which files will be served"`
	Host       string `help:"Host interface"`
	Port       int    `help:"Listening port"`
	LiveReload bool   `help:"Enable LiveReload, a websocket server, which triggers page a refresh after each file change"`
	PushState  bool   `help:"Enable PushState mode, causes missing directory paths will return the root index.html file, instead of returning a 404. This allows correct use of the HTML5 History API"`
	NoIndex    bool   `help:"Disable automatic loading of index.html"`
	NoList     bool   `help:"Disable directory listing"`
	Quiet      bool   `help:"Disable all output"`
	FastMode   bool   `help:"Requests are not hashed and measured, useful for serving large files"`
	Open       bool   `help:"Automatically runs the 'open' command to open the listening page in the default browser"`
	Fallback   string `help:"A proxy path to request if a given request 404's. This allows you customize one file of a live site"`
}
