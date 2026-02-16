package collegemanagementsystem_test

import (
	collegemanagementsystem "college_management_system/college_management_system"
	"database/sql"
	"os"
	"testing"
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
