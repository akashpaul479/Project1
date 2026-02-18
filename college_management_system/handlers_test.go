package collegemanagementsystem_test

import (
	"bytes"
	collegemanagementsystem "college_management_system/college_management_system"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

func TestStudentHandler_CreateStudent(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		student  collegemanagementsystem.Student
		willpass bool
	}{
		{
			name:   "mysql valid Student",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name:   "mongo valid Student",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name:   "mysql invalid name or empty name",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo invalid name or empty name",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql withspace name",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "   ",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo withspace name",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "   ",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql Invaild Age",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo Invaild Age",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid email",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid email",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql gmail without prefix",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo gmail without prefix",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid dept",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid dept",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "",
			},
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	mysqlStudent := collegemanagementsystem.NewMySQLStudentRepo(mysqlDB)
	MongoStudent := collegemanagementsystem.NewMongoDBStudentRepo(mongodb.Collection("students"))

	Studentshandler := &collegemanagementsystem.StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM students")
			mongodb.Collection("students").DeleteMany(context.Background(), map[string]interface{}{})
			redisClient.FlushAll(context.Background())

			userBytes, err := json.Marshal(tt.student)
			if err != nil {
				log.Panic("failed to marshal")
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/students?db="+tt.dbtype, buffer)
			w := httptest.NewRecorder()

			Studentshandler.CreateStudent(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected OK status , got %d", w.Code)
				}
				// validate Response
				var students collegemanagementsystem.Student
				if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if students.Name != tt.student.Name {
					t.Fatalf("Expected name %s, got %s", tt.student.Name, students.Name)
				}
				if students.Email != tt.student.Email {
					t.Fatalf("Expected email %s got %s", tt.student.Email, students.Email)
				}
				if students.Dept != tt.student.Dept {
					t.Fatalf("Expected dept %s got %s", tt.student.Dept, students.Dept)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestStudentHandler_GetAllStudent(t *testing.T) {
	tests := []struct {
		name          string // description of this test case
		dbtype        string
		insertcount   int
		expectedcount int
	}{
		{
			name:          "mysql empty database",
			dbtype:        "mysql",
			insertcount:   0,
			expectedcount: 0,
		},
		{
			name:          "mongo empty database",
			dbtype:        "mongo",
			insertcount:   0,
			expectedcount: 0,
		},
		{
			name:          "mysql one student",
			dbtype:        "mysql",
			insertcount:   1,
			expectedcount: 1,
		},
		{
			name:          "mongo one student",
			dbtype:        "mongo",
			insertcount:   1,
			expectedcount: 1,
		},
		{
			name:          "mysql multiple students",
			dbtype:        "mysql",
			insertcount:   3,
			expectedcount: 3,
		},
		{
			name:          "mongo multiple students",
			dbtype:        "mongo",
			insertcount:   3,
			expectedcount: 3,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	mysqlStudent := collegemanagementsystem.NewMySQLStudentRepo(mysqlDB)
	MongoStudent := collegemanagementsystem.NewMongoDBStudentRepo(mongodb.Collection("students"))

	Studentshandler := &collegemanagementsystem.StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM students")
			mongodb.Collection("students").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			for i := 0; i < tt.insertcount; i++ {
				if tt.dbtype == "mysql" {
					mysqlDB.Exec("INSERT INTO students(name , age , email , dept) VALUES (? , ? , ? , ?)", "Akash", 22, "akash@gmail.com", "CSE")
				} else {
					mongodb.Collection("students").InsertOne(context.Background(), bson.M{"id": i + 1, "name": "Akash", "age": 22, "email": "akash@gmail.com", "dept": "CSE"})
				}
			}
			r := httptest.NewRequest(http.MethodGet, "/students?db="+tt.dbtype, nil)
			w := httptest.NewRecorder()

			Studentshandler.GetAllStudent(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected Ok status , got %d", w.Code)
			}
			var students []collegemanagementsystem.Student

			if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			// validate count
			if len(students) != tt.expectedcount {
				t.Fatalf("Expected %d students , got %d", tt.expectedcount, len(students))
			}
		})
	}
}

func TestStudentHandler_GetByIDStudent(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		id       int
		willpass bool
	}{
		{
			name:     "mysql valid id",
			dbtype:   "mysql",
			id:       1,
			willpass: true,
		},
		{
			name:     "mongo valid id",
			dbtype:   "mongo",
			id:       1,
			willpass: true,
		},
		{
			name:     "mysql Invalid id",
			dbtype:   "mysql",
			id:       0,
			willpass: false,
		},
		{
			name:     "mongo Invalid id",
			dbtype:   "mongo",
			id:       0,
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	mysqlStudent := collegemanagementsystem.NewMySQLStudentRepo(mysqlDB)
	MongoStudent := collegemanagementsystem.NewMongoDBStudentRepo(mongodb.Collection("students"))

	Studentshandler := &collegemanagementsystem.StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM students")
			mongodb.Collection("students").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var validID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO students (name, age, email, dept) VALUES (? , ? , ? , ?)", "Akash", 22, "akashpaul@gmail.com", "CSE")
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					validID = int(id)
				} else {
					validID = 1
					mongodb.Collection("students").InsertOne(context.Background(), bson.M{"id": validID, "name": "Akash", "age": 22, "email": "akashpaul@gmail.com", "dept": "CSE"})

				}
			} else {
				// Invalid ID not in DB
				validID = 9999
			}

			r := httptest.NewRequest(http.MethodGet, "/students?db="+tt.dbtype, nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(validID)})
			w := httptest.NewRecorder()

			Studentshandler.GetByIDStudent(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var students collegemanagementsystem.Student
				if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if students.ID != validID {
					t.Fatalf("Expected ID %d , got %d", validID, students.ID)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected Not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestStudentHandler_UpdateStudent(t *testing.T) {

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	mysqlStudent := collegemanagementsystem.NewMySQLStudentRepo(mysqlDB)
	MongoStudent := collegemanagementsystem.NewMongoDBStudentRepo(mongodb.Collection("students"))

	Studentshandler := &collegemanagementsystem.StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent, Redis: redisClient}

	tests := []struct {
		name     string // description of this test case
		dbtype   string
		student  collegemanagementsystem.Student
		willpass bool
	}{
		{
			name:   "mysql Valid Update",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash paul",
				Age:   21,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name:   "mongo Valid Update",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash paul",
				Age:   21,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name:   "mysql invalid name",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   21,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid name ",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   21,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql withspace name",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "   ",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo withspace name",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "   ",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql Invaild Age",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo Invaild Age",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid email",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid email",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql gmail without prefix",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mongo gmail without prefix",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid dept",
			dbtype: "mysql",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid dept",
			dbtype: "mongo",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "",
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM students")
			mongodb.Collection("students").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var studentID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO students (name, age, email, dept) VALUES (? , ? , ? , ?)", "Akash", 22, "akash@gmail.com", "CSE")
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					studentID = int(id)
				} else {
					studentID = 1
					mongodb.Collection("students").InsertOne(context.Background(), bson.M{"id": studentID, "name": "Akash", "age": 22, "email": "akash@gmail.com", "dept": "CSE"})

				}
			} else {
				// Invalid ID not in DB
				studentID = 9999
			}
			tt.student.ID = studentID

			userBytes, err := json.Marshal(tt.student)
			if err != nil {
				panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPut, "/students/"+strconv.Itoa(studentID)+"?db="+tt.dbtype, buffer)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(studentID)})
			w := httptest.NewRecorder()

			Studentshandler.UpdateStudent(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected OK status , got %d", w.Code)
				}

				if tt.dbtype == "mysql" {
					var name string

					err := mysqlDB.QueryRow("SELECT name FROM students WHERE id=?", studentID).Scan(&name)
					if err != nil {
						t.Fatal(err)
					}
					if name != tt.student.Name {
						t.Fatalf("Expected name %s , got %s", tt.student.Name, name)
					}
				} else {
					var res bson.M

					err := mongodb.Collection("students").FindOne(context.Background(), bson.M{"id": studentID}).Decode(&res)
					if err != nil {
						t.Fatal(err)
					}
					if res["name"] != tt.student.Name {
						t.Fatalf("Expected %v , got %v", tt.student.Name, res["name"])
					}
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestStudentHandler_DeleteStudent(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		id       int
		willpass bool
	}{
		{
			name:     "mysql valid id",
			dbtype:   "mysql",
			id:       1,
			willpass: true,
		},
		{
			name:     "mongo valid id",
			dbtype:   "mongo",
			id:       1,
			willpass: true,
		},
		{
			name:     "mysql Invalid id",
			dbtype:   "mysql",
			id:       9999,
			willpass: false,
		},
		{
			name:     "mongo Invalid id",
			dbtype:   "mongo",
			id:       9999,
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	mysqlStudent := collegemanagementsystem.NewMySQLStudentRepo(mysqlDB)
	MongoStudent := collegemanagementsystem.NewMongoDBStudentRepo(mongodb.Collection("students"))

	Studentshandler := &collegemanagementsystem.StudentHandler{MySQLRepo: mysqlStudent, MongoRepo: MongoStudent, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM students")
			mongodb.Collection("students").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var studentID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO students (name, age, email, dept) VALUES (? , ? , ? , ?)", "Akash", 22, "akashpaul@gmail.com", "CSE")
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					studentID = int(id)
				} else {
					studentID = 1
					mongodb.Collection("students").InsertOne(context.Background(), bson.M{"id": studentID, "name": "Akash", "age": 22, "email": "akashpaul@gmail.com", "dept": "CSE"})

				}
			} else {
				// Invalid ID not in DB
				studentID = 9999
			}

			r := httptest.NewRequest(http.MethodDelete, "/students/"+strconv.Itoa(studentID)+"?db="+tt.dbtype, nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(studentID)})
			w := httptest.NewRecorder()

			Studentshandler.DeleteStudent(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected Ok status , got %d", w.Code)
				}
				var res map[string]string
				json.NewDecoder(w.Body).Decode(&res)

				if res["status"] != "deleted" {
					t.Fatalf("Expected deleted, got %v", res)
				}

			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not Ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestLecturerHandler_CreateLecturer(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		lecturer collegemanagementsystem.Lecturer
		willpass bool
	}{
		{
			name:   "mysql valid Lecturer",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name:   "mongo valid Lecturer",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name:   "mysql invalid name or empty name",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo invalid name or empty name",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql withspace name",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "   ",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo withspace name",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "   ",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql Invaild Age",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo Invaild Age",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid email",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid email",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql gmail without prefix",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo gmail without prefix",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid designation",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid designation",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "",
			},
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	mysqlLecturer := collegemanagementsystem.NewMySQLLecturerRepo(mysqlDB)
	mongoLecturer := collegemanagementsystem.NewMongoDBLecturerRepo(mongodb.Collection("lecturers"))

	Lecturershandler := &collegemanagementsystem.LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UserBytes, err := json.Marshal(tt.lecturer)
			if err != nil {
				panic(err)
			}
			buffer := bytes.NewBuffer(UserBytes)

			r := httptest.NewRequest(http.MethodPost, "/lecturers?db="+tt.dbtype, buffer)
			w := httptest.NewRecorder()
			Lecturershandler.CreateLecturer(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected Ok status , got %d", w.Code)
				}
				var lectureres collegemanagementsystem.Lecturer
				if err := json.NewDecoder(w.Body).Decode(&lectureres); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if lectureres.Name != tt.lecturer.Name {
					t.Fatalf("Expected name %s, got %s", tt.lecturer.Name, lectureres.Name)
				}
				if lectureres.Email != tt.lecturer.Email {
					t.Fatalf("Expected email %s got %s", tt.lecturer.Email, lectureres.Email)
				}
				if lectureres.Designation != tt.lecturer.Designation {
					t.Fatalf("Expected dept %s got %s", tt.lecturer.Designation, lectureres.Designation)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestLecturerHandler_GetAllLecturer(t *testing.T) {
	tests := []struct {
		name          string // description of this test case
		dbtype        string
		insertcount   int
		expectedcount int
	}{
		{
			name:          "mysql empty database",
			dbtype:        "mysql",
			insertcount:   0,
			expectedcount: 0,
		},
		{
			name:          "mongo empty database",
			dbtype:        "mongo",
			insertcount:   0,
			expectedcount: 0,
		},
		{
			name:          "mysql one lecturer",
			dbtype:        "mysql",
			insertcount:   1,
			expectedcount: 1,
		},
		{
			name:          "mongo one lecturer",
			dbtype:        "mongo",
			insertcount:   1,
			expectedcount: 1,
		},
		{
			name:          "mysql multiple lecturers",
			dbtype:        "mysql",
			insertcount:   3,
			expectedcount: 3,
		},
		{
			name:          "mongo multiple lecturers",
			dbtype:        "mongo",
			insertcount:   3,
			expectedcount: 3,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	mysqlLecturer := collegemanagementsystem.NewMySQLLecturerRepo(mysqlDB)
	mongoLecturer := collegemanagementsystem.NewMongoDBLecturerRepo(mongodb.Collection("lecturers"))

	Lecturershandler := &collegemanagementsystem.LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM lecturers")
			mongodb.Collection("lecturers").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			for i := 0; i < tt.insertcount; i++ {
				if tt.dbtype == "mysql" {
					mysqlDB.Exec("INSERT INTO lecturers(name , age , email , designation) VALUES (? , ? , ? , ?)", "Akash", 22, "akash@gmail.com", "HOD")
				} else {
					mongodb.Collection("lecturers").InsertOne(context.Background(), bson.M{"id": i + 1, "name": "Akash", "age": 22, "email": "akash@gmail.com", "designation": "HOD"})
				}
			}
			r := httptest.NewRequest(http.MethodGet, "/lecturers?db="+tt.dbtype, nil)
			w := httptest.NewRecorder()

			Lecturershandler.GetAllLecturer(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected ok status , got %d", w.Code)
			}
			var lecturers []collegemanagementsystem.Lecturer
			if err := json.NewDecoder(w.Body).Decode(&lecturers); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if len(lecturers) != tt.expectedcount {
				t.Fatalf("Expected %d lecturers , got %d", tt.expectedcount, len(lecturers))
			}
		})
	}
}

func TestLecturerHandler_GetByIDLectuurer(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		id       int
		willpass bool
	}{
		{
			name:     "mysql valid id",
			dbtype:   "mysql",
			id:       1,
			willpass: true,
		},
		{
			name:     "mongo valid id",
			dbtype:   "mongo",
			id:       1,
			willpass: true,
		},
		{
			name:     "mysql Invalid id",
			dbtype:   "mysql",
			id:       0,
			willpass: false,
		},
		{
			name:     "mongo Invalid id",
			dbtype:   "mongo",
			id:       0,
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLecturer := collegemanagementsystem.NewMySQLLecturerRepo(mysqlDB)
	mongoLecturer := collegemanagementsystem.NewMongoDBLecturerRepo(mongodb.Collection("lecturers"))

	Lecturershandler := &collegemanagementsystem.LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM lecturers")
			mongodb.Collection("lecturers").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var validID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO lecturers (name, age, email, designation) VALUES (? , ? , ? , ?)", "Akash", 22, "akashpaul@gmail.com", "HOD")
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					validID = int(id)
				} else {
					validID = 1
					mongodb.Collection("lecturers").InsertOne(context.Background(), bson.M{"id": validID, "name": "Akash", "age": 22, "email": "akashpaul@gmail.com", "dept": "HOD"})

				}
			} else {
				// Invalid ID not in DB
				validID = 9999
			}

			r := httptest.NewRequest(http.MethodGet, "/lecturers?db="+tt.dbtype, nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(validID)})
			w := httptest.NewRecorder()

			Lecturershandler.GetByIDLectuurer(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected OK status , got %d", w.Code)
				}
				var lecturers collegemanagementsystem.Lecturer
				if err := json.NewDecoder(w.Body).Decode(&lecturers); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if lecturers.ID != validID {
					t.Fatalf("Expected Id %d , got %d", validID, lecturers.ID)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestLecturerHandler_UpdateLecturer(t *testing.T) {

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLecturer := collegemanagementsystem.NewMySQLLecturerRepo(mysqlDB)
	mongoLecturer := collegemanagementsystem.NewMongoDBLecturerRepo(mongodb.Collection("lecturers"))

	Lecturershandler := &collegemanagementsystem.LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer, Redis: redisClient}

	tests := []struct {
		name     string // description of this test case
		dbtype   string
		lecturer collegemanagementsystem.Lecturer
		willpass bool
	}{
		{
			name:   "mysql Valid Update",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash paul",
				Age:         21,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name:   "mongo Valid Update",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash paul",
				Age:         21,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name:   "mysql invalid name",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         21,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid name ",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         21,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql withspace name",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "   ",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo withspace name",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "   ",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql Invaild Age",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo Invaild Age",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid email",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid email",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql gmail without prefix",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mongo gmail without prefix",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name:   "mysql Invalid designation",
			dbtype: "mysql",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "",
			},
			willpass: false,
		},
		{
			name:   "mongo Invalid designation",
			dbtype: "mongo",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "",
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM lecturers")
			mongodb.Collection("lecturers").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var lecturerID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO lecturers (name, age, email, designation) VALUES (? , ? , ? , ?)", "Akash", 22, "akash@gmail.com", "HOD")
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					lecturerID = int(id)
				} else {
					lecturerID = 1
					mongodb.Collection("lecturers").InsertOne(context.Background(), bson.M{"id": lecturerID, "name": "Akash", "age": 22, "email": "akash@gmail.com", "designation": "HOD"})

				}
			} else {
				// Invalid ID not in DB
				lecturerID = 9999
			}
			tt.lecturer.ID = lecturerID

			userBytes, err := json.Marshal(tt.lecturer)
			if err != nil {
				panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPut, "/lecturers/"+strconv.Itoa(lecturerID)+"?db="+tt.dbtype, buffer)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(lecturerID)})
			w := httptest.NewRecorder()

			Lecturershandler.UpdateLecturer(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected OK status , got %d", w.Code)
				}

				if tt.dbtype == "mysql" {
					var name string

					err := mysqlDB.QueryRow("SELECT name FROM lecturers WHERE id=?", lecturerID).Scan(&name)
					if err != nil {
						t.Fatal(err)
					}
					if name != tt.lecturer.Name {
						t.Fatalf("Expected name %s , got %s", tt.lecturer.Name, name)
					}
				} else {
					var res bson.M

					err := mongodb.Collection("lecturers").FindOne(context.Background(), bson.M{"id": lecturerID}).Decode(&res)
					if err != nil {
						t.Fatal(err)
					}
					if res["name"] != tt.lecturer.Name {
						t.Fatalf("Expected %v , got %v", tt.lecturer.Name, res["name"])
					}
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestLecturerHandler_DeleteLecturer(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		id       int
		willpass bool
	}{
		{
			name:     "mysql valid id",
			dbtype:   "mysql",
			id:       1,
			willpass: true,
		},
		{
			name:     "mongo valid id",
			dbtype:   "mongo",
			id:       1,
			willpass: true,
		},
		{
			name:     "mysql Invalid id",
			dbtype:   "mysql",
			id:       9999,
			willpass: false,
		},
		{
			name:     "mongo Invalid id",
			dbtype:   "mongo",
			id:       9999,
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLecturer := collegemanagementsystem.NewMySQLLecturerRepo(mysqlDB)
	mongoLecturer := collegemanagementsystem.NewMongoDBLecturerRepo(mongodb.Collection("lecturers"))

	Lecturershandler := &collegemanagementsystem.LecturerHandler{MySQLRepo: mysqlLecturer, MongoRepo: mongoLecturer, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysqlDB.Exec("DELETE FROM lecturers")
			mongodb.Collection("lecturers").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var lecturerID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO lecturers (name, age, email, designation) VALUES (? , ? , ? , ?)", "Akash", 22, "akashpaul@gmail.com", "HOD")
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					lecturerID = int(id)
				} else {
					lecturerID = 1
					mongodb.Collection("lecturers").InsertOne(context.Background(), bson.M{"id": lecturerID, "name": "Akash", "age": 22, "email": "akashpaul@gmail.com", "designation": "HOD"})

				}
			} else {
				// Invalid ID not in DB
				lecturerID = 9999
			}

			r := httptest.NewRequest(http.MethodDelete, "/lecturers/"+strconv.Itoa(lecturerID)+"?db="+tt.dbtype, nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(lecturerID)})
			w := httptest.NewRecorder()

			Lecturershandler.DeleteLecturer(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected Ok status , got %d", w.Code)
				}
				var res map[string]string
				json.NewDecoder(w.Body).Decode(&res)

				if res["status"] != "deleted" {
					t.Fatalf("Expected deleted, got %v", res)
				}

			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not Ok status , got %d", w.Code)
				}
			}

		})
	}
}

