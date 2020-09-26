package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/julienschmidt/httprouter"
)

func main() {
	mux := httprouter.New()

	mux.GET("/", func(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		http.ServeFile(res, req, "pages/todo/index.html")
	})

	mux.HandlerFunc(http.MethodGet, "/assets/*filepath", AssetHandler(osfs.New("assets")))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}

func AssetHandler(fs billy.Filesystem) http.HandlerFunc {

	hasHiddenPart := func(fp string) bool {
		dir := fp
		for {
			var file string
			dir, file = path.Split(dir)

			if strings.HasPrefix(file, ".") {
				return true
			}

			if dir == "/" || dir == "" || file == "" {
				break
			}
		}

		return false
	}

	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Del("content-type")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			res.Header().Set("content-type", "application/wasm")
		}
		fp := filepath.Clean(httprouter.ParamsFromContext(req.Context()).ByName("filepath"))
		if hasHiddenPart(fp) {
			http.Error(res, "file not found", http.StatusNotFound)
			return
		}
		s, err := fs.Stat(fp)
		if err != nil {
			http.Error(res, "file not found", http.StatusNotFound)
			return
		}
		f, err := fs.Open(fp)
		if err != nil {
			http.Error(res, "file not found", http.StatusNotFound)
			return
		}
		defer func() { _ = f.Close() }()
		http.ServeContent(res, req, fp, s.ModTime(), f)
	}
}
