package restful

import (
	"bytes"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	webresources "github.com/ak47less/step1go/lite2/src/webui/resources"
)

var (
	spaAssets, spaAssetsErr = fs.Sub(webresources.WWW, "www")
)

func (inst *innerWebServer) handleSPA(w http.ResponseWriter, r *http.Request) {
	if isAPIPath(r.URL.Path) {
		inst.handleNotFound(w, r)
		return
	}
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	if spaAssetsErr != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"web application assets are unavailable", nil)
		return
	}

	assetName := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if assetName == "" || strings.HasSuffix(r.URL.Path, "/") {
		assetName = "index.html"
	}
	if serveSPAAsset(w, r, assetName) {
		return
	}

	// Client-side routes do not have a file extension and are resolved by the
	// SPA's own router. Missing assets keep a 404 instead of returning HTML.
	if path.Ext(assetName) == "" {
		if serveSPAAsset(w, r, "index.html") {
			return
		}
	}
	inst.handleNotFound(w, r)
}

func isAPIPath(requestPath string) bool {
	return requestPath == "/api" || strings.HasPrefix(requestPath, "/api/")
}

func serveSPAAsset(w http.ResponseWriter, r *http.Request, name string) bool {
	data, err := fs.ReadFile(spaAssets, name)
	if err != nil {
		return false
	}

	contentType := mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if path.Ext(name) == ".html" {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	return true
}
