package collegemanagementsystem_test

import (
	collegemanagementsystem "college_management_system/college_management_system"
	"context"
	"database/sql"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func setupMySQLTestDB(t *testing.T) *sql.DB {

	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	db, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func clearStudentsTable(db *sql.DB) {
	db.Exec("DELETE FROM students")
}
func setupMongoTestDB(t *testing.T) *mongo.Collection {

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")

	client, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		t.Fatal(err)
	}

	return client.Collection("students")
}

func clearMongoStudents(col *mongo.Collection) {
	col.DeleteMany(context.TODO(), map[string]interface{}{})
}

func TestMySQLStudentRepo_CreateStudent(t *testing.T) {

	db := setupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLStudentRepo(db)

	tests := []struct {
		name     string // description of this test case
		student  collegemanagementsystem.Student
		willpass bool
	}{
		{
			name: "Valid student",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name: "invalid name or empty name",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid dept or empty dept",
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

			clearStudentsTable(db)

			result, err := repo.CreateStudent(tt.student)
			if tt.willpass {

				if err != nil {
					t.Fatalf("Expected succes got error: %v", err)
				}
				if result.ID == 0 {
					t.Fatal("Expected ID , got 0")
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
			}
		})
	}
}

func TestMySQLStudentRepo_GetAllStudent(t *testing.T) {

	db := setupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLStudentRepo(db)

	tests := []struct {
		name          string // description of this test case
		insertCount   int
		expectedCount int
	}{
		{
			name:          "Empty database",
			insertCount:   0,
			expectedCount: 0,
		},
		{
			name:          "one student",
			insertCount:   1,
			expectedCount: 1,
		},
		{
			name:          "multiple student",
			insertCount:   3,
			expectedCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearStudentsTable(db)

			for i := 0; i < tt.insertCount; i++ {
				db.Exec("INSERT INTO students (name , age , email , dept) VALUES ('Akash', 22, 'akashpaul@gmail.com','CSE')")

			}
			students, err := repo.GetAllStudent()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(students) != tt.expectedCount {
				t.Fatalf("Expected %d students got %d", tt.expectedCount, len(students))
			}
		})
	}
}

func TestMySQLStudentRepo_GetByIDStudent(t *testing.T) {

	db := setupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLStudentRepo(db)

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid id",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid id",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearStudentsTable(db)

			var id int
			if tt.willpass {
				res, _ := db.Exec("INSERT INTO students (name , age , email , dept) VALUES ('Akash', 22 , 'akashpaul@gmail.com','CSE')")
				lastID, _ := res.LastInsertId()
				id = int(lastID)
			} else {
				id = 9999
			}
			student, err := repo.GetByIDStudent(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Expected succes got error: %v", err)
				}
				if student.ID != id {
					t.Fatalf("Expected id %d , got %d", id, student.ID)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
			}
		})
	}
}

func TestMySQLStudentRepo_UpdateStudent(t *testing.T) {

	db := setupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLStudentRepo(db)

	tests := []struct {
		name     string // description of this test case
		student  collegemanagementsystem.Student
		willpass bool
	}{
		{
			name: "valid student",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name: "invalid name",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid dept or empty dept",
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
			clearStudentsTable(db)
			var id int

			if tt.willpass {
				res, _ := db.Exec("INSERT INTO students (name , age , email , dept) VALUES ('Akash', 22 , 'akashpaul@gmail.com','CSE')")
				lastID, _ := res.LastInsertId()
				id = int(lastID)
			} else {
				id = 9999
			}
			tt.student.ID = id

			err := repo.UpdateStudent(tt.student)

			if tt.willpass {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
			}
		})
	}
}

func TestMySQLStudentRepo_DeleteStudent(t *testing.T) {

	db := setupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLStudentRepo(db)

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid ID",
			id:       1,
			willpass: true,
		},
		{
			name:     "Invalid ID",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clearStudentsTable(db)

			var id int
			if tt.willpass {
				res, _ := db.Exec("INSERT INTO students (name , age , email , dept) VALUES ('Akash', 22 , 'akashpaul@gmail.com','CSE')")
				lastID, _ := res.LastInsertId()
				id = int(lastID)
			} else {
				id = 9999
			}
			err := repo.DeleteStudent(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
			}

		})
	}
}

