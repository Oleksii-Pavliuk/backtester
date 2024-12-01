package processor

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

type Api struct {
	Address   string
	Port      int
	Processor *Processor
	Router    *chi.Mux
	handlers  map[string]time.Time
}

type ErrResponse struct {
	HTTPStatusCode int
	Message        string
}

func (a *Api) initRouter() {
	a.Router = chi.NewRouter()
	a.Router.Route("/process", func(router chi.Router) {
		router.Post("/", a.StartTaskHandler)
	})
	a.Router.HandleFunc("/socket/{id}", a.StartStreamHandler)
}

func (a *Api) Start() {
	a.initRouter()
	log.Printf("Starting process API on http://%s:%d", a.Address, a.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
}
