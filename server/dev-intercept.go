package server

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/jpillora/sizestr"
)

func (s *Server) devIntercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//capture response
		dummyw := httptest.NewRecorder()
		//track timing
		t0 := time.Now()
		//perform response
		next.ServeHTTP(dummyw, r)
		//log result
		t := time.Now().Sub(t0)

		//use etag instead of last-modified
		// dummyw.Header().Del("Last-Modified")
		//use all other headers
		for name, _ := range dummyw.HeaderMap {
			w.Header().Set(name, dummyw.Header().Get(name))
		}

		body := dummyw.Body.Bytes()

		//hash file first time through
		//subsequent times are empty due to cache match
		if len(body) > 0 {
			h := md5.New()
			h.Write(body)
			etag := `"` + hex.EncodeToString(h.Sum(nil)) + `"`
			w.Header().Set("ETag", etag)
		}

		//actual response
		w.WriteHeader(dummyw.Code)
		if len(body) > 0 {
			w.Write(body)
		}

		//log result
		if !s.c.Quiet {
			fmt.Print(c(r.Method+" "+r.URL.Path, "grey") + " " +
				fmtcode(dummyw.Code) + " " +
				c(fmtduration(t)+" "+
					sizestr.ToString(int64(len(body))), "grey"))
			// fmt.Print(" " + etag)
			fmt.Println()
		}
	})
}