func TestLibraryHandler_CreateLibrary(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		library  collegemanagementsystem.Library
		willpass bool
	}{
		{
			name:   "mysql valid Library",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name:   "mongo valid Library",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name:   "mysql invalid book_name",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid book_name",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql withspace book_name",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "   ",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo withspace book_name",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "   ",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid title",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid title",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid author",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid author",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid available_copies",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid available_copies",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
		{
			name:   "mysql withspace title",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "   ",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo withspace title",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "   ",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
		{
			name:   "mysql withspace author",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "   ",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo withspace author",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 0,
			},
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLibrary := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := collegemanagementsystem.NewMongoDBLibraryRepo(mongodb.Collection("library"), mongodb.Collection("borrow_records"))

	Libraryshandler := &collegemanagementsystem.LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysqlDB.Exec("DELETE FROM libraries")
			mongodb.Collection("libraries").DeleteMany(context.Background(), map[string]interface{}{})
			redisClient.FlushAll(context.Background())

			userBytes, err := json.Marshal(tt.library)
			if err != nil {
				log.Panic("failed to marshal")
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/libraries?db="+tt.dbtype, buffer)
			w := httptest.NewRecorder()
			Libraryshandler.CreateLibrary(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected Ok status , got %d", w.Code)
				}
				var libraries collegemanagementsystem.Library
				if err := json.NewDecoder(w.Body).Decode(&libraries); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if libraries.Book_name != tt.library.Book_name {
					t.Fatalf("Expected book_name %s, got %s", tt.library.Book_name, libraries.Book_name)
				}
				if libraries.Title != tt.library.Title {
					t.Fatalf("Expected title %s got %s", tt.library.Title, libraries.Title)
				}
				if libraries.Author != tt.library.Author {
					t.Fatalf("Expected author %s got %s", tt.library.Author, libraries.Author)
				}
				if libraries.Available_copies != tt.library.Available_copies {
					t.Fatalf("Expected available_copies %d , got %d", tt.library.Available_copies, libraries.Available_copies)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestLibraryHandler_GetAllLibrary(t *testing.T) {
	tests := []struct {
		name          string // description of this test case
		dbtype        string
		insertcount   int
		expectedcount int
	}{
		{
			name:          "mysql empty database",
			dbtype:        "mysql",
			insertcount:   0,
			expectedcount: 0,
		},
		{
			name:          "mongo empty database",
			dbtype:        "mongo",
			insertcount:   0,
			expectedcount: 0,
		},
		{
			name:          "mysql one library_book",
			dbtype:        "mysql",
			insertcount:   1,
			expectedcount: 1,
		},
		{
			name:          "mongo one library_book",
			dbtype:        "mongo",
			insertcount:   1,
			expectedcount: 1,
		},
		{
			name:          "mysql multiple library_books",
			dbtype:        "mysql",
			insertcount:   3,
			expectedcount: 3,
		},
		{
			name:          "mongo multiple library_books",
			dbtype:        "mongo",
			insertcount:   3,
			expectedcount: 3,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLibrary := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := collegemanagementsystem.NewMongoDBLibraryRepo(mongodb.Collection("libraries"), mongodb.Collection("borrow_records"))

	Libraryshandler := &collegemanagementsystem.LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysqlDB.Exec("SET FOREIGN_KEY_CHECKS=0")
			mysqlDB.Exec("TRUNCATE TABLE borrow_records")
			mysqlDB.Exec("TRUNCATE TABLE libraries")
			mysqlDB.Exec("SET FOREIGN_KEY_CHECKS=1")

			mongodb.Collection("libraries").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			for i := 0; i < tt.insertcount; i++ {
				if tt.dbtype == "mysql" {
					mysqlDB.Exec("INSERT INTO libraries(book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)
				} else {
					mongodb.Collection("libraries").InsertOne(context.Background(), bson.M{"book_id": i + 1, "book_name": "comics", "title": "The boys", "author": "abhi", "available_copies": 10})
				}
			}
			r := httptest.NewRequest(http.MethodGet, "/libraries?db="+tt.dbtype, nil)
			w := httptest.NewRecorder()

			Libraryshandler.GetAllLibrary(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected Ok status , got %d", w.Code)
			}
			var libraries []collegemanagementsystem.Library

			if err := json.NewDecoder(w.Body).Decode(&libraries); err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			// validate count
			if len(libraries) != tt.expectedcount {
				t.Fatalf("Expected %d libraries , got %d", tt.expectedcount, len(libraries))
			}

		})
	}
}

func TestLibraryHandler_GetByIDLibrary(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		id       int
		willpass bool
	}{
		{
			name:     "mysql valid id",
			dbtype:   "mysql",
			id:       1,
			willpass: true,
		},
		{
			name:     "mongo valid id",
			dbtype:   "mongo",
			id:       1,
			willpass: true,
		},
		{
			name:     "mysql Invalid id",
			dbtype:   "mysql",
			id:       0,
			willpass: false,
		},
		{
			name:     "mongo Invalid id",
			dbtype:   "mongo",
			id:       0,
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLibrary := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := collegemanagementsystem.NewMongoDBLibraryRepo(mongodb.Collection("libraries"), mongodb.Collection("borrow_records"))

	Libraryshandler := &collegemanagementsystem.LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM libraries")
			mongodb.Collection("libraries").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var validID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO libraries (book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					validID = int(id)
				} else {
					validID = 1
					mongodb.Collection("libraries").InsertOne(context.Background(), bson.M{"book_id": validID, "book_name": "comics", "title": "The boys", "author": "abhi", "available_copies": 10})

				}
			} else {
				// Invalid ID not in DB
				validID = 9999
			}

			r := httptest.NewRequest(http.MethodGet, "/libraries?db="+tt.dbtype, nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(validID)})
			w := httptest.NewRecorder()

			Libraryshandler.GetByIDLibrary(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var libraries collegemanagementsystem.Library
				if err := json.NewDecoder(w.Body).Decode(&libraries); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if libraries.Book_id != validID {
					t.Fatalf("Expected ID %d , got %d", validID, libraries.Book_id)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected Not OK status , got %d", w.Code)
				}
			}

		})
	}
}

func TestLibraryHandler_UpdateLibrary(t *testing.T) {

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLibrary := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := collegemanagementsystem.NewMongoDBLibraryRepo(mongodb.Collection("libraries"), mongodb.Collection("borrow_records"))

	Libraryshandler := &collegemanagementsystem.LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	tests := []struct {
		name     string // description of this test case
		dbtype   string
		library  collegemanagementsystem.Library
		willpass bool
	}{
		{
			name:   "mysql valid Library",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "cartoon",
				Title:            "The boys",
				Author:           "abhi biswas",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name:   "mongo valid Library",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "cartoon",
				Title:            "The boys",
				Author:           "abhi biswas",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name:   "mysql invalid book_name",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid book_name",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql withspace book_name",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "   ",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo withspace book_name",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "   ",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid title",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid title",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid author",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid author",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid available_copies",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid available_copies",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
		{
			name:   "mysql withspace title",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "   ",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo withspace title",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "   ",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
		{
			name:   "mysql withspace author",
			dbtype: "mysql",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "   ",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name:   "mongo withspace author",
			dbtype: "mongo",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 0,
			},
			willpass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM libraries")
			mongodb.Collection("libraries").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var libraryID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO libraries (book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					libraryID = int(id)
				} else {
					libraryID = 1
					mongodb.Collection("libraries").InsertOne(context.Background(), bson.M{"book_id": libraryID, "book_name": "comics", "title": "The boys", "author": "abhi", "available_copies": 10})

				}
			} else {
				// Invalid ID not in DB
				libraryID = 9999
			}

			tt.library.Book_id = libraryID

			userBytes, err := json.Marshal(tt.library)
			if err != nil {
				panic(err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPut, "/libraries/"+strconv.Itoa(libraryID)+"?db="+tt.dbtype, buffer)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(libraryID)})
			w := httptest.NewRecorder()

			Libraryshandler.UpdateLibrary(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected OK status , got %d", w.Code)
				}

				if tt.dbtype == "mysql" {
					var book_name string

					err := mysqlDB.QueryRow("SELECT book_name FROM libraries WHERE book_id=?", libraryID).Scan(&book_name)
					if err != nil {
						t.Fatal(err)
					}
					if book_name != tt.library.Book_name {
						t.Fatalf("Expected name %s , got %s", tt.library.Book_name, book_name)
					}
				} else {
					var res bson.M

					err := mongodb.Collection("libraries").FindOne(context.Background(), bson.M{"book_id": libraryID}).Decode(&res)
					if err != nil {
						t.Fatal(err)
					}
					if res["book_name"] != tt.library.Book_name {
						t.Fatalf("Expected %v , got %v", tt.library.Book_name, res["book_name"])
					}
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not OK status , got %d", w.Code)
				}
			}
		})
	}
}

