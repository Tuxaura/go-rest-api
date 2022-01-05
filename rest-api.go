package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}



func CreatePersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var person Person
	_ = json.NewDecoder(request.Body).Decode(&person)
	collection := client.Database("mydatabase").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(response).Encode(result)
}

func GetPersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	collection := client.Database("mydatabase").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(person)
	fmt.Println("Endpoint Hit: get person, Time: ", time.Now())
}

func GetPeopleEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var people []Person
	collection := client.Database("mydatabase").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(people)
	fmt.Println("Endpoint Hit: get people, Time: ", time.Now())
}

func UpdatePerson(w http.ResponseWriter, r *http.Request){
    params := mux.Vars(r)
    id := params["id"]

	reqBody, _ := ioutil.ReadAll(r.Body)
    var d Person 
	json.Unmarshal(reqBody, &d)
    

	collection := client.Database("mydatabase").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

    data_update := bson.M{
		"$set": d,
    }

    objID, _ := primitive.ObjectIDFromHex(id)
    response, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, data_update)
    if err != nil {
        log.Fatal(err.Error())
    }

    json.NewEncoder(w).Encode(response)
	
	fmt.Println("Endpoint Hit: update person, Time: ", time.Now())
}

func DeletePerson(w http.ResponseWriter, r *http.Request){
    params := mux.Vars(r)
    id := params["id"]

	collection := client.Database("mydatabase").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

    objID, _ := primitive.ObjectIDFromHex(id)
    response, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        log.Fatal(err.Error())
    }

    json.NewEncoder(w).Encode(response)
	
	fmt.Println("Endpoint Hit: delete person, Time: ", time.Now())
}



func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", GetPersonEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", UpdatePerson).Methods("PATCH")
	router.HandleFunc("/person/{id}", DeletePerson).Methods("DELETE")
	http.ListenAndServe(":12345", router)
}

