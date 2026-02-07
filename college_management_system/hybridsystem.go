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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MySQLStudentRepo struct {
	DB *sql.DB
}

type MongoDBStudentRepo struct {
	Collection *mongo.Collection
}

type MySQLLecturerRepo struct {
	DB *sql.DB
}

type MongoDBLecturerRepo struct {
	Collection *mongo.Collection
}

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

type StudentRepository interface {
	CreateStudent(student Student) (*Student, error)

	GetAllStudent() ([]Student, error)

	GetByIDStudent(id int) (*Student, error)

	UpdateStudent(student Student) error

	DeleteStudent(id int) error
}

type LecturerRepository interface {
	CreateLecturer(l Lecturer) (*Lecturer, error)

	GetAllLecturer() ([]Lecturer, error)

	GetByIDLecturer(id int) (*Lecturer, error)

	UpdateLecturer(l Lecturer) error

	DeleteLecturer(id int) error
}

func (h *StudentHandler) GetRepo(r *http.Request) StudentRepository {
	dbType := r.URL.Query().Get("db")

	if dbType == "mongo" {
		return h.MongoRepo
	}

	// default
	return h.MySQLRepo
}
func (h *LecturerHandler) GetRepo(r *http.Request) LecturerRepository {
	dbType := r.URL.Query().Get("db")

	if dbType == "mongo" {
		return h.MongoRepo
	}

	// default
	return h.MySQLRepo
}

func College_Management_System() {
	godotenv.Load()

	// Connect MySQl
	mysqlDB, err := ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	//  Connect MongoDB
	mongodb, err := ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlStudent := NewMySQLStudentRepo(mysqlDB)
	MongoStudent := NewMongoDBStudentRepo(mongodb.Collection("students"))

	mysqlLecturer := NewMySQLLecturerRepo(mysqlDB)
	mongoLecturer := NewMongoDBLecturerRepo(mongodb.Collection("lecturers"))

	Studentshandler := &StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent}

	Lecturershandler := &LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer}

	r := mux.NewRouter()

	r.HandleFunc("/students", Studentshandler.CreateStudent).Methods("POST")
	r.HandleFunc("/students", Studentshandler.GetAllStudent).Methods("GET")
	r.HandleFunc("/students/{id}", Studentshandler.GetByIDStudent).Methods("GET")
	r.HandleFunc("/students/{id}", Studentshandler.UpdateStudent).Methods("PUT")
	r.HandleFunc("/students/{id}", Studentshandler.DeleteStudent).Methods("DELETE")

	r.HandleFunc("/lecturers", Lecturershandler.CreateLecturer).Methods("POST")
	r.HandleFunc("/lecturers", Lecturershandler.GetAllLecturer).Methods("GET")
	r.HandleFunc("/lecturers/{id}", Lecturershandler.GetByIDLectuurer).Methods("GET")
	r.HandleFunc("/lecturers/{id}", Lecturershandler.UpdateLecturer).Methods("PUT")
	r.HandleFunc("/lecturers/{id}", Lecturershandler.DeleteLecturer).Methods("DELETE")

	fmt.Println("Server running on port:8080")

	http.ListenAndServe(":8080", r)

}
