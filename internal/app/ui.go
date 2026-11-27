package app

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// serveFrontend serves built UI files for GET and HEAD.
// Existing files are served directly. Missing paths that look like assets
// return 404; other missing paths serve index.html for SPA client-side routing.
func (a *Application) serveFrontend(w http.ResponseWriter, r *http.Request) {
	urlPath := cleanURLPath(r.URL.Path)

	if frontendPathExists(a.deps.Frontend, urlPath) {
		a.fileServer.ServeHTTP(w, r)
		return
	}
	if looksLikeStaticAsset(urlPath) {
		http.NotFound(w, r)
		return
	}

	// Serve index.html in place so the browser URL stays unchanged.
	// Routing through FileServer as /index.html would redirect to /.
	serveFrontendFile(w, r, a.deps.Frontend, "index.html")
}

// serveFrontendFile writes one file from fsys, honouring GET and HEAD.
func serveFrontendFile(
	w http.ResponseWriter,
	r *http.Request,
	fsys fs.FS,
	name string,
) {
	file, err := fsys.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(
			w,
			"frontend file is not seekable",
			http.StatusInternalServerError,
		)
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), seeker)
}

// frontendPathExists reports whether urlPath maps to a file FileServer
// can serve, including a directory that contains index.html.
func frontendPathExists(fsys fs.FS, urlPath string) bool {
	name := strings.TrimPrefix(urlPath, "/")
	if name == "" {
		_, err := fs.Stat(fsys, "index.html")
		return err == nil
	}

	info, err := fs.Stat(fsys, name)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return true
	}
	_, err = fs.Stat(fsys, path.Join(name, "index.html"))
	return err == nil
}

// looksLikeStaticAsset reports whether urlPath should 404 when missing
// rather than fall back to index.html. /assets, paths under /assets/, and
// any path whose final segment contains a '.' count as assets.
func looksLikeStaticAsset(urlPath string) bool {
	if urlPath == "/assets" || strings.HasPrefix(urlPath, "/assets/") {
		return true
	}
	return strings.Contains(path.Base(urlPath), ".")
}
