package handlers

import (
	"github.com/gorilla/mux"
	"main/internal/middlewars"
	"net/http"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router.Use(middlewars.SetContentType)
	router.HandleFunc("/notes/create", Create).Methods(http.MethodPost)
	router.HandleFunc("/notes/read/{id}", Read).Methods(http.MethodGet)
	router.HandleFunc("/notes/update", Update).Queries("id", "{id}").Methods(http.MethodPatch)
	router.HandleFunc("/notes/delete/{id}", Delete).Methods(http.MethodDelete)
	return router
}
