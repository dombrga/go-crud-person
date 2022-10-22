package main

import (
	"context"
	"fmt"

	// "log"
	"net/http"
	// "os"
	"time"

	"github.com/gorilla/mux"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"practice.com/mongodb-crud/pkg/constants"
	"practice.com/mongodb-crud/pkg/handlers"
)

var client *mongo.Client

func main() {
	fmt.Println("Starting backend app...")
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect to mongo
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI(constants.MONGO_URI))

	var router = mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Default path accessed.")
	}).Methods("GET")

	// routers
	router.HandleFunc("/persons", handlers.HandleGetPersons(client)).Methods(http.MethodGet)
	router.HandleFunc("/person/{id}", handlers.HandleGetPerson(client)).Methods(http.MethodGet)
	router.HandleFunc("/person", handlers.HandleCreatePerson(client)).Methods(http.MethodPost)
	router.HandleFunc("/person/{id}", handlers.HandleUpdatePerson(client)).Methods(http.MethodPut)
	router.HandleFunc("/hard-delete-person/{id}", handlers.HandleHardDeletePerson(client)).Methods(http.MethodDelete)
	router.HandleFunc("/soft-delete-person/{id}", handlers.HandleSoftDeletePerson(client)).Methods(http.MethodDelete)

	fmt.Println("--------------------------------------------")

	// listen and serve
	http.ListenAndServe(":8080", router)
}
