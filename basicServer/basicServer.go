package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"math/rand"

	"github.com/gorilla/mux"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  string `json:"age"`
}

var users []User
var mu sync.Mutex

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applicaion/json")
	mu.Lock()
	defer mu.Unlock()
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

	mu.Lock()
	defer mu.Unlock()
	for _, user := range users {
		if user.ID == id {
			json.NewEncoder(w).Encode(user)
			return
		}
	}
	http.Error(w, "Пользователь не найден", http.StatusNotFound)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Неправильные данные", http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	newUser.ID = rand.Intn(100)
	users = append(users, newUser)
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

	mu.Lock()
	defer mu.Unlock()
	for i, user := range users {
		if user.ID == id {
			users[i].Name = updatedUser.Name
			users[i].Age = user.Age
			json.NewEncoder(w).Encode(users[i])
			return
		}
	}
	http.Error(w, "Пользователь не найден", http.StatusNotFound)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Неправильный ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			json.NewEncoder(w).Encode(map[string]string{"message": "пользователь удален"})
			return
		}
	}
	http.Error(w, "Пользователь не найден", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	users = append(users, User{ID: 1, Name: "Виктор", Age: "21"})
	users = append(users, User{ID: 2, Name: "Аркадий", Age: "45"})

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
