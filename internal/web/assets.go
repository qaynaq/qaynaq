// Package web embeds the built UI assets and serves them with SPA fallback
// and Brotli/gzip pre-compressed variants.
package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed all:dist
var distFS embed.FS

// AssetsFS returns the embedded UI assets rooted at dist/.
func AssetsFS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to sub embedded UI assets")
	}
	return sub
}

// ServeSPA serves the embedded UI. Unknown /assets/ paths return 404 so
// stale-hash references fail loudly; other unknown paths fall back to
// indexFile. Pre-compressed .br/.gz variants are served when the client
// advertises support via Accept-Encoding.
func ServeSPA(assets fs.FS, indexFile string) http.HandlerFunc {
	httpFS := http.FS(assets)
	fileServer := http.FileServer(httpFS)
	return func(w http.ResponseWriter, r *http.Request) {
		reqPath := path.Clean(r.URL.Path)
		if reqPath == "/" || reqPath == "." {
			reqPath = "/" + indexFile
		}

		f, err := httpFS.Open(reqPath)
		if err != nil {
			if os.IsNotExist(err) {
				if strings.HasPrefix(reqPath, "/assets/") {
					http.NotFound(w, r)
					return
				}
				serveIndex(w, r, httpFS, indexFile)
				return
			}
			log.Error().Err(err).Str("path", reqPath).Msg("Error opening file from embedded FS")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_ = f.Close()

		if reqPath == "/"+indexFile {
			serveIndex(w, r, httpFS, indexFile)
			return
		}

		if encoding, variantPath, ok := negotiateEncoding(assets, strings.TrimPrefix(reqPath, "/"), r.Header.Get("Accept-Encoding")); ok {
			serveCompressed(w, r, httpFS, "/"+variantPath, reqPath, encoding)
			return
		}

		fileServer.ServeHTTP(w, r)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request, httpFS http.FileSystem, indexFile string) {
	index, err := httpFS.Open("/" + indexFile)
	if err != nil {
		log.Error().Err(err).Str("file", indexFile).Msg("Failed to open index file from embedded FS")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = index.Close() }()

	fi, err := index.Stat()
	if err != nil {
		log.Error().Err(err).Str("file", indexFile).Msg("Failed to stat index file from embedded FS")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, indexFile, fi.ModTime(), index)
}

func negotiateEncoding(assets fs.FS, assetPath, acceptEncoding string) (string, string, bool) {
	accepted := parseAcceptEncoding(acceptEncoding)
	for _, enc := range []struct {
		name string
		ext  string
	}{
		{"br", ".br"},
		{"gzip", ".gz"},
	} {
		if !accepted[enc.name] {
			continue
		}
		variant := assetPath + enc.ext
		if f, err := assets.Open(variant); err == nil {
			_ = f.Close()
			return enc.name, variant, true
		}
	}
	return "", "", false
}

func parseAcceptEncoding(header string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(header, ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		if semi := strings.IndexByte(token, ';'); semi >= 0 {
			token = strings.TrimSpace(token[:semi])
		}
		out[strings.ToLower(token)] = true
	}
	return out
}

func serveCompressed(w http.ResponseWriter, r *http.Request, httpFS http.FileSystem, variantPath, originalPath, encoding string) {
	f, err := httpFS.Open(variantPath)
	if err != nil {
		log.Error().Err(err).Str("variant", variantPath).Msg("Failed to open pre-compressed variant")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	fi, err := f.Stat()
	if err != nil {
		log.Error().Err(err).Str("variant", variantPath).Msg("Failed to stat pre-compressed variant")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if ct := contentTypeFor(originalPath); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Content-Encoding", encoding)
	w.Header().Set("Vary", "Accept-Encoding")
	http.ServeContent(w, r, filepath.Base(originalPath), fi.ModTime(), f)
}

func contentTypeFor(p string) string {
	ext := filepath.Ext(p)
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}
