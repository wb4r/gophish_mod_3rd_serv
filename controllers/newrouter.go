package controllers

import (
	"compress/gzip"
	"context"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/controllers/api"
	log "github.com/gophish/gophish/logger"
	mid "github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gophish/gophish/worker"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/unindexed"
)

type PhishServerOption2 func(*PhishingServer2)
type PhishingServer2 struct {
	server *http.Server
	worker worker.Worker
	config config.PhishServer2
}

// func WithWorker(w worker.Worker) PhishServerOption2 {
// 	return func(pst *PhishServer2) {
// 		pst.worker = w
// 	}
// }

func NewPhishingServer2(config config.PhishServer2, options ...PhishServerOption2) *PhishingServer2 {
	defaultServer := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         config.ListenURL,
	}
	pst := &PhishingServer2{
		server: defaultServer,
		config: config,
	}
	for _, opt := range options {
		opt(pst)
	}
	pst.registerRoutes()
	return pst
}

func (pst *PhishingServer2) Start() error {
	if pst.worker != nil {
		go pst.worker.Start()
	}
	if pst.config.UseTLS {
		err := util.CheckAndCreateSSL(pst.config.CertPath, pst.config.KeyPath)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Infof("Starting PHISHADDITIONALSERVER server at https://%s", pst.config.ListenURL)
		return pst.server.ListenAndServeTLS(pst.config.CertPath, pst.config.KeyPath)
	}
	// If TLS isn't configured, just listen on HTTP
	log.Infof("Starting PHISHADDITIONALSERVER server at http://%s", pst.config.ListenURL)
	return pst.server.ListenAndServe()
}

func (pst *PhishingServer2) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return pst.server.Shutdown(ctx)
}

func (pst *PhishingServer2) registerRoutes() {
	router := mux.NewRouter()
	router.HandleFunc("/", pst.Redir)
}

func (pst *PhishingServer2) Redir(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.google.com", 301)
	return
}
