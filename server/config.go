package server

//Config is a server configuration
type Config struct {
	Directory  string `type:"arg" help:"[directory] from which files will be served"`
	Host       string `help:"Host interface"`
	Port       int    `help:"Listening port"`
	LiveReload bool   `help:"Enable LiveReload, a websocket server, which triggers page a refresh after each file change"`
	PushState  bool   `help:"Enable PushState mode, causes missing directory paths will return the root index.html file, instead of returning a 404. This allows correct use of the HTML5 History API"`
	NoIndex    bool   `help:"Disable automatic loading of index.html"`
	NoSlash    bool   `help:"Disables automatic slash insertion when loading an index.html or directory"`
	NoList     bool   `help:"Disable directory listing"`
	NoArchive  bool   `help:"Disable directory archiving (download directories by appending .zip .tar .tar.gz, archives are streamed without buffering)"`
	Quiet      bool   `help:"Disable all output"`
	TimeFmt    string `help:"Timestamp format"`
	Open       bool   `help:"On server startup, open the root in the default browser (uses the 'open' command)"`
	Fallback   string `help:"Requests that yeild a 404, will instead proxy through to the provided path (swaps in the appropriate Host header)"`
}
