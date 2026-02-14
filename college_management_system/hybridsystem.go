package collegemanagementsystem

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "college_management_system/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// MySQLStudentRepo implements StudentRepository using MySQL
type MySQLStudentRepo struct {
	DB *sql.DB
}

// MongoDBStudentRepo implements StudentRepository using MongoDB
type MongoDBStudentRepo struct {
	Collection *mongo.Collection
}

// MySQLLecturerRepo implements LecturerRepository using MySQL
type MySQLLecturerRepo struct {
	DB *sql.DB
}

// MongoDBLecturerRepo implements LecturerRepository using MongoDB
type MongoDBLecturerRepo struct {
	Collection *mongo.Collection
}

// MySQLLibraryRepo implements LecturerRepository using MySQL
type MySQLLibraryRepo struct {
	DB *sql.DB
}

// MongoDBLibraryRepo implements LecturerRepository using MongoDB
type MongoDBLibraryRepo struct {
	Collection       *mongo.Collection
	BorrowCollection *mongo.Collection
}

// Database connection
// ConnectMySQL establishes a connection to MySQL using the DSN
func ConnectMySQL() (*sql.DB, error) {

	db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	return db, nil
}

// ConnectMongo establishes a connection to MongoDB using the URI
func ConnectMongo() (*mongo.Database, error) {

	client, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(os.Getenv("MONGO_URI")),
	)

	if err != nil {
		return nil, err
	}

	return client.Database(os.Getenv("MONGO_DB")), nil
}

// Global context used for Redis operations
var Ctx = context.Background()

// ConnectRedis initializes and returns Redis client
func ConnectRedis() *redis.Client {

	// Create Redis client with address and DB
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})
	return rdb
}

// Repository Interface
// StudentRepository defines CRUD operations for Student entities.
type StudentRepository interface {
	CreateStudent(student Student) (*Student, error)

	GetAllStudent() ([]Student, error)

	GetByIDStudent(id int) (*Student, error)

	UpdateStudent(student Student) error

	DeleteStudent(id int) error
}

// LecturerRepository defines CRUD operations for Lecturer entities.
type LecturerRepository interface {
	CreateLecturer(l Lecturer) (*Lecturer, error)

	GetAllLecturer() ([]Lecturer, error)

	GetByIDLecturer(id int) (*Lecturer, error)

	UpdateLecturer(l Lecturer) error

	DeleteLecturer(id int) error
}

// LecturerRepository defines CRUD operations for Lecturer entities.
type LibraryRepository interface {
	CreateLibrary(l Library) (*Library, error)

	GetAllLibrary() ([]Library, error)

	GetByIDLibrary(id int) (*Library, error)

	UpdateLibrary(l Library) error

	DeleteLibrary(id int) error

	BorrowBook(info BorrowInfo) error
	ReturnBook(info BorrowInfo) error

	CheckUserExists(userID int, userType string) (bool, error)
}

// Handler Repository Selectors

// GetRepo selects the appropriate StudentRepository implementation
func (h *StudentHandler) GetRepo(r *http.Request) StudentRepository {
	dbType := r.URL.Query().Get("db")

	if dbType == "mongo" {
		return h.MongoRepo
	}

	// default
	return h.MySQLRepo
}

// GetRepo selects the appropriate LecturerRepository implementation
func (h *LecturerHandler) GetRepo(r *http.Request) LecturerRepository {
	dbType := r.URL.Query().Get("db")

	if dbType == "mongo" {
		return h.MongoRepo
	}

	// default
	return h.MySQLRepo
}

// GetRepo selects the appropriate LibraryRepository implementation
func (h *LibraryHandler) GetRepo(r *http.Request) LibraryRepository {
	dbType := r.URL.Query().Get("db")

	if dbType == "mongo" {
		return h.MongoRepo
	}

	// default
	return h.MySQLRepo
}

