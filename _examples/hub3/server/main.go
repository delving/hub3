package main

import (
	"log"

	"github.com/delving/hub3/pkg/server/http"
	"github.com/delving/hub3/pkg/server/http/handlers"
)

func main() {
	routers := []http.RouterCallBack{
		handlers.RegisterBulkIndexer,
		handlers.RegisterCSV,
		handlers.RegisterDatasets,
		handlers.RegisterStaticAssets,
		handlers.RegisterEAD,
		handlers.RegisterElasticSearchProxy,
		handlers.RegisterLOD,
		handlers.RegisterLinkedDataFragments,
		handlers.RegisterSparql,
	}
	server, err := http.NewServer(
		http.SetIntroSpection(true),
		http.SetRouters(routers...),
		http.SetPort(3010),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
