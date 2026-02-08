package collegemanagementsystem

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// StudentHandler handles HTTP requests for students
// It contains MySQL , MongoDB and Redis connections
type StudentHandler struct {
	MySQLRepo StudentRepository //MySQL implementation
	MongoRepo StudentRepository //MongoDB implementation
	Redis     *redis.Client     //Redis client for caching
}

// LecturersHandler handles HTTP requests for lecturers
// It contains MySQL , MongoDB and Redis connections
type LecturerHandler struct {
	MySQLRepo LecturerRepository //MySQL implementation
	MongoRepo LecturerRepository //MongoDB implementation
	Redis     *redis.Client      //Redis client for caching
}

// LibraryHandler handles HTTP requests for library
// It contains MySQL , MongoDB and Redis connections
type LibraryHandler struct {
	MySQLRepo LibraryRepository //MySQL implementation
	MongoRepo LibraryRepository //MongoDB implementation
	Redis     *redis.Client     //Redis client for caching
}

// validateStudents validates incoming students data
func ValidateStudent(s Student) error {
	// validate name
	if s.Name == "" {
		return fmt.Errorf("name required")
	}
	// validate age
	if s.Age <= 0 {
		return fmt.Errorf("invalid age")
	}
	// validate email
	if s.Email == "" {
		return fmt.Errorf("email required")
	}
	// validate dept
	if s.Dept == "" {
		return fmt.Errorf("dept required")
	}

	return nil
}

// validateLecturer validates incoming lecturers data
func ValidateLecturer(l Lecturer) error {
	// validate name
	if l.Name == "" {
		return fmt.Errorf("name required")
	}
	// validate age
	if l.Age <= 0 {
		return fmt.Errorf("invalid age")
	}
	// validate email
	if l.Email == "" {
		return fmt.Errorf("email required")
	}
	// validate designation
	if l.Designation == "" {
		return fmt.Errorf("dept required")
	}

	return nil
}

// validateLecturer validates incoming library data
func ValidateLibrary(l Library) error {
	// validate name
	if l.Book_name == "" {
		return fmt.Errorf("name required")
	}
	// validate title
	if l.Title == "" {
		return fmt.Errorf("title required")
	}
	// validate author
	if l.Author == "" {
		return fmt.Errorf("author required")
	}
	// validate available_copies
	if l.Available_copies <= 0 {
		return fmt.Errorf("available_copies should not be less than 0")
	}

	return nil
}

