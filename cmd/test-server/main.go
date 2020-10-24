package main

import (
	"net/http"

	"github.com/calebamiles/example-money-making-fortune-api/service"
)

func main() {
	// Don't use Cadence backend
	http.HandleFunc("/fortune", service.HandleGetFigletizedFortuneDirect)
	http.HandleFunc("/healthz", service.HandleGetHealthz)

	http.ListenAndServe(":8092", nil)
}