// Background Utilities
// Logging (gouroutine safe)
func LogActivity(action, actor string) {
	log.Printf("[LOG] %s by %s at %s\n", action, actor, time.Now())
}

// Audit Trail
func AuditLog(action, entity string, id any, actor string) {
	log.Printf("[AUDIT] avtion=%s, entity=%s , id=%v , actor=%s , time=%s\n", action, entity, id, actor, time.Now())
}

// @title College Management System API
// @version 1.0
// @description REST API for College Management System
// @termsOfService http://example.com/terms/

// @contact.name Akash Paul
// @contact.email akashpaul@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// College_Management_System initializes all services
// and starts HTTP server
func College_Management_System() {

	// Load environment variables
	godotenv.Load()

	SecretKey = []byte(os.Getenv("JWT_SECRET"))

	// Initialize Redis client
	redisClient := ConnectRedis()

	// Connect to MySQl
	mysqlDB, err := ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	// Create repositories
	mysqlStudent := NewMySQLStudentRepo(mysqlDB)
	MongoStudent := NewMongoDBStudentRepo(mongodb.Collection("students"))

	mysqlLecturer := NewMySQLLecturerRepo(mysqlDB)
	mongoLecturer := NewMongoDBLecturerRepo(mongodb.Collection("lecturers"))

	mysqlLibrary := NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := NewMongoDBLibraryRepo(mongodb.Collection("libraries"), mongodb.Collection("borrow_records"))

	// Create handlers with dependencies
	Studentshandler := &StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent, Redis: redisClient}

	Lecturershandler := &LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer, Redis: redisClient}

	Libraryshandler := &LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	// Initialize router
	r := mux.NewRouter()

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Authentication routes
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/refresh", RefreshHandler).Methods("POST")
	r.HandleFunc("/logout", LogoutHandler).Methods("POST")

	// Protected route
	api := r.PathPrefix("/api").Subrouter()
	api.Use(JwtMiddleware)

	// Student routes
	api.HandleFunc("/students", Studentshandler.CreateStudent).Methods("POST")
	api.HandleFunc("/students", Studentshandler.GetAllStudent).Methods("GET")
	api.HandleFunc("/students/{id}", Studentshandler.GetByIDStudent).Methods("GET")
	api.HandleFunc("/students/{id}", Studentshandler.UpdateStudent).Methods("PUT")
	api.HandleFunc("/students/{id}", Studentshandler.DeleteStudent).Methods("DELETE")

	// Lecturer routes
	api.HandleFunc("/lecturers", Lecturershandler.CreateLecturer).Methods("POST")
	api.HandleFunc("/lecturers", Lecturershandler.GetAllLecturer).Methods("GET")
	api.HandleFunc("/lecturers/{id}", Lecturershandler.GetByIDLectuurer).Methods("GET")
	api.HandleFunc("/lecturers/{id}", Lecturershandler.UpdateLecturer).Methods("PUT")
	api.HandleFunc("/lecturers/{id}", Lecturershandler.DeleteLecturer).Methods("DELETE")

	// Library routes
	api.HandleFunc("/libraries", Libraryshandler.CreateLibrary).Methods("POST")
	api.HandleFunc("/libraries", Libraryshandler.GetAllLibrary).Methods("GET")
	api.HandleFunc("/libraries/{id}", Libraryshandler.GetByIDLibrary).Methods("GET")
	api.HandleFunc("/libraries/{id}", Libraryshandler.UpdateLibrary).Methods("PUT")
	api.HandleFunc("/libraries/{id}", Libraryshandler.DeleteLibrary).Methods("DELETE")

	// borrow and return routes
	api.HandleFunc("/borrow", Libraryshandler.BorrowBookHandler).Methods("POST")
	api.HandleFunc("/return", Libraryshandler.ReturnBookHandler).Methods("POST")

	// Start HTTP server
	fmt.Println("Server running on port:8080")
	http.ListenAndServe(":8080", r)

}
