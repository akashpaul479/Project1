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
	"testing"

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
	redisClient := collegemanagementsystem.ConnectRedis()

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
	redisClient := collegemanagementsystem.ConnectRedis()

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
