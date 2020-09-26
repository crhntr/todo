package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func main() {
	mux := httprouter.New()

	mux.GET("/", func(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		http.ServeFile(res, req, "pages/todo/index.html")
	})

	fileServer := http.FileServer(http.Dir("assets"))
	mux.GET("/assets/*filepath", func(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		res.Header().Del("content-type")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			res.Header().Set("content-type", "application/wasm")
		}
		http.StripPrefix("/assets", fileServer).ServeHTTP(res, req)
	})

	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), mux))
}
