package collegemanagementsystem_test

import (
	collegemanagementsystem "college_management_system/college_management_system"
	"context"
	"database/sql"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func SetupMySQLTestDB(t *testing.T) *sql.DB {

	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	db, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func ClearlecturerTable(db *sql.DB) {
	db.Exec("DELETE FROM lecturers")
}

func SetupMongoTestDB(t *testing.T) *mongo.Collection {

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")

	client, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		t.Fatal(err)
	}

	return client.Collection("lecturers")
}

func ClearMongoLecturer(col *mongo.Collection) {
	col.DeleteMany(context.TODO(), map[string]interface{}{})
}

func TestMySQLLecturerRepo_CreateLecturer(t *testing.T) {

	db := SetupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLecturerRepo(db)

	tests := []struct {
		name     string // description of this test case
		lecturer collegemanagementsystem.Lecturer
		willpass bool
	}{
		{
			name: "Valid student",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name: "invalid name or empty name",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid designation or empty designation",

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

			clearStudentsTable(db)

			result, err := repo.CreateLecturer(tt.lecturer)
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

func TestMySQLLecturerRepo_GetAllLecturer(t *testing.T) {

	db := SetupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLecturerRepo(db)

	tests := []struct {
		name          string // description of this test case
		insertCount   int
		expectedCount int
	}{
		{
			name:          "empty database",
			insertCount:   0,
			expectedCount: 0,
		},
		{
			name:          "one lecturer",
			insertCount:   1,
			expectedCount: 1,
		},
		{
			name:          "multiple lecturer",
			insertCount:   3,
			expectedCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ClearlecturerTable(db)

			for i := 0; i < tt.insertCount; i++ {
				db.Exec("INSERT INTO lecturers (name , age , email , designation) VALUES ('Akash', 22, 'akashpaul@gmail.com','HOD')")
			}
			lecturers, err := repo.GetAllLecturer()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(lecturers) != tt.expectedCount {
				t.Fatalf("Expected %d lecturers , got %d", tt.expectedCount, len(lecturers))
			}
		})
	}
}

func TestMySQLLecturerRepo_GetByIDLecturer(t *testing.T) {

	db := SetupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLecturerRepo(db)

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid lecturer",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid lecturer",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearlecturerTable(db)

			var id int
			if tt.willpass {
				res, _ := db.Exec("INSERT INTO lecturers (name , age , email , designation) VALUES ('Akash', 22, 'akashpaul@gmail.com','HOD')")
				LastID, _ := res.LastInsertId()
				id = int(LastID)
			}
			lecturers, err := repo.GetByIDLecturer(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if lecturers.ID != id {
					t.Fatalf("Expected id %d , got %d", id, lecturers.ID)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
			}

		})
	}
}

func TestMySQLLecturerRepo_UpdateLecturer(t *testing.T) {

	db := SetupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLecturerRepo(db)

	tests := []struct {
		name     string // description of this test case
		lecturer collegemanagementsystem.Lecturer
		willpass bool
	}{
		{
			name: "Valid student",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash paul",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name: "invalid name or empty name",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid designation or empty designation",

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

			ClearlecturerTable(db)

			var id int
			if tt.willpass {
				res, _ := db.Exec("INSERT INTO lecturers (name , age , email , designation) VALUES ('Akash', 22, 'akashpaul@gmail.com','HOD')")
				LastID, _ := res.LastInsertId()
				id = int(LastID)
			}

			tt.lecturer.ID = id

			err := repo.UpdateLecturer(tt.lecturer)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
			}
		})
	}
}

func TestMySQLLecturerRepo_DeleteLecturer(t *testing.T) {

	db := SetupMySQLTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLecturerRepo(db)

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid lecturer",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid lecturer",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ClearlecturerTable(db)

			var id int
			if tt.willpass {
				res, _ := db.Exec("INSERT INTO lecturers (name , age , email , designation) VALUES ('Akash', 22, 'akashpaul@gmail.com','HOD')")
				LastID, _ := res.LastInsertId()
				id = int(LastID)
			}
			err := repo.DeleteLecturer(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
			}

		})
	}
}

