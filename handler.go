package assetfs

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/moisespsena-go/assetfs/assetfsapi"
)

var cacheSince = time.Now().Format(http.TimeFormat)

type StaticHandler struct {
	EtagTimeLife time.Duration
	cacheSince   string

	FS   assetfsapi.Interface
	etag map[string]struct {
		last time.Time
		sum  []byte
	}
	etagMu sync.RWMutex
}

func (this *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.ServeAsset(w, r, r.URL.Path, true)
}

func NewStaticHandler(fs assetfsapi.Interface) *StaticHandler {
	return &StaticHandler{
		FS:           fs,
		EtagTimeLife: time.Hour,
		cacheSince:   time.Now().Format(http.TimeFormat),
	}
}

func (this *StaticHandler) getEtag(asset assetfsapi.FileInfo) (etag string, err error) {
	var key = fmt.Sprint(asset.Path(), asset.ModTime(), asset.Size())
	if this.etag != nil {
		func() {
			defer this.etagMu.RUnlock()
			this.etagMu.RLock()
			if value, ok2 := this.etag[key]; ok2 && value.last.Add(this.EtagTimeLife).After(time.Now()) {
				etag = string(value.sum)
			}
		}()
		if etag != "" {
			return
		}
	}

	hash := md5.New()
	var r io.ReadCloser
	if r, err = asset.Reader(); err != nil {
		return
	}
	defer r.Close()

	if _, err = io.Copy(hash, r); err != nil {
		return
	}

	sum := hash.Sum(nil)
	etag = fmt.Sprintf("%x", sum)

	this.etagMu.Lock()
	defer this.etagMu.Unlock()

	if this.etag == nil {
		this.etag = map[string]struct {
			last time.Time
			sum  []byte
		}{}
	}
	this.etag[key] = struct {
		last time.Time
		sum  []byte
	}{time.Now(), sum}
	return
}

func (this *StaticHandler) ServeAsset(w http.ResponseWriter, r *http.Request, pth string, notFound ...bool) {
	if r.Header.Get("If-Modified-Since") == this.cacheSince {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Last-Modified", this.cacheSince)

	if fspath := RootPath(this.FS); fspath != "" {
		pth = strings.TrimPrefix(pth, fspath)
	}

	pth = strings.TrimPrefix(pth, "/")

	if asset, err := this.FS.AssetInfoC(r.Context(), pth); err == nil {
		var etag string
		if etag, err = this.getEtag(asset); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		if ctype := mime.TypeByExtension(filepath.Ext(pth)); ctype != "" {
			w.Header().Set("Content-Type", ctype)
		}

		w.Header().Set("Cache-control", "private, must-revalidate, max-age=300")
		w.Header().Set("ETag", etag)
		var reader io.ReadCloser
		if reader, err = asset.Reader(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer reader.Close()
		io.Copy(w, reader)
		return
	}

	if len(notFound) > 0 && notFound[0] {
		http.NotFound(w, r)
	}
}

func HttpStaticHandler(fs assetfsapi.Interface) http.Handler {
	return NewStaticHandler(fs)
}

func RootPath(fs assetfsapi.Interface) string {
	pth := strings.Replace(fs.GetPath(), "\\", "/", -1)
	if pth != "" {
		pth = "/" + pth
	}
	return pth
}
