package assetfs

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/httpu"
)

var cacheSince = time.Now()

type StaticHandler struct {
	EtagTimeLife time.Duration
	cacheSince   string

	FS   assetfsapi.Interface
	etag map[string]struct {
		last time.Time
		sum  []byte
	}
	etagMu sync.RWMutex
	gziped bool
}

func (this *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.ServeAsset(w, r, r.URL.Path)
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
	if fspath := RootPath(this.FS); fspath != "" {
		pth = strings.TrimPrefix(pth, fspath)
	}

	pth = strings.TrimPrefix(pth, "/")

	if asset, err := this.FS.AssetInfoC(r.Context(), pth); err == nil {
		var rc io.ReadCloser
		if rc, err = asset.Reader(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rc.Close()
		if cmpr, ok := rc.(Compresseder); ok && cmpr.Compressed() {
			httpu.ServeContent(w, r, path.Base(pth), cacheSince, rc, func() (int64, error) {
				return asset.Size(), nil
			})
		} else {
			httpu.ServeContent(w, r, path.Base(pth), cacheSince, struct {
				io.ReadCloser
			}{rc}, nil)
		}
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
