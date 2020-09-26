package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/julienschmidt/httprouter"
)

func TestAssetHandler(t *testing.T) {
	t.Run("hidden file", func(t *testing.T) {
		fs := memfs.New()

		_, _ = fs.Create(".gitignore")

		mux := httprouter.New()
		mux.HandlerFunc(http.MethodGet, "/assets/*filepath", AssetHandler(fs))

		req, _ := http.NewRequest(http.MethodGet, "/assets/.gitignore", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()

		ExpectInt(t, res.StatusCode, http.StatusNotFound)
	})

	t.Run("some file", func(t *testing.T) {
		fs := memfs.New()

		_, _ = fs.Create("file")

		mux := httprouter.New()
		mux.HandlerFunc(http.MethodGet, "/assets/*filepath", AssetHandler(fs))

		req, _ := http.NewRequest(http.MethodGet, "/assets/file", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()

		ExpectInt(t, res.StatusCode, http.StatusOK)
	})

	t.Run("missing file", func(t *testing.T) {
		fs := memfs.New()

		_, _ = fs.Create("file")

		mux := httprouter.New()
		mux.HandlerFunc(http.MethodGet, "/assets/*filepath", AssetHandler(fs))

		req, _ := http.NewRequest(http.MethodGet, "/assets/missing-file", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		res := rec.Result()

		ExpectInt(t, res.StatusCode, http.StatusNotFound)
	})
}

func ExpectInt(t *testing.T, got, exp int) {
	t.Helper()
	if got != exp {
		t.Errorf("%d does not equal expected %d", got, exp)
	}
}
