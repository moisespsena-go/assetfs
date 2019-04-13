package assetfs

import (
	"crypto/md5"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/moisespsena-go/assetfs/assetfsapi"
)

var cacheSince = time.Now().Format(http.TimeFormat)

func HTTPStaticHandler(fs assetfsapi.Interface) http.Handler {
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
			var data []byte
			if data, err = asset.Data(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