func TestMongoDBLecturerRepo_CreateLecturer(t *testing.T) {

	col := SetupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLecturerRepo(col)

	tests := []struct {
		name     string // description of this test case
		lecturer collegemanagementsystem.Lecturer
		willpass bool
	}{
		{
			name: "Valid student",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash paul",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name: "invalid name or empty name",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid designation or empty designation",

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
			ClearMongoLecturer(col)

			result, err := repo.CreateLecturer(tt.lecturer)

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

func TestMongoDBLecturerRepo_GetAllLecturer(t *testing.T) {

	col := SetupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLecturerRepo(col)

	tests := []struct {
		name          string // description of this test case
		insertCount   int
		expectedCount int
	}{
		{
			name:          "empty database",
			insertCount:   0,
			expectedCount: 0,
		},
		{
			name:          " one lecturer",
			insertCount:   1,
			expectedCount: 1,
		},
		{
			name:          " multiple lecturer",
			insertCount:   3,
			expectedCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ClearMongoLecturer(col)

			for i := 0; i < tt.insertCount; i++ {
				_, err := col.InsertOne(context.TODO(), collegemanagementsystem.Lecturer{ID: i + 1, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Designation: "HOD"})
				if err != nil {
					t.Fatalf("Insert failed: %v", err)
				}
			}
			lecturers, err := repo.GetAllLecturer()
			if err != nil {
				t.Fatalf("Unecpected error , got %d", err)
			}
			if len(lecturers) != tt.expectedCount {
				t.Fatalf("Expected %d lecturers , got %d", tt.expectedCount, len(lecturers))
			}
		})
	}
}

func TestMongoDBLecturerRepo_GetByIDLecturer(t *testing.T) {

	col := SetupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLecturerRepo(col)

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid lecturer",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid lecturer",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearMongoLecturer(col)

			var id int
			if tt.willpass {
				col.InsertOne(context.TODO(), collegemanagementsystem.Lecturer{ID: id, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Designation: "HOD"})
			} else {
				id = 9999
			}
			lecturers, err := repo.GetByIDLecturer(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if lecturers.ID != id {
					t.Fatalf("Expected id %d , got %d", id, lecturers.ID)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
			}
		})
	}
}

func TestMongoDBLecturerRepo_UpdateLecturer(t *testing.T) {

	col := SetupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLecturerRepo(col)

	tests := []struct {
		name     string // description of this test case
		lecturer collegemanagementsystem.Lecturer
		willpass bool
	}{
		{
			name: "Valid student",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash paul",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: true,
		},
		{
			name: "invalid name or empty name",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "",
				Age:         22,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid age",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         0,
				Email:       "akashpaul@gmail.com",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid email or empty email",
			lecturer: collegemanagementsystem.Lecturer{
				Name:        "Akash",
				Age:         22,
				Email:       "",
				Designation: "HOD",
			},
			willpass: false,
		},
		{
			name: "invalid designation or empty designation",

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

			ClearMongoLecturer(col)

			var id int
			if tt.willpass {
				col.InsertOne(context.TODO(), collegemanagementsystem.Lecturer{ID: id, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Designation: "HOD"})

			}
			tt.lecturer.ID = id

			err := repo.UpdateLecturer(tt.lecturer)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected err: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil ")
				}
			}
		})
	}
}

func TestMongoDBLecturerRepo_DeleteLecturer(t *testing.T) {

	col := SetupMongoTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLecturerRepo(col)

	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid lecturer",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid lecturer",
			id:       0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ClearMongoLecturer(col)

			var id int
			if tt.willpass {
				col.InsertOne(context.TODO(), collegemanagementsystem.Lecturer{ID: id, Name: "Akash", Age: 22, Email: "akashpaul@gmail.com", Designation: "HOD"})

			}

			err := repo.DeleteLecturer(id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected err: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil ")
				}
			}

		})
	}
}
