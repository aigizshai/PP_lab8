package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

func handleError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func validateUser(user User) (bool, string) {
	if strings.TrimSpace(user.Name) == "" {
		return false, "Имя не может быть пустым"
	}
	return true, ""
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applicaion/json")

	name := r.URL.Query().Get("name")
	minAgeParam := r.URL.Query().Get("min_age")
	maxAgeParam := r.URL.Query().Get("max_age")
	limitParam := r.URL.Query().Get("limit")
	pageParam := r.URL.Query().Get("page")

	var minAge, maxAge, limit, page int
	var err error

	if limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
		if err != nil || limit <= 0 {
			handleError(w, "Неверное значение limit", http.StatusBadRequest)
			return
		}
	} else {
		limit = 10
	}

	if pageParam != "" {
		page, err := strconv.Atoi(pageParam)
		if err != nil || page <= 0 {
			handleError(w, "Неверное значение page", http.StatusBadRequest)
			return
		}
	} else {
		page = 1
	}

	if minAgeParam != "" {
		minAge, err = strconv.Atoi(minAgeParam)
		if err != nil {
			handleError(w, "Неверное значение min_age", http.StatusBadRequest)
			return
		}
	}

	if maxAgeParam != "" {
		maxAge, err = strconv.Atoi(maxAgeParam)
		if err != nil {
			handleError(w, "Неверное значение max_age", http.StatusBadRequest)
			return
		}
	}

	//смещение
	skip := (page - 1) * limit

	var users []User
	collection := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if name != "" {
		filter["name"] = bson.M{"$regex": name, "$options": "i"}
	}

	if minAge > 0 || maxAge > 0 {
		ageFilter := bson.M{}
		if minAge > 0 {
			ageFilter["$gte"] = minAge
		}
		if maxAge > 0 {
			ageFilter["$lte"] = maxAge
		}
		filter["age"] = ageFilter
	}

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	cur, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		handleError(w, "Ошибка чтения из бд", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	// totalUsers, err := collection.CountDocuments(ctx, filter)
	// if err != nil {
	// 	handleError(w, "Ошибка при подсчете данных", http.StatusInternalServerError)
	// 	return
	// }

	for cur.Next(ctx) {
		var user User
		err := cur.Decode(&user)
		if err != nil {
			handleError(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := cur.Err(); err != nil {
		handleError(w, "Ошибка чтения из бд", http.StatusInternalServerError)
		return
	}
	//fmt.Println(totalUsers)
	//totalPages := (int(totalUsers) + limit - 1) / limit

	// json.NewEncoder(w).Encode(map[string]interface{}{
	// 	"users":        users,
	// 	"total_pages":  totalPages,
	// 	"total_users":  totalUsers,
	// 	"current_page": page,
	// })

	json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	var user User
	colletion := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := colletion.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		handleError(w, "Пользователь не найдем", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		handleError(w, "Неправильные данные", http.StatusBadRequest)
		return
	}

	if valid, message := validateUser(newUser); !valid {
		handleError(w, message, http.StatusBadRequest)
		return
	}

	collection := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newUser.ID = primitive.NewObjectID()
	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		handleError(w, "Ошибка при добавлении пользователя", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newUser)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		handleError(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	var updatedUser User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		handleError(w, "Неправильные данные", http.StatusBadRequest)
		return
	}

	if valid, message := validateUser(updatedUser); !valid {
		handleError(w, message, http.StatusBadRequest)
		return
	}

	collection := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectId}
	update := bson.M{
		"$set": bson.M{
			"name": updatedUser.Name,
			"age":  updatedUser.Age,
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		handleError(w, "Ошибка при обновлении данных", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь обновлен"})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		handleError(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	collection := client.Database("lab8").Collection("test")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		handleError(w, "Ошибка при удалении пользователя", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		handleError(w, "Пользователь не найден", http.StatusNotFound)
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

	//GET http://localhost:8080/users?name=alice&limit=5&page=2
	// users = append(users, User{ID: 1, Name: "Виктор", Age: "21"})
	// users = append(users, User{ID: 2, Name: "Аркадий", Age: "45"})

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
