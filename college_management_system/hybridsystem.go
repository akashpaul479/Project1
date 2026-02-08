package collegemanagementsystem

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	Collection *mongo.Collection
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

// College_Management_System initializes all services
// and starts HTTP server
func College_Management_System() {

	// Load environment variables
	godotenv.Load()

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
	mongoLibrary := NewMongoDBLibraryRepo(mongodb.Collection("library"))

	// Create handlers with dependencies
	Studentshandler := &StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent, Redis: redisClient}

	Lecturershandler := &LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer, Redis: redisClient}

	Libraryshandler := &LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	// Initialize router
	r := mux.NewRouter()

	// Student routes
	r.HandleFunc("/students", Studentshandler.CreateStudent).Methods("POST")
	r.HandleFunc("/students", Studentshandler.GetAllStudent).Methods("GET")
	r.HandleFunc("/students/{id}", Studentshandler.GetByIDStudent).Methods("GET")
	r.HandleFunc("/students/{id}", Studentshandler.UpdateStudent).Methods("PUT")
	r.HandleFunc("/students/{id}", Studentshandler.DeleteStudent).Methods("DELETE")

	// Lecturer routes
	r.HandleFunc("/lecturers", Lecturershandler.CreateLecturer).Methods("POST")
	r.HandleFunc("/lecturers", Lecturershandler.GetAllLecturer).Methods("GET")
	r.HandleFunc("/lecturers/{id}", Lecturershandler.GetByIDLectuurer).Methods("GET")
	r.HandleFunc("/lecturers/{id}", Lecturershandler.UpdateLecturer).Methods("PUT")
	r.HandleFunc("/lecturers/{id}", Lecturershandler.DeleteLecturer).Methods("DELETE")

	// Library routes
	r.HandleFunc("/libraries", Libraryshandler.CreateLibrary).Methods("POST")
	r.HandleFunc("/libraries", Libraryshandler.GetAllLibrary).Methods("GET")
	r.HandleFunc("/libraries/{id}", Libraryshandler.GetByIDLibrary).Methods("GET")
	r.HandleFunc("/libraries/{id}", Libraryshandler.UpdateLibrary).Methods("PUT")
	r.HandleFunc("/libraries/{id}", Libraryshandler.DeleteLibrary).Methods("DELETE")

	// Start HTTP server
	fmt.Println("Server running on port:8080")
	http.ListenAndServe(":8080", r)

}
