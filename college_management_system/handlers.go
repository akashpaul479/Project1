package collegemanagementsystem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type StudentHandler struct {
	MySQLRepo StudentRepository
	MongoRepo StudentRepository
}
type LecturerHandler struct {
	MySQLRepo LecturerRepository
	MongoRepo LecturerRepository
}

func ValidateStudent(s Student) error {

	if s.Name == "" {
		return fmt.Errorf("name required")
	}
	if s.Age <= 0 {
		return fmt.Errorf("invalid age")
	}
	if s.Email == "" {
		return fmt.Errorf("email required")
	}
	if s.Dept == "" {
		return fmt.Errorf("dept required")
	}

	return nil
}

// validate lecturer
func ValidateLecturer(l Lecturer) error {

	if l.Name == "" {
		return fmt.Errorf("name required")
	}
	if l.Age <= 0 {
		return fmt.Errorf("invalid age")
	}
	if l.Email == "" {
		return fmt.Errorf("email required")
	}
	if l.Designation == "" {
		return fmt.Errorf("dept required")
	}

	return nil
}

// CREATE student
func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {

	var s Student

	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}
	if err := ValidateStudent(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Repo := h.GetRepo(r)
	data, err := Repo.CreateStudent(s)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// READ ALL students
func (h *StudentHandler) GetAllStudent(w http.ResponseWriter, r *http.Request) {

	Repo := h.GetRepo(r)
	data, err := Repo.GetAllStudent()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// READ BY ID students
func (h *StudentHandler) GetByIDStudent(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	Repo := h.GetRepo(r)
	data, err := Repo.GetByIDStudent(id)

	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// UPDATE student
func (h *StudentHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {

	var s Student

	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if err := ValidateStudent(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Repo := h.GetRepo(r)
	err := Repo.UpdateStudent(s)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

// DELETE Student
func (h *StudentHandler) DeleteStudent(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	Repo := h.GetRepo(r)
	err := Repo.DeleteStudent(id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

/* CREATE */
func (h *LecturerHandler) CreateLecturer(w http.ResponseWriter, r *http.Request) {

	var l Lecturer

	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}
	if err := ValidateLecturer(l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Repo := h.GetRepo(r)
	data, err := Repo.CreateLecturer(l)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(data)
}

/* READ ALL */
func (h *LecturerHandler) GetAllLecturer(w http.ResponseWriter, r *http.Request) {

	Repo := h.GetRepo(r)
	data, err := Repo.GetAllLecturer()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// READ BY ID lecturer
func (h *LecturerHandler) GetByIDLectuurer(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	Repo := h.GetRepo(r)
	data, err := Repo.GetByIDLecturer(id)

	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// UPDATE lecturer
func (h *LecturerHandler) UpdateLecturer(w http.ResponseWriter, r *http.Request) {

	var l Lecturer

	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if err := ValidateLecturer(l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Repo := h.GetRepo(r)
	err := Repo.UpdateLecturer(l)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

/* DELETE */
func (h *LecturerHandler) DeleteLecturer(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	Repo := h.GetRepo(r)
	err := Repo.DeleteLecturer(id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}
