package main

import (
	"log"
	"os"
	"strconv"

	"github.com/delving/hub3/ikuzo"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Unable to parse port: %#v", err)
		return
	}

	svr, err := ikuzo.NewServer(
		ikuzo.SetPort(p),
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
