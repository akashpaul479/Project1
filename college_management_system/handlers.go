package collegemanagementsystem

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	// Name validation
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("Empty name or invalid name")
	}
	// validate age
	if s.Age <= 0 {
		return fmt.Errorf("invalid age")
	}
	// Email validation
	if strings.TrimSpace(s.Email) == "" {
		return fmt.Errorf("Empty email or invalid email")
	}
	if !strings.HasSuffix(s.Email, "@gmail.com") {
		return fmt.Errorf("email is invalid and does not contains @gmail.com")
	}
	prefix := strings.TrimSuffix(s.Email, "@gmail.com")
	if prefix == "" {
		return fmt.Errorf("email must contains a prefix before @gmail.com")
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

// CreateStudent godoc
// @Summary Create a new student
// @Description Add a student to MySQL or MongoDB
// @Tags Students
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param db query string false "Database type (mysql/mongo)"
// @Param student body Student true "Student Data"
// @Success 200 {object} Student
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /api/students [post]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Lod activity and Audit trail
	go LogActivity("CREATE_STUDENT", actor)
	go AuditLog("CREATE", "STUDENT", data.ID, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetAllStudent godoc
// @Summary Get all students
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param db query string false "Database type (mysql/mongo)"
// @Success 200 {array} Student
// @Router /api/students [get]
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

// GetByIDStudent godoc
// @Summary Get student by ID
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param id path int true "Student ID"
// @Param db query string false "Database type"
// @Success 200 {object} Student
// @Failure 400 {string} string
// @Router /api/students/{id} [get]
// ReadBYIDstudents handles GET /students /{id}
func (h *StudentHandler) GetByIDStudent(w http.ResponseWriter, r *http.Request) {

	// Read  ID from URL
	id, _ := mux.Vars(r)["id"]

	// Read data type
	db := r.URL.Query().Get("db")

	// Create cache key
	key := "student:" + db + ":" + id

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log Get activity
	go LogActivity("GET_STUDENT", actor)

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

// UpdateeStudent godoc
// @Summary Update student
// @Tags Students
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param student body Student true "Updated Student"
// @Success 200 {object} Student
// @Router /api/students/{id} [put]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Lod activity and Audit trail
	go LogActivity("UPDATE_STUDENT", actor)
	go AuditLog("UPDATE", "STUDENT", s.ID, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

// DeleteStudent godoc
// @Summary Delete student
// @Tags Students
// @Security BearerAuth
// @Param id path int true "Student ID"
// @Success 200 {object} map[string]string
// @Router /api/students/{id} [delete]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log delete response
	go LogActivity("DELETE_STUDENTS", actor)
	go AuditLog("DELETE", "STUDENT", id, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// CreateLecturer godoc
// @Summary Create a new lecturer
// @Description Add a lecturer to MySQL or MongoDB
// @Tags Lecturers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param db query string false "Database type (mysql/mongo)"
// @Param lecturer body Lecturer true "Lecturer Data"
// @Success 200 {object} Lecturer
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /api/lecturers [post]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("CREATE_LECTURER", actor)
	go AuditLog("CREATE", "LECTURER", data.ID, actor)

	// Send respponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetAllLecturer godoc
// @Summary Get all lecturers
// @Tags Lecturers
// @Security BearerAuth
// @Produce json
// @Param db query string false "Database type (mysql/mongo)"
// @Success 200 {array} Lecturer
// @Router /api/lecturers [get]
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

// GetByIDLecturer godoc
// @Summary Get lecturer by ID
// @Tags Lecturers
// @Security BearerAuth
// @Produce json
// @Param id path int true "Lecturer ID"
// @Param db query string false "Database type"
// @success 200 {object} Lecturer
// @Failure 400 {string} string
// @Router /api/lecturers/{id} [get]
// GetByIDLectrurer handles GET /lecturer/{id}
// It uses Redis cache for faster access
func (h *LecturerHandler) GetByIDLectuurer(w http.ResponseWriter, r *http.Request) {

	// Read id from URL
	id, _ := mux.Vars(r)["id"]

	// Read db type
	db := r.URL.Query().Get("db")

	// Create cache key
	key := "lecturer:" + db + ":" + id

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log GET By ID response
	go LogActivity("GET_LECTURER", actor)

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

// UpdateLecturer godoc
// @Summary Update lecturer
// @Tags Lecturers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param lecturer body Lecturer true "Updated Lecturer"
// @Success 200 {object} Lecturer
// @Router /api/lecturers/{id} [put]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("UPDATE_LECTURER", actor)
	go AuditLog("UPDATE", "LECTURER", l.ID, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

// DeleteLecturer godoc
// @Summary Delete lecturer
// @Tags Lecturers
// @Security BearerAuth
// @Param id path int true "Lecturer ID"
// @Success 200 {object} map[string]string
// @Router /api/lecturers/{id} [delete]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("DELETE_LECTURER", actor)
	go AuditLog("DELETE", "LECTURER", id, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// CreateLibrary godoc
// @Summary Create a new librarybook
// @Description Add a librarybook to MySQL or MongoDB
// @Tags Library
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param db query string false "Database type (mysql/mongo)"
// @Param library body Library true "Library Data"
// @Success 200 {object} Library
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /api/libraries [post]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("CREATE_LIBRARY", actor)
	go AuditLog("CREATE", "LIBRARY", data.Book_id, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetAllLibrary godoc
// @Summary Get all librarybook
// @Tags Library
// @Security BearerAuth
// @Produce json
// @Param db query string false "Database type (mysql/mongo)"
// @Success 200 {array} Library
// @Router /api/libraries [get]
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

// GetByIDLibrary godoc
// @Summary Get librarybook by ID
// @Tags Library
// @Security BearerAuth
// @Produce json
// @Param id path int true "Library ID"
// @Param db query string false "Database type"
// @success 200 {object} Library
// @Failure 400 {string} string
// @Router /api/libraries/{id} [get]
// ReadBYIDLibrary handles GET /libraries /{id}
func (h *LibraryHandler) GetByIDLibrary(w http.ResponseWriter, r *http.Request) {

	// Read  ID from URL
	id, _ := mux.Vars(r)["id"]

	// Read data type
	db := r.URL.Query().Get("db")

	// Create cache key
	key := "library:" + db + ":" + id

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}
	// Log activity and Audit trail
	go LogActivity("GET_LIBRARY", actor)

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

// UpdateLibrary godoc
// @Summary Update librarybook
// @Tags Library
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param library body Library true "Updated Library"
// @Success 200 {object} Library
// @Router /api/libraries/{id} [put]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("UPDATE_LIBRARY", actor)
	go AuditLog("UPDATE", "LIBRARY", l.Book_id, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
}

// DeleteLibrary godoc
// @Summary Delete librarybook
// @Tags Library
// @Security BearerAuth
// @Param id path int true "Library ID"
// @Success 200 {object} map[string]string
// @Router /api/libraries/{id} [delete]
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

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("DELETE_LIBRARY", actor)
	go AuditLog("DELETE", "LIBRARY", l.Book_id, actor)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// BorrowBookHandler godoc
// @Summary Borrow a book
// @Tags Borrow_Return
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param db query string true "Database type: mysql or mongo"
// @Param borrow body BorrowInfo true "Borrow Info"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string
// @Router /api/borrow [post]
// BorrowBookHandler handles borrowing a book request.
func (h *LibraryHandler) BorrowBookHandler(w http.ResponseWriter, r *http.Request) {

	// Set response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Check which database to use (mysql or mongo)
	db := r.URL.Query().Get("db")
	if db != "mysql" && db != "mongo" {
		http.Error(w, "Please specify ?db=mysql or ?db=mongo", http.StatusBadRequest)
		return
	}

	// Get repository implementation based on request (MySQL or Mongo)
	repo := h.GetRepo(r)

	var info BorrowInfo

	// Decode JSON request body into BorrowInfo struct
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if info.BookID == 0 || info.UserID == 0 {
		http.Error(w, "book_id and user_id required", http.StatusBadRequest)
		return
	}

	// Call repository method to borrow the book
	err = repo.BorrowBook(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("BORROW_RECORDS", actor)
	go AuditLog("BORROW", "RECORDS", info.BookID, actor)

	// Send response
	json.NewEncoder(w).Encode(map[string]string{"message": "Book borrowed succesfully"})

}

// ReturnBookHandler godoc
// @Summary Return a book
// @Tags Borrow_Return
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param db query string true "Database type: mysql or mongo"
// @Param return body BorrowInfo true "Return Info"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /api/return [post]
func (h *LibraryHandler) ReturnBookHandler(w http.ResponseWriter, r *http.Request) {

	// Set response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Get repository implementation (MySQL or Mongo)
	repo := h.GetRepo(r)

	var info BorrowInfo

	// Decode JSON request body into BorrowInfo struct
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if info.BookID == 0 || info.UserID == 0 {
		http.Error(w, "book_id and user_id required", http.StatusBadRequest)
		return
	}

	// Call repository method to return the book
	err = repo.ReturnBook(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// ✅ Get logged-in user
	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "unknown"
	}

	// Log activity and Audit trail
	go LogActivity("RETURN_RECORDS", actor)
	go AuditLog("RETURN", "RECORDS", info.BookID, actor)
	// Send response
	json.NewEncoder(w).Encode(map[string]string{"message": "Book Returned successfully!"})
}