// CreateStudent handles POST / students
// It creates a student in MySQL or MongoDB and clears Redis cache
func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {

	// Decode JSON requests body into student struct
	var s Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}

	// validate input data
	if err := ValidateStudent(s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Select database repository (MySQL/MongoDB)
	Repo := h.GetRepo(r)

	// Insert students into database
	data, err := Repo.CreateStudent(s)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Clear Redis cache after write operation
	h.Redis.Del(Ctx, "students:mysql")
	h.Redis.Del(Ctx, "students:mongo")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ReadAllStudents handles GET / students
// It first checks redis cache, if  not found , queries DB and store in redis
func (h *StudentHandler) GetAllStudent(w http.ResponseWriter, r *http.Request) {

	// Generate cache Key based on DB type
	key := "students:" + r.URL.Query().Get("db")

	// Try fetching from redis
	val, err := h.Redis.Get(Ctx, key).Result()
	if err == nil {
		log.Println("cache Hit...")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(val))
		return
	}

	// cache miss fetch from DB
	fmt.Println("cache miss querying DB")

	// Select database repository (MySQL/MongoDB)
	repo := h.GetRepo(r)

	// GET all students from database
	data, err := repo.GetAllStudent()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// converts data to json
	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store result in redis with TTL(Time to live)
	h.Redis.Set(Ctx, key, bytes, 60*time.Second)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ReadBYIDstudents handles GET /students /{id}
func (h *StudentHandler) GetByIDStudent(w http.ResponseWriter, r *http.Request) {

	// Read  ID from URL
	id, _ := mux.Vars(r)["id"]

	// Read data type
	db := r.URL.Query().Get("db")

	// Create cache key
	key := "student:" + db + ":" + id

	// Check redis
	val, err := h.Redis.Get(Ctx, key).Result()
	if err == nil {
		log.Println("cache Hit...")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(val))
		return
	}
	// cache miss querying DB
	fmt.Println("cache miss querying DB...")
	Repo := h.GetRepo(r)

	IDINT, _ := strconv.Atoi(id)

	// Get students by ID from DB
	data, err := Repo.GetByIDStudent(IDINT)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	// Convert data to json
	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store result in redis with TTL(Time to live)
	h.Redis.Set(Ctx, key, bytes, 60*time.Second)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// UpdateStudent handles PUT / students /{id}
func (h *StudentHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {

	// Decode JSON requests body into students struct
	var s Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	// validate input data
	if err := ValidateStudent(s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Select database repository (MySQL/MongoDB)
	Repo := h.GetRepo(r)

	// Update into students DB
	err := Repo.UpdateStudent(s)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// clear redis cache after write operations
	h.Redis.Del(Ctx, "students:mysql")
	h.Redis.Del(Ctx, "students:mongo")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

// DeleteStudent handles DELETE / students /{id}
func (h *StudentHandler) DeleteStudent(w http.ResponseWriter, r *http.Request) {

	// read ID from URL and convert it into integer
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	// Select database repository (MySQL/MongoDB)
	Repo := h.GetRepo(r)

	// DELETE from students DB
	err := Repo.DeleteStudent(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// clear redis cache after write operations
	h.Redis.Del(Ctx, "students:mysql")
	h.Redis.Del(Ctx, "students:mongo")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// CreateLecturers handles POST /lecturers
// It creates a lecturers in MySQL or MongoDB
// and clears Redis cache
func (h *LecturerHandler) CreateLecturer(w http.ResponseWriter, r *http.Request) {

	// Decode JSON request body into Lecturers struct
	var l Lecturer
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}

	// Validate input data
	if err := ValidateLecturer(l); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Select database repository (mysql/mongo)
	Repo := h.GetRepo(r)

	// Insert lecturer into database
	data, err := Repo.CreateLecturer(l)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Clear cache
	h.Redis.Del(Ctx, "lecturers:mysql")
	h.Redis.Del(Ctx, "lecturers:mongo")

	// Send respponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetAllLecturers handles GET /lecturers
// It first checks Redis cache,
// if not found, queries DB and stores in Redis
func (h *LecturerHandler) GetAllLecturer(w http.ResponseWriter, r *http.Request) {

	// Generate cache key based on DB type
	key := "lecturers:" + r.URL.Query().Get("id")

	// Try fetching from Redis
	val, err := h.Redis.Get(Ctx, key).Result()
	if err == nil {
		log.Println("cache Hit...")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(val))
		return
	}

	// Cache miss: fetch from DB
	fmt.Println("cache miss querying DB...")

	// Select database repository (mysql/mongo)
	Repo := h.GetRepo(r)

	// Getall lecturer from database
	data, err := Repo.GetAllLecturer()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Convert data to JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store result in Redis with TTL
	h.Redis.Set(Ctx, key, bytes, 60*time.Second)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetByIDLectrurer handles GET /lecturer/{id}
// It uses Redis cache for faster access
func (h *LecturerHandler) GetByIDLectuurer(w http.ResponseWriter, r *http.Request) {

	// Read id from URL
	id, _ := mux.Vars(r)["id"]

	// Read db type
	db := r.URL.Query().Get("db")

	// Create cache key
	key := "lecturer:" + db + ":" + id

	// Check Redis
	val, err := h.Redis.Get(Ctx, key).Result()
	if err == nil {
		log.Println("cache Hit...")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(val))
		return
	}

	// Cache miss → query database
	fmt.Println("cache miss querying DB...")

	// Select database repository (mysql/mongo)
	Repo := h.GetRepo(r)

	IDINT, _ := strconv.Atoi(id)

	// Get lecturer by ID from DB
	data, err := Repo.GetByIDLecturer(IDINT)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	// Convert data to JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store in Redis
	h.Redis.Set(Ctx, key, bytes, 60*time.Second)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// UpdateLecturer handles PUT / lecturers /{id}
func (h *LecturerHandler) UpdateLecturer(w http.ResponseWriter, r *http.Request) {

	// Decode JSOn requests body into Lecturers struct
	var l Lecturer
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	// validate input data
	if err := ValidateLecturer(l); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Select database repository (mysql/mongo)
	Repo := h.GetRepo(r)

	// update lecturer into database
	err := Repo.UpdateLecturer(l)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Clear cache
	h.Redis.Del(Ctx, "lecturers:mysql")
	h.Redis.Del(Ctx, "lecturers:mongo")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

// DeleteLecturer handles DELETE / lecturers /{id}
func (h *LecturerHandler) DeleteLecturer(w http.ResponseWriter, r *http.Request) {

	// read ID from URL and convert it into integer
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	// Select database repository (MySQL/MongoDB)
	Repo := h.GetRepo(r)

	// DELETE from lecturers DB
	err := Repo.DeleteLecturer(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Clear cache
	h.Redis.Del(Ctx, "lecturers:mysql")
	h.Redis.Del(Ctx, "lecturers:mongo")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// CreateLibrary handles POST / library
// It creates a library in MySQL or MongoDB and clears Redis cache
func (h *LibraryHandler) CreateLibrary(w http.ResponseWriter, r *http.Request) {

	// Decode JSON requests body into library struct
	var l Library
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}

	// validate input data
	if err := ValidateLibrary(l); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Select database repository (MySQL/MongoDB)
	Repo := h.GetRepo(r)

	// Insert library into database
	data, err := Repo.CreateLibrary(l)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Clear Redis cache after write operation
	db := r.URL.Query().Get("db")
	h.Redis.Del(Ctx, "libraries:"+db)
	h.Redis.Del(Ctx, "libraries:"+db+":"+strconv.Itoa(l.Book_id))

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ReadAllLibrary handles GET / Library
// It first checks redis cache, if  not found , queries DB and store in redis
func (h *LibraryHandler) GetAllLibrary(w http.ResponseWriter, r *http.Request) {

	// Generate cache Key based on DB type
	key := "libraries:" + r.URL.Query().Get("db")

	// Try fetching from redis
	val, err := h.Redis.Get(Ctx, key).Result()
	if err == nil {
		log.Println("cache Hit...")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(val))
		return
	}

	// cache miss fetch from DB
	fmt.Println("cache miss querying DB")

	// Select database repository (MySQL/MongoDB)
	repo := h.GetRepo(r)

	// GET all libraries from database
	data, err := repo.GetAllLibrary()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// converts data to json
	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store result in redis with TTL(Time to live)
	h.Redis.Set(Ctx, key, bytes, 60*time.Second)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ReadBYIDLibrary handles GET /libraries /{id}
func (h *LibraryHandler) GetByIDLibrary(w http.ResponseWriter, r *http.Request) {

	// Read  ID from URL
	id, _ := mux.Vars(r)["id"]

	// Read data type
	db := r.URL.Query().Get("db")

	// Create cache key
	key := "library:" + db + ":" + id

	// Check redis
	val, err := h.Redis.Get(Ctx, key).Result()
	if err == nil {
		log.Println("cache Hit...")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(val))
		return
	}
	// cache miss querying DB
	fmt.Println("cache miss querying DB...")
	Repo := h.GetRepo(r)

	IDINT, _ := strconv.Atoi(id)

	// Get libraries by ID from DB
	data, err := Repo.GetByIDLibrary(IDINT)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	// Convert data to json
	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store result in redis with TTL(Time to live)
	h.Redis.Set(Ctx, key, bytes, 60*time.Second)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// UpdateStudent handles PUT / students /{id}
func (h *LibraryHandler) UpdateLibrary(w http.ResponseWriter, r *http.Request) {

	// Decode JSON requests body into libraries struct
	var l Library
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	// validate input data
	if err := ValidateLibrary(l); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Select database repository (MySQL/MongoDB)
	Repo := h.GetRepo(r)

	// Update into students DB
	err := Repo.UpdateLibrary(l)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Clear Redis cache after write operation
	db := r.URL.Query().Get("db")
	h.Redis.Del(Ctx, "libraries:"+db)
	h.Redis.Del(Ctx, "libraries:"+db+":"+strconv.Itoa(l.Book_id))

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

// DeleteLibraries handles DELETE / libraries /{id}
func (h *LibraryHandler) DeleteLibrary(w http.ResponseWriter, r *http.Request) {

	// read ID from URL and convert it into integer
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	// Select database repository (MySQL/MongoDB)
	Repo := h.GetRepo(r)

	// DELETE from libraries DB
	err := Repo.DeleteLibrary(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Clear Redis cache after write operation
	var l Library
	db := r.URL.Query().Get("db")
	h.Redis.Del(Ctx, "libraries:"+db)
	h.Redis.Del(Ctx, "libraries:"+db+":"+strconv.Itoa(l.Book_id))

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}
