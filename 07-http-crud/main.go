package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Email string `json:"email"`
}

type userRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Email string `json:"email"`
}

type Store struct {
      users  []User
      nextID int
  }

func newStore() *Store {
	return &Store{nextID: 1}
}

func main() {
	store := newStore()
	http.HandleFunc("/users", store.usersHandler)
	http.HandleFunc("/users/", store.userByIDHandler)

	fmt.Println("server started: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("server error:", err)
	}
}

func (s *Store) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getUsers(w, r)
	case http.MethodPost:
		s.createUser(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Store) userByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getUser(w, id)
	case http.MethodPut:
		s.updateUser(w, r, id)
	case http.MethodDelete:
		s.deleteUser(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Store) getUsers(w http.ResponseWriter, r *http.Request) {
	minAgeStr := r.URL.Query().Get("minAge")
	if minAgeStr == "" {
		writeJSON(w, http.StatusOK, s.users)
		return
	}

	minAge, err := strconv.Atoi(minAgeStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid minAge")
		return
	}

	var result []User
	for _, user := range s.users {
		if user.Age >= minAge {               
			result = append(result, user)
		}
	}

    writeJSON(w, http.StatusOK, result)

}

func (s *Store) getUser(w http.ResponseWriter, id int) {
	for _, user := range s.users {
		if user.ID == id {
			writeJSON(w, http.StatusOK, user)
			return
		}
	}

	writeError(w, http.StatusNotFound, "user not found")
}

func (s *Store) createUser(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeUserRequest(w, r)
	if !ok {
		return
	}

	user := User{
		ID:    s.nextID,
		Name:  req.Name,
		Age:   req.Age,
		Email: req.Email,
	}
	s.nextID++
	s.users = append(s.users, user)

	writeJSON(w, http.StatusCreated, user)
}

func (s *Store) updateUser(w http.ResponseWriter, r *http.Request, id int) {
	req, ok := decodeUserRequest(w, r)
	if !ok {
		return
	}

	for i := range s.users {
		if s.users[i].ID == id {
			s.users[i].Name = req.Name
			s.users[i].Age = req.Age
			s.users[i].Email = req.Email
			writeJSON(w, http.StatusOK, s.users[i])
			return
		}
	}

	writeError(w, http.StatusNotFound, "user not found")
}

func (s *Store) deleteUser(w http.ResponseWriter, id int) {
	for i := range s.users {
		if s.users[i].ID == id {
			deleted := s.users[i]
			s.users = append(s.users[:i], s.users[i+1:]...)
			writeJSON(w, http.StatusOK, deleted)
			return
		}
	}

	writeError(w, http.StatusNotFound, "user not found")
}

func decodeUserRequest(w http.ResponseWriter, r *http.Request) (userRequest, bool) {
	defer r.Body.Close()

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return userRequest{}, false
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return userRequest{}, false
	}

	if req.Age < 0 {
		writeError(w, http.StatusBadRequest, "age must be greater than or equal to 0")
		return userRequest{}, false
	}

	return req, true
}

func idFromPath(path string) (int, bool) {
	idText := strings.TrimPrefix(path, "/users/")
	if idText == "" || strings.Contains(idText, "/") {
		return 0, false
	}

	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