func TestLibraryHandler_DeleteLibrary(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dbtype   string
		id       int
		willpass bool
	}{
		{
			name:     "mysql valid id",
			dbtype:   "mysql",
			id:       1,
			willpass: true,
		},
		{
			name:     "mongo valid id",
			dbtype:   "mongo",
			id:       1,
			willpass: true,
		},
		{
			name:     "mysql Invalid id",
			dbtype:   "mysql",
			id:       9999,
			willpass: false,
		},
		{
			name:     "mongo Invalid id",
			dbtype:   "mongo",
			id:       9999,
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLibrary := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := collegemanagementsystem.NewMongoDBLibraryRepo(mongodb.Collection("libraries"), mongodb.Collection("borrow_records"))

	Libraryshandler := &collegemanagementsystem.LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM libraries")
			mongodb.Collection("libraries").DeleteMany(context.Background(), bson.M{})
			redisClient.FlushAll(context.Background())

			var libraryID int

			if tt.willpass {
				if tt.dbtype == "mysql" {
					res, err := mysqlDB.Exec("INSERT INTO libraries (book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)
					if err != nil {
						t.Fatal(err)
					}
					id, _ := res.LastInsertId()
					libraryID = int(id)
				} else {
					libraryID = 1
					mongodb.Collection("libraries").InsertOne(context.Background(), bson.M{"book_id": libraryID, "book_name": "comics", "title": "The boys", "author": "abhi", "available_copies": 10})

				}
			} else {
				// Invalid ID not in DB
				libraryID = 9999
			}

			r := httptest.NewRequest(http.MethodDelete, "/students/"+strconv.Itoa(libraryID)+"?db="+tt.dbtype, nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(libraryID)})
			w := httptest.NewRecorder()

			Libraryshandler.DeleteLibrary(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected Ok status , got %d", w.Code)
				}
				var res map[string]string
				json.NewDecoder(w.Body).Decode(&res)

				if res["status"] != "deleted" {
					t.Fatalf("Expected deleted, got %v", res)
				}

			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not Ok status , got %d", w.Code)
				}
			}

		})
	}
}

