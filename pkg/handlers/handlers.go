package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"practice.com/mongodb-crud/pkg/constants"
	"practice.com/mongodb-crud/pkg/helpers"
	"practice.com/mongodb-crud/pkg/models"
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
			// filter = bson.M{"$or": []interface{}{
			// 	bson.M{"isSoftDeleted": true},
			// 	bson.M{"isSoftDeleted": false},
			// }}
		}

		// container of persons
		var persons []models.Person

		// db
		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// find all persons
		var cursor, err = collection.Find(ctx, filter)

		if err != nil {
			log.Fatal(err)
		}

		// decode
		if err = cursor.All(context.TODO(), &persons); err != nil {
			log.Fatal(err)
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
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// find
		err = collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&person)
		if err != nil {
			// if did not match any document
			if err == mongo.ErrNoDocuments {
				json.NewEncoder(res).Encode("No person found by this ID.")
				return
			}
			log.Fatal(err)
		}

		json.NewEncoder(res).Encode(person)
	}
}

var validate = validator.New()

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
		var bodyErr = validate.Struct(person)
		if bodyErr != nil {
			var validationErrors = bodyErr.(validator.ValidationErrors)
			var errorTranslated = helpers.TranslateErrors(validationErrors)
			json.NewEncoder(res).Encode(errorTranslated)
			return
		}

		// db
		var collection = client.Database(constants.DATABASE).Collection(constants.PERSONS_COLLECTION)
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// insert new person to db
		var insertResult, err = collection.InsertOne(ctx, person)
		fmt.Println("insert result:", insertResult)
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
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// for update
		var update = bson.M{
			"$set": bson.M{
				"firstName": person.FirstName,
				"lastName":  person.LastName,
				"birthdate": person.Birthdate,
			}}
		fmt.Println("update:", update)

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
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
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
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
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
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
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
