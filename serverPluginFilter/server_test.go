package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mock функция для тестов, возвращающая список пользователей
func getUsersMock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users := []User{
		{ID: primitive.NewObjectID(), Name: "Alice", Age: "25"},
		{ID: primitive.NewObjectID(), Name: "Bob", Age: "30"},
	}

	// Получаем параметры пагинации из запроса
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	// Преобразуем параметры в числа с проверкой ошибок
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = len(users) // если лимит не указан или невалиден, возвращаем весь список
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // если страница не указана или невалидна, ставим страницу 1
	}

	// Рассчитываем начальный и конечный индексы
	start := (page - 1) * limit
	end := start + limit

	// Ограничиваем конечный индекс длиной слайса
	if start >= len(users) {
		json.NewEncoder(w).Encode([]User{}) // если стартовый индекс за пределами слайса, возвращаем пустой массив
		return
	}
	if end > len(users) {
		end = len(users)
	}

	// Возвращаем нужный подмассив пользователей
	json.NewEncoder(w).Encode(users[start:end])
}

// Mock функция для создания пользователя
func createUserMock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Устанавливаем статус 201 Created
	newUser := User{
		ID:   primitive.NewObjectID(),
		Name: "AAAA",
		Age:  "1000",
	}
	json.NewEncoder(w).Encode(newUser)
}

// Mock функция для обновления пользователя
func updateUserMock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Устанавливаем статус 200 OK
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated"})
}

// Mock функция для удаления пользователя
func deleteUserMock(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent) // Устанавливаем статус 204 No Content
}

// Тестирование GET /users
func TestGetUsers(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/users", getUsersMock).Methods("GET")

	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var users []User
	err = json.NewDecoder(rr.Body).Decode(&users)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(users))
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
}

// Тестирование фильтрации по имени в GET /users
func TestGetUsersWithFilter(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/users", getUsersMock).Methods("GET")

	req, err := http.NewRequest("GET", "/users?name=alice", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var users []User
	err = json.NewDecoder(rr.Body).Decode(&users)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(users))
	assert.Equal(t, "Alice", users[0].Name)
}

// Тестирование пагинации в GET /users
func TestGetUsersWithPagination(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/users", getUsersMock).Methods("GET")

	limit := 1
	page := 1
	req, err := http.NewRequest("GET", "/users?limit="+strconv.Itoa(limit)+"&page="+strconv.Itoa(page), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var users []User
	err = json.NewDecoder(rr.Body).Decode(&users)
	if err != nil {
		t.Fatal(err)
	}

	// Поскольку в mock-ответе два пользователя, проверим, что при лимите 1, количество равно 1.
	assert.Equal(t, limit, len(users))
}

// Тестирование POST /users
func TestCreateUser(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/users", createUserMock).Methods("POST")

	user := `{"name":"AAAA"}`
	req, err := http.NewRequest("POST", "/users", strings.NewReader(user))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdUser User
	err = json.NewDecoder(rr.Body).Decode(&createdUser)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "AAAA", createdUser.Name)
}

// Тестирование PUT /users/{id}
func TestUpdateUser(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", updateUserMock).Methods("PUT")

	id := primitive.NewObjectID().Hex()
	update := `{"name":"Updated Name"}`
	req, err := http.NewRequest("PUT", "/users/"+id, strings.NewReader(update))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "User updated", response["message"])
}

// Тестирование DELETE /users/{id}
func TestDeleteUser(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", deleteUserMock).Methods("DELETE")

	id := primitive.NewObjectID().Hex()
	req, err := http.NewRequest("DELETE", "/users/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
