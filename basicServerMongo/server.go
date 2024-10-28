package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempy"`
	Name string             `json:"name" bson:"name"`
	Age  string             `json:"age" bson:"age"`
}

var client *mongo.Client

func connectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Подключение к бд успешно")
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applicaion/json")
	var users []User
	collection := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Ошибка чтения из бд", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var user User
		err := cur.Decode(&user)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	if err := cur.Err(); err != nil {
		http.Error(w, "Ошибка чтения из бд", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Неверный id пользователя", http.StatusBadRequest)
		return
	}

	var user User
	colletion := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = colletion.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		http.Error(w, "Пользователь не найдем", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Неправильные данные", http.StatusBadRequest)
		return
	}

	collection := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newUser.ID = primitive.NewObjectID()
	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		http.Error(w, "Ошибка при добавлении пользователя", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newUser)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	var updatedUser User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Неправильные данные", http.StatusBadRequest)
		return
	}

	collection := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"name": updatedUser.Name,
			"age":  updatedUser.Age,
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Ошибка при обновлении данных", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь обновлен"})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	colletion := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = colletion.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Ошибка при удалении пользователя", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь удален"})
}

func main() {
	connectDB()
	r := mux.NewRouter()

	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	// users = append(users, User{ID: 1, Name: "Виктор", Age: "21"})
	// users = append(users, User{ID: 2, Name: "Аркадий", Age: "45"})

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
