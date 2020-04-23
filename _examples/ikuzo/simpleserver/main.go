package main

import (
	"log"

	"github.com/delving/hub3/ikuzo"
)

func main() {

	svr, err := ikuzo.NewServer(
		ikuzo.SetPort(3001),
	)

	if err != nil {
		log.Fatalf("unable to initialize ikuzo server: %#v", err)
	}

	err = svr.ListenAndServe()
	if err != nil {
		log.Fatalf("server stopped with an error: %#v", err)
		return
	}
}
