package function

import (
	"net/http"
	"sync"

	"github.com/delving/hub3/ikuzo"
)

var (
	svr  ikuzo.Server
	err  error
	once = &sync.Once{}
)

func initServer() {
	svr, err = ikuzo.NewServer()
}

// F is the cloud function entrypoint
func F(w http.ResponseWriter, r *http.Request) {
	once.Do(initServer)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	svr.ServeHTTP(w, r)
}
