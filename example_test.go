package rest_test

import (
	"net/http"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
)

func getUser(w http.ResponseWriter, r *http.Request) (int, error) {
	name := rest.Params(r).Get("name")
	data := rest.JSON{"name": name}
	return rest.Write(w, r, http.StatusOK, data)
}

func getFile(w http.ResponseWriter, r *http.Request) (int, error) {
	name := rest.Params(r).Get("filename")
	http.ServeFile(w, r, name)
	return http.StatusOK, nil
}

func Example() {
	mux := &rest.ServeMux{
		Headers: map[string]string{
			"X-API-Version": "1.0",
		},
		NotCompress: false,
		Options: &rest.Options{
			Encoder:       rest.JSONEncoder(true),
			DataAdapter:   rest.Adapter,
			AllowMultiple: false,
		},
		Debug:  true,
		Logger: log.Default,
	}
	mux.Handle("GET", "/user/:name", getUser)
	mux.Handle("GET", "/files/*filename", getFile)

	http.ListenAndServe(":8080", mux)
}
