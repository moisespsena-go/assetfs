package assetfs

import (
	"strings"
	"fmt"
	"crypto/md5"
	"mime"
	"path/filepath"
	"net/http"
	"time"
)

var cacheSince = time.Now().Format(http.TimeFormat)

func HTTPStaticHandler(fs Interface) http.Handler {
	fspath := strings.Replace(fs.GetPath(), "\\", "/", -1)
	if fspath != "" {
		fspath = "/" + fspath
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-Modified-Since") == cacheSince {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Last-Modified", cacheSince)

		requestPath := r.URL.Path

		if fspath != "" {
			requestPath = strings.TrimPrefix(requestPath, fspath)
		}

		requestPath = strings.TrimPrefix(requestPath, "/")

		if asset, err := fs.Asset(requestPath); err == nil {
			data := asset.GetData()
			etag := fmt.Sprintf("%x", md5.Sum(data))
			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			if ctype := mime.TypeByExtension(filepath.Ext(requestPath)); ctype != "" {
				w.Header().Set("Content-Type", ctype)
			}

			w.Header().Set("Cache-control", "private, must-revalidate, max-age=300")
			w.Header().Set("ETag", etag)
			w.Write(data)
			return
		}

		http.NotFound(w, r)
	})
}
