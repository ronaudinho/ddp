package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ronaudinho/ddp"
)

func main() {
	path := "/proxy"
	store := ddp.NewMockStore()
	logger := ddp.NewLogger(store)
	prx := ddp.NewProxy(path, logger)

	// TODO other routing more performant than http.Handle
	http.Handle(path, prx)
	http.HandleFunc("/stale", func(w http.ResponseWriter, r *http.Request) {
		data, err := logger.Get("some time limit")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		b, _ := json.Marshal(data)
		w.Write(b)
	})

	// TODO graceful shutdown
	log.Fatal(http.ListenAndServe(":12345", nil))
}
