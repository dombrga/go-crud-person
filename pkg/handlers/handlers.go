package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dombrga/go-crud-person/pkg/constants"
	"github.com/dombrga/go-crud-person/pkg/helpers"
	"github.com/dombrga/go-crud-person/pkg/models"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleGetPersons(client *mongo.Client) http.HandlerFunc {

	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("Getting...")

		res.Header().Add("Content-Type", "application/json")

		// filter
		var filter = bson.M{"isSoftDeleted": false}
		// get params
		var isIncludeDeleted, boolErr = strconv.ParseBool(req.URL.Query().Get("includeDeleted"))
		if boolErr != nil {
			panic(boolErr)
		}

		if isIncludeDeleted {
			filter = bson.M{}
		}

		// container of persons
		var persons []models.Person

		// db
		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = helpers.CreateContext()
		defer cancel()

		// find all persons
		var cursor, err = collection.Find(ctx, filter)

		if err != nil {
			panic(err)
		}

		// decode
		if err = cursor.All(context.TODO(), &persons); err != nil {
			panic(err)
		}

		// send result
		json.NewEncoder(res).Encode(persons)
	}
}

func HandleGetPerson(client *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("Getting...")

		res.Header().Set("Content-Type", "application/json")

		// person container
		var person models.Person

		// id of person to get
		var params = mux.Vars(req)
		var objId, err = primitive.ObjectIDFromHex(params["id"])
		if err != nil {
			json.NewEncoder(res).Encode("ID given is not valid.")
			return
		}

		// db
		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = helpers.CreateContext()
		defer cancel()

		// find
		err = collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&person)
		if err != nil {
			// if did not match any document
			if err == mongo.ErrNoDocuments {
				json.NewEncoder(res).Encode("No person found by this ID.")
				return
			}
			panic(err)
		}

		json.NewEncoder(res).Encode(person)
	}
}

func HandleCreatePerson(client *mongo.Client) http.HandlerFunc {

	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("Creating...")

		res.Header().Add("Content-Type", "application/json")

		// check if body is empty
		if req.Body == http.NoBody {
			fmt.Println("Request body is empty")
			json.NewEncoder(res).Encode("Request body is empty.")
			return
		}

		// container of new person
		var person models.PersonRequest
		// decode request body
		var _ = json.NewDecoder(req.Body).Decode(&person)
		person.IsSoftDeleted = false

		// validate body
		var structErr = helpers.ValidateBody(person)
		if structErr != nil {
			var validationErrors = structErr.(validator.ValidationErrors)
			var errorTranslated = helpers.TranslateErrors(validationErrors)
			json.NewEncoder(res).Encode(errorTranslated)
			return
		}

		// db
		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// insert new person to db
		var _, err = collection.InsertOne(ctx, person)
		if err != nil {
			panic(err)
		}

		// get all persons to send back
		var allPersons = GetPersons(client)

		// send result
		res.WriteHeader(http.StatusCreated)
		json.NewEncoder(res).Encode(allPersons)
	}
}

func HandleUpdatePerson(client *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("Updating...")

		res.Header().Add("Content-Type", "application/json")

		// get person id
		var params = mux.Vars(req)
		var id, err = primitive.ObjectIDFromHex(params["id"])
		if err != nil {
			json.NewEncoder(res).Encode("ID given is not valid.")
			return
		}

		var person models.Person
		// decode body
		json.NewDecoder(req.Body).Decode(&person)

		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = helpers.CreateContext()
		defer cancel()

		// for update
		var update = bson.M{
			"$set": bson.M{
				"firstName": person.FirstName,
				"lastName":  person.LastName,
				"birthdate": person.Birthdate,
			}}

		// call update
		var updateResult, updateErr = collection.UpdateByID(ctx, id, update)
		if updateErr != nil {
			panic(updateErr)
		}
		if updateResult.ModifiedCount == 0 {
			json.NewEncoder(res).Encode("No person found by given id.")
			return
		}

		// get all persons to send back
		var allPersons = GetPersons(client)

		// send
		json.NewEncoder(res).Encode(allPersons)
	}
}

func HandleSoftDeletePerson(client *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("Deleting...")

		res.Header().Add("Content-Type", "application/json")

		// params
		var params = mux.Vars(req)
		var objId, err = primitive.ObjectIDFromHex(params["id"])
		if err != nil {
			json.NewEncoder(res).Encode("ID given is not valid.")
			return
		}

		// db
		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = helpers.CreateContext()
		defer cancel()

		// filter
		var filter = bson.M{"_id": objId}
		var update = bson.M{"$set": bson.M{"isSoftDeleted": true}}
		// update
		var updateResult bson.M
		var updateErr = collection.FindOneAndUpdate(ctx, filter, update).Decode(&updateResult)
		if updateErr != nil {
			if updateErr == mongo.ErrNoDocuments {
				json.NewEncoder(res).Encode("No person found by given id.")
				return
			}
			panic(updateErr)
		}

		// get all persons to send back
		var allPersons = GetPersons(client)

		// send
		json.NewEncoder(res).Encode(allPersons)
	}
}

func HandleHardDeletePerson(client *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("Deleting...")

		res.Header().Add("Content-Type", "application/json")

		// params
		var params = mux.Vars(req)
		var id, err = primitive.ObjectIDFromHex(params["id"])
		if err != nil {
			json.NewEncoder(res).Encode("ID given is not valid.")
			return
		}

		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = helpers.CreateContext()
		defer cancel()

		// filter
		var filter = bson.M{"_id": id}

		// delete
		var deleteResult, delErr = collection.DeleteOne(ctx, filter)
		if delErr != nil {
			panic(delErr)
		}
		if deleteResult.DeletedCount == 0 {
			json.NewEncoder(res).Encode("No person found by given id.")
			return
		}

		// get all persons to send back
		var allPersons = GetPersons(client)

		json.NewEncoder(res).Encode(allPersons)
	}
}

func GetPersons(client *mongo.Client) []models.Person {
	var persons []models.Person

	var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
	var ctx, cancel = helpers.CreateContext()
	defer cancel()

	// filter
	var filter = bson.M{"isSoftDeleted": false}

	// find
	var cursor, err = collection.Find(ctx, filter)
	if err != nil {
		panic(err)
	}

	var cursorErr = cursor.All(context.TODO(), &persons)
	if cursorErr != nil {
		panic(err)
	}

	return persons
}
