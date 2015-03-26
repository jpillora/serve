
# serve

A basic HTTP file server in Go (Golang)

### Features

* Fast
* Colourful logs
* Displays response times
* Supports PushState URLs (Missing paths returns the root)
* Single binary

### Install

**Binaries**

See [Releases](https://github.com/jpillora/serve/releases/latest)

**Source**

``` sh
$ go get -v github.com/jpillora/serve
```
### Usage

`serve --help`

<tmpl,code:serve --help>
```
	Usage: serve [options] [directory]

	Serves the files in [directory], where [directory]
	defaults to the current working directory.

	Options:

	--host, Host interface (defaults to 0.0.0.0)
	--port, Listening port (defaults to 3000)
	--pushstate, Missing paths (with no extension)
	will return 200 and the root index.html file,
	instead of returning of 404 Not found.
	--nodirlist, Disable directory listing.
	--noindex, Disable use of index.html automatic
	redirection.
	--open, Automatically runs the 'open' command
	to open the listening page in the default
	browse.
	--help, This help text.

	Read more:
	  https://github.com/jpillora/serve
```
</tmpl>

#### Credits

TJ's [serve](https://npmjs.com/package/serve)

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