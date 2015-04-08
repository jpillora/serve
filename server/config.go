package server

//Config is a server configuration
type Config struct {
	Dir           string
	Host          string
	Port          int
	PushState     bool
	NoDirList     bool
	NoIndex       bool
	NoLogging     bool
	Open          bool
	FallbackProxy string
}
