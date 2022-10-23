package main

import (
	"fmt"

	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/dombrga/go-crud-person/pkg/constants"
	"github.com/dombrga/go-crud-person/pkg/handlers"
	"github.com/dombrga/go-crud-person/pkg/helpers"
)

func main() {
	fmt.Println("Starting backend app...")
	var ctx, cancel = helpers.CreateContext()
	defer cancel()

	// connect to mongo
	var client, _ = mongo.Connect(ctx, options.Client().ApplyURI(constants.MONGO_URI))

	// routers
	var router = mux.NewRouter()
	router.HandleFunc("/persons", handlers.HandleGetPersons(client)).Methods(http.MethodGet)
	router.HandleFunc("/person/{id}", handlers.HandleGetPerson(client)).Methods(http.MethodGet)
	router.HandleFunc("/person", handlers.HandleCreatePerson(client)).Methods(http.MethodPost)
	router.HandleFunc("/person/{id}", handlers.HandleUpdatePerson(client)).Methods(http.MethodPut)
	router.HandleFunc("/hard-delete-person/{id}", handlers.HandleHardDeletePerson(client)).Methods(http.MethodDelete)
	router.HandleFunc("/soft-delete-person/{id}", handlers.HandleSoftDeletePerson(client)).Methods(http.MethodDelete)

	fmt.Println("--------------------------------------------")

	// start listening
	http.ListenAndServe(":8080", router)
}
