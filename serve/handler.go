package serve

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jaschaephraim/lrserver"
	"github.com/jpillora/archive"
	"github.com/jpillora/cookieauth"
	"github.com/jpillora/requestlog"
	"gopkg.in/fsnotify.v1"
)

//Handler is custom file server
type Handler struct {
	c            Config
	root         string
	hasIndex     bool
	servedMut    sync.Mutex
	served       map[string]bool
	fallback     *httputil.ReverseProxy
	fallbackHost string
	watcherMut   sync.Mutex
	watcher      *fsnotify.Watcher
	watching     map[string]bool
	lr           *lrserver.Server
}

//NewServer creates a new Server
func NewHandler(c Config) (http.Handler, error) {
	s := &Handler{
		c:      c,
		served: map[string]bool{},
	}
	_, err := os.Stat(c.Directory)
	if c.Directory == "" || err != nil {
		return nil, fmt.Errorf("Missing directory: %s", c.Directory)
	}

	if c.PushState {
		s.root = filepath.Join(c.Directory, "index.html")
		if _, err := os.Stat(s.root); err != nil {
			return nil, fmt.Errorf("'%s' is required for pushstate", s.root)
		}
		s.hasIndex = true
	}

	if c.Fallback != "" {
		u, err := url.Parse(c.Fallback)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(u.Scheme, "http") {
			return nil, fmt.Errorf("Invalid fallback protocol scheme")
		}
		s.fallbackHost = u.Host
		s.fallback = httputil.NewSingleHostReverseProxy(u)
	}

	if c.LiveReload {
		s.lr = lrserver.New("serve-lr", lrserver.DefaultPort)
		discard := log.New(ioutil.Discard, "", 0)
		s.lr.SetErrorLog(discard)
		s.lr.SetStatusLog(discard)
		s.watching = map[string]bool{}
		s.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
	}

	if s.c.LiveReload {
		go func() {
			if err := s.lr.ListenAndServe(); err != nil {
				fmt.Printf("LiveReload server closed: %s", err)
			}
		}()
		go func() {
			for {
				event := <-s.watcher.Events
				switch event.Op {
				case fsnotify.Create, fsnotify.Rename, fsnotify.Write:
					s.lr.Reload(event.Name)
				case fsnotify.Remove:
					if s.watching[event.Name] {
						delete(s.watching, event.Name)
					}
				}
			}
		}()
	}

	h := http.Handler(s)
	//basic auth
	if c.Auth != "" {
		auth := strings.SplitN(c.Auth, ":", 2)
		if len(auth) < 2 {
			return nil, fmt.Errorf("should be in the form 'user:pass'")
		}
		h = cookieauth.Wrap(h, auth[0], auth[1])
	}
	//logging is enabled
	if !c.Quiet {
		h = requestlog.WrapWith(h, requestlog.Options{TimeFormat: c.TimeFmt})
	}
	//listen
	return h, nil
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	//shorthand
	reply := func(c int, msg string) {
		w.WriteHeader(c)
		if msg != "" {
			w.Write([]byte(msg))
		}
	}
	//requested file
	p := filepath.Join(s.c.Directory, path)
	//check file or dir
	isdir := false
	missing := false
	if info, err := os.Stat(p); err != nil {
		missing = true
	} else {
		isdir = info.IsDir()
	}

	if s.c.PushState && missing && filepath.Ext(p) == "" {
		//missing and pushstate and no ext
		p = s.root //change to request for the root
		isdir = false
		missing = false
	}

	if s.fallback != nil && (missing || isdir) {
		//fallback proxy enabled
		r.Host = s.fallbackHost
		s.fallback.ServeHTTP(w, r)
		return
	}

	if !s.c.NoArchive && missing {
		//check if is archivable
		ok := false
		ext := archive.Extension(p)
		dir := ""
		if ext != "" {
			var err error
			if dir, err = filepath.Abs(strings.TrimSuffix(p, ext)); err == nil {
				if info, err := os.Stat(dir); err == nil && info.IsDir() {
					ok = true
				}
			}
		}
		if ok {
			w.Header().Set("Content-Type", mime.TypeByExtension(ext))
			w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(dir)+ext)
			w.WriteHeader(200)
			//write archive
			a, _ := archive.NewWriter(ext, w)
			if err := a.AddDir(dir); err != nil {
				w.Write([]byte("\n\nERROR: " + err.Error()))
				return
			}
			if err := a.Close(); err != nil {
				w.Write([]byte("\n\nERROR: " + err.Error()))
				return
			}
			return
		}
	}

	if !isdir && missing {
		//file not found!!
		reply(404, "Not found")
		return
	}

	//force trailing slash
	if isdir && !s.c.NoSlash && !strings.HasSuffix(path, "/") {
		w.Header().Set("Location", path+"/")
		w.WriteHeader(302)
		w.Write([]byte("Redirecting (must use slash for directories)"))
		return
	}

	//optionally use index instead of directory list
	if isdir && !s.c.NoIndex {
		dirindex := filepath.Join(p, "index.html")
		if _, err := os.Stat(dirindex); err == nil {
			p = dirindex
			isdir = false
		}
	}

	//directory list
	if isdir {
		if s.c.NoList {
			reply(403, "Listing not allowed")
			return
		}
		s.dirlist(w, r, p)
		return
	}

	//check file again
	info, err := os.Stat(p)
	if err != nil {
		reply(404, "Not found")
		return
	}

	//stream file
	f, err := os.Open(p)
	if err != nil {
		reply(500, err.Error())
		return
	}

	//add all served file's parent dirs to the watcher
	if s.c.LiveReload {
		dir, _ := filepath.Split(p)
		s.watcherMut.Lock()
		if _, watching := s.watching[dir]; !watching {
			if err := s.watcher.Add(dir); err == nil {
				s.watching[dir] = true
			}
		}
		s.watcherMut.Unlock()
	}

	modtime := info.ModTime()
	//first time - dont use cache
	s.servedMut.Lock()
	if !s.served[p] {
		s.served[p] = true
		modtime = time.Now()
	}
	s.servedMut.Unlock()

	//http.ServeContent handles caching and range requests
	http.ServeContent(w, r, info.Name(), modtime, f)
}