func TestLibraryHandler_BorrowBookHandler(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		dbtype    string
		borroInfo collegemanagementsystem.BorrowInfo
		willpass  bool
	}{
		{
			name:   "mysql valid student borrow",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   1,
				UserType: "student",
				BookID:   1,
			},
			willpass: true,
		},
		{
			name:   "mongo valid student borrow",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   1,
				UserType: "student",
				BookID:   1,
			},
			willpass: true,
		},
		{
			name:   "mysql invalid userID",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid userID",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid usertype",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid usertype",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid bookID",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   0,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid bookID",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   0,
			},
			willpass: false,
		},
	}
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLibrary := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := collegemanagementsystem.NewMongoDBLibraryRepo(mongodb.Collection("libraries"), mongodb.Collection("borrow_records"))

	Libraryshandler := &collegemanagementsystem.LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM students")
			mysqlDB.Exec("DELETE FROM lecturers")
			mysqlDB.Exec("DELETE FROM libraries")
			mysqlDB.Exec("DELETE FROM borrow_records")

			mongodb.Collection("students").DeleteMany(context.Background(), map[string]interface{}{})
			mongodb.Collection("lecturers").DeleteMany(context.Background(), map[string]interface{}{})
			mongodb.Collection("libraries").DeleteMany(context.Background(), map[string]interface{}{})
			mongodb.Collection("borrow_records").DeleteMany(context.Background(), map[string]interface{}{})

			redisClient.FlushAll(context.Background())

			// Insert users

			if tt.borroInfo.UserType == "student" {
				mysqlDB.Exec("INSERT INTO students(id,name,age,email,dept) VALUES(1,'Akash',22,'akash@gmail.com','CSE')")

				mongodb.Collection("students").InsertOne(context.Background(), map[string]interface{}{"id": 1, "name": "Akash", "age": 22, "email": "akash@gmail.com", "dept": "CSE"})
			} else if tt.borroInfo.UserType == "lecturer" {

				mysqlDB.Exec("INSERT INTO lecturers(id,name,age,email,designation) VALUES(1,'Abhi',30,'abhi@gmail.com','Prof')")

				mongodb.Collection("lecturers").InsertOne(context.Background(),
					map[string]interface{}{
						"id":          1,
						"name":        "Abhi",
						"age":         30,
						"email":       "abhi@gmail.com",
						"designation": "Prof",
					})
			}

			// Insert Books
			mysqlDB.Exec(
				"INSERT INTO libraries(book_id,book_name,title,author,available_copies) VALUES(1,'Go','GoLang','Alan',5)",
			)

			mongodb.Collection("libraries").InsertOne(context.Background(),
				map[string]interface{}{
					"book_id":          1,
					"book_name":        "Go",
					"title":            "GoLang",
					"author":           "Alan",
					"available_copies": 5,
				})
			data := tt.borroInfo

			Userbytes, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			buffer := bytes.NewBuffer(Userbytes)
			r := httptest.NewRequest(http.MethodPost, "/borrow?db="+tt.dbtype, buffer)
			w := httptest.NewRecorder()

			Libraryshandler.BorrowBookHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected 200, got %d", w.Code)
				}

			} else {

				if w.Code == http.StatusOK {
					t.Fatal("Expected failure, got success")
				}
			}
		})
	}
}

