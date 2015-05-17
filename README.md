
# serve

A basic HTTP file server in Go

[![GoDoc](https://godoc.org/github.com/jpillora/serve/server?status.svg)](https://godoc.org/github.com/jpillora/serve/server)

### Features

* Fast
* Single binary
* Colourful logs with modifiable timestamps and response times
* Directory listing supporting multiple content types (`html`,`json` and `xml`) via the `Accept` header
* Directory downloads via on-demand `zip` and `tar` [archive](https://github.com/jpillora/archive)s
* Optional PushState (HTML5 History API) mode (missing directories returns the root)
* LiveReload (combines with [this Chrome extension](https://chrome.google.com/webstore/detail/livereload/jnihajbhpnppcggbcgedagnkighmdlei?hl=en))
* Fallback proxy (missing requests defer to another server)

### Install

**Binaries**

See [the latest release](https://github.com/jpillora/serve/releases/latest) or download it with this one-liner: `curl i.jpillora.com/serve | bash`

**Source**

``` sh
$ go get -v github.com/jpillora/serve
```

### Usage

```
serve --help
```

<tmpl,code: go run main.go --help>
``` plain 

  Usage: serve [options] [directory]
  
  [directory] from which files will be served (default .)
  
  Options:
  --host, -h        Host interface (default 0.0.0.0)
  --port, -p        Listening port (default 3000)
  --livereload, -l  Enable LiveReload, a websocket server, which triggers 
                    page a refresh after each file change
  --pushstate       Enable PushState mode, causes missing directory paths 
                    will return the root index.html file, instead of returning 
                    a 404. This allows correct use of the HTML5 History API
  --noindex, -n     Disable automatic loading of index.html
  --noslash         Disables automatic slash insertion when loading an 
                    index.html or directory
  --nolist          Disable directory listing
  --noarchive       Disable directory archiving (download directories by 
                    appending .zip .tar .tar.gz, archives are streamed 
                    without buffering)
  --quiet, -q       Disable all output
  --timefmt, -t     Timestamp format (default [2006-01-02 15:04:05.000] )
  --open, -o        On server startup, open the root in the default browser 
                    (uses the 'open' command)
  --fallback, -f    Requests that yeild a 404, will instead proxy through 
                    to the provided path (swaps in the appropriate Host 
                    header)
  --help          
  --version, -v   
  
  Version:
    0.0.0
  
  Read more:
    https://github.com/jpillora/serve
  
```
</tmpl>

#### Credits

TJ's [serve](https://github.com/tj/serve)

#### MIT License

Copyright © 2015 Jaime Pillora &lt;dev@jpillora.com&gt;

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.