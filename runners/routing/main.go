package main

import (
	"github.com/gorilla/mux"
	"github.com/underscorenygren/metrics/ingest/server"
	"log"
	"net/http"
)

func main() {

	zapper, err := zap.NewProductionConfig().Build()
	if err != nil {
		log.Fatal(err)
	}

	pathMap := map[string]string{
		"one": "one route",
		"two": "two route",
	}
	cfg := server.NewConfig()
	router := cfg.Router
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) error {
		w.Write("rootpath\n")
	})
	router.HandleFunc("/{path}", func(w http.ResponseWriter, req *http.Request) error {
		vars := mux.Vars(req)
		path := vars["path"]
		match := pathMap[path]
		w.Write(fmt.Sprintf("path(%s) match(%s)\n", path, match))
	})
	cfg.SetLogger(zapper)
	zapper.Fatal(cfg.RunForever())
}