func TestLibraryHandler_ReturnBookHandler(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		dbtype    string
		borroInfo collegemanagementsystem.BorrowInfo
		willpass  bool
	}{
		{
			name:   "mysql valid student return",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   1,
				UserType: "student",
				BookID:   1,
			},
			willpass: true,
		},
		{
			name:   "mongo valid student return",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   1,
				UserType: "student",
				BookID:   1,
			},
			willpass: true,
		},
		{
			name:   "mysql invalid userID",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid userID",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid usertype",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid usertype",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "",
				BookID:   1,
			},
			willpass: false,
		},
		{
			name:   "mysql invalid bookID",
			dbtype: "mysql",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   0,
			},
			willpass: false,
		},
		{
			name:   "mongo invalid bookID",
			dbtype: "mongo",
			borroInfo: collegemanagementsystem.BorrowInfo{
				UserID:   0,
				UserType: "student",
				BookID:   0,
			},
			willpass: false,
		},
	}
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	// Initialize Redis client
	redisClient, err := collegemanagementsystem.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MySQl
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		log.Fatal(err)
	}

	// Connectto  MongoDB
	mongodb, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}
	mysqlLibrary := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)
	mongoLibrary := collegemanagementsystem.NewMongoDBLibraryRepo(mongodb.Collection("libraries"), mongodb.Collection("borrow_records"))

	Libraryshandler := &collegemanagementsystem.LibraryHandler{MySQLRepo: mysqlLibrary, MongoRepo: mongoLibrary, Redis: redisClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mysqlDB.Exec("DELETE FROM students")
			mysqlDB.Exec("DELETE FROM lecturers")
			mysqlDB.Exec("DELETE FROM libraries")
			mysqlDB.Exec("DELETE FROM borrow_records")

			mongodb.Collection("students").DeleteMany(context.Background(), map[string]interface{}{})
			mongodb.Collection("lecturers").DeleteMany(context.Background(), map[string]interface{}{})
			mongodb.Collection("libraries").DeleteMany(context.Background(), map[string]interface{}{})
			mongodb.Collection("borrow_records").DeleteMany(context.Background(), map[string]interface{}{})

			redisClient.FlushAll(context.Background())

			// Insert users

			if tt.borroInfo.UserType == "student" {
				mysqlDB.Exec("INSERT INTO students(id,name,age,email,dept) VALUES(1,'Akash',22,'akash@gmail.com','CSE')")

				mongodb.Collection("students").InsertOne(context.Background(), map[string]interface{}{"id": 1, "name": "Akash", "age": 22, "email": "akash@gmail.com", "dept": "CSE"})
			} else if tt.borroInfo.UserType == "lecturer" {

				mysqlDB.Exec("INSERT INTO lecturers(id,name,age,email,designation) VALUES(1,'Abhi',30,'abhi@gmail.com','Prof')")

				mongodb.Collection("lecturers").InsertOne(context.Background(),
					map[string]interface{}{
						"id":          1,
						"name":        "Abhi",
						"age":         30,
						"email":       "abhi@gmail.com",
						"designation": "Prof",
					})
			}

			// Insert Books
			mysqlDB.Exec(
				"INSERT INTO libraries(book_id,book_name,title,author,available_copies) VALUES(1,'Go','GoLang','Alan',5)",
			)

			mongodb.Collection("libraries").InsertOne(context.Background(),
				map[string]interface{}{
					"book_id":          1,
					"book_name":        "Go",
					"title":            "GoLang",
					"author":           "Alan",
					"available_copies": 5,
				})

			// Borrow first
			mysqlDB.Exec("INSERT INTO borrow_records(user_id,user_type,book_id) VALUES(1,'student',1)")

			mongodb.Collection("borrow_records").InsertOne(context.Background(),
				map[string]interface{}{
					"user_id":     1,
					"user_type":   "student",
					"book_id":     1,
					"return_date": "",
				})
			data := tt.borroInfo

			Userbytes, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			buffer := bytes.NewBuffer(Userbytes)
			r := httptest.NewRequest(http.MethodPost, "/return?db="+tt.dbtype, buffer)
			w := httptest.NewRecorder()

			Libraryshandler.ReturnBookHandler(w, r)

			if tt.willpass {

				if w.Code != http.StatusOK {
					t.Fatalf("Expected 200, got %d", w.Code)
				}

			} else {

				if w.Code == http.StatusOK {
					t.Fatalf("Expected failure but got success")
				}
			}
		})
	}
}
