package main

import (
	"linkShortener/internal/handlers"
	_ "linkShortener/internal/handlers"
	"linkShortener/pkg/middlewares"
	_ "linkShortener/templates"
	htmlTemplates "linkShortener/templates"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus := logrus.New()

	r := chi.NewRouter()

	r.Route("/links", func(r chi.Router) {
		r.Use(middlewares.ManageLongLink)

		r.Post("/", handlers.AddLink)
	})
	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.ManageShortLink)

		r.Get("/*", handlers.RedirectLink)
		r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
			err := htmlTemplates.GetTmpls().ExecuteTemplate(rw, "index.html", htmlTemplates.IndexPageData{})
			if err != nil {
				logrus.Error("Error sending index.html: ", err)
			}
		})
	})

	err := http.ListenAndServe(":15001", r)
	if err == nil {
		logrus.Fatal("Error closing connection: ", err)
	}
}
