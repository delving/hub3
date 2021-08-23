package webapp

import (
	"embed"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

//go:embed testdata
var testEmbed embed.FS

func TestNewStaticHandler(t *testing.T) {
	// nolint:gocritic
	is := is.New(t)
	h := NewStaticHandler(testEmbed)

	var lastMod, fileMod string

	t.Run("getting file from embedded FS", func(t *testing.T) {
		// get file from embedded FS
		f, err := testEmbed.Open("testdata/text.txt")
		is.NoErr(err)
		stat, err := f.Stat()
		is.NoErr(err)
		is.True(stat.ModTime().IsZero())
		fileMod = stat.ModTime().Format("Mon, 02 Jan 2006 15:04:05 MST")
	})

	t.Run("retrieve file via handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testdata/text.txt", nil)
		w := httptest.NewRecorder()

		h(w, req)
		res := w.Result()

		defer res.Body.Close()

		is.Equal(res.StatusCode, http.StatusOK)
		is.True(res.ContentLength != 0)
		data, err := ioutil.ReadAll(res.Body)
		is.NoErr(err) // you should be able to read the body
		is.Equal(data, []byte("test data!!\n"))

		lastMod = res.Header.Get("Last-Modified")
		is.True(lastMod != "")
	})

	t.Run("second request should have the same Last-Modified time", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testdata/text.txt", nil)
		w := httptest.NewRecorder()
		h(w, req)

		res := w.Result()
		defer res.Body.Close()

		is.Equal(lastMod, res.Header.Get("Last-Modified"))
		is.True(lastMod != fileMod)
	})

	t.Run("nothing found should return http.StatusNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/testdata/notfound", nil)
		w := httptest.NewRecorder()
		h(w, req)

		res := w.Result()
		defer res.Body.Close()
		is.Equal(res.StatusCode, http.StatusNotFound)
	})
}