func TestMongoDBStudentRepo_CreateStudent(t *testing.T) {

	col := setupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBStudentRepo(col)

	tests := []struct {
		name     string // description of this test case
		student  collegemanagementsystem.Student
		willpass bool
	}{
		{
			name: "valid student",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name: "invalid name",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid dept or empty dept",
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

			clearMongoStudents(col)

			result, err := repo.CreateStudent(tt.student)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Expected success got error: %v", err)
				}

				if result.ID == 0 {
					t.Fatalf("Expected Id , got 0")
				}

			} else {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			}
		})
	}
}

func TestMongoDBStudentRepo_GetAllStudent(t *testing.T) {

	col := setupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBStudentRepo(col)

	tests := []struct {
		name          string // description of this test case
		insertCount   int
		expectedCount int
	}{
		{
			name:          "Empty database",
			insertCount:   0,
			expectedCount: 0,
		},
		{
			name:          "one student",
			insertCount:   1,
			expectedCount: 1,
		},
		{
			name:          "multiple student",
			insertCount:   3,
			expectedCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clearMongoStudents(col)

			for i := 0; i < tt.insertCount; i++ {
				_, err := col.InsertOne(context.TODO(), collegemanagementsystem.Student{ID: i + 1, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Dept: "CSE"})
				if err != nil {
					t.Fatalf("insert failed: %v", err)
				}
			}

			students, err := repo.GetAllStudent()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(students) != tt.expectedCount {
				t.Fatalf("Expected %d students got %d", tt.expectedCount, len(students))
			}

		})
	}
}

func TestMongoDBStudentRepo_GetByIDStudent(t *testing.T) {

	col := setupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBStudentRepo(col)

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id       int
		willpass bool
	}{
		{
			name:     "valid id",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid id",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clearMongoStudents(col)

			var id int
			if tt.willpass {
				col.InsertOne(context.TODO(), collegemanagementsystem.Student{ID: id, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Dept: "CSE"})
			} else {
				id = 9999
			}
			student, err := repo.GetByIDStudent(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Expected succes got error: %v", err)
				}
				if student.ID != id {
					t.Fatalf("Expected id %d , got %d", id, student.ID)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
			}

		})
	}
}

func TestMongoDBStudentRepo_UpdateStudent(t *testing.T) {

	col := setupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBStudentRepo(col)

	tests := []struct {
		name     string // description of this test case
		student  collegemanagementsystem.Student
		willpass bool
	}{
		{
			name: "valid student",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: true,
		},
		{
			name: "invalid name",
			student: collegemanagementsystem.Student{
				Name:  "",
				Age:   22,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   0,
				Email: "akashpaul@gmail.com",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			student: collegemanagementsystem.Student{
				Name:  "Akash",
				Age:   22,
				Email: "",
				Dept:  "CSE",
			},
			willpass: false,
		},
		{
			name: "invalid dept or empty dept",
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
			clearMongoStudents(col)
			var id int

			if tt.willpass {
				col.InsertOne(context.TODO(), collegemanagementsystem.Student{ID: id, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Dept: "CSE"})
			} else {
				id = 9999
			}
			tt.student.ID = id

			err := repo.UpdateStudent(tt.student)

			if tt.willpass {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
			}

		})
	}
}

func TestMongoDBStudentRepo_DeleteStudent(t *testing.T) {

	col := setupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBStudentRepo(col)

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid ID",
			id:       1,
			willpass: true,
		},
		{
			name:     "Invalid ID",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMongoStudents(col)
			var id int
			if tt.willpass {
				col.InsertOne(context.TODO(), collegemanagementsystem.Student{ID: id, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Dept: "CSE"})
			} else {
				id = 9999
			}
			err := repo.DeleteStudent(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
			}
		})
	}
}
