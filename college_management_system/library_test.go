package collegemanagementsystem_test

import (
	collegemanagementsystem "college_management_system/college_management_system"
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupMySQLLibraryTestDB(t *testing.T) *sql.DB {

	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")

	db, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func clearLibraryTable(db *sql.DB) {

	// Delete child first
	_, err := db.Exec("DELETE FROM borrow_records")
	if err != nil {
		panic(err)
	}

	// Then parent
	_, err = db.Exec("DELETE FROM libraries")
	if err != nil {
		panic(err)
	}
}

func setupMongoLibraryTestDB(t *testing.T) (*mongo.Collection, *mongo.Collection) {

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db_test")

	client, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		t.Fatal(err)
	}

	db := client.Client().Database("college_db_test")

	libCol := db.Collection("libraries_test")
	borrowCol := db.Collection("borrow_test")

	return libCol, borrowCol
}

func ClearMongoLibrary(libCol, borrowCol *mongo.Collection) {

	_, err := libCol.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Println(err)
	}

	_, err = borrowCol.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Println(err)
	}
}

func TestMySQLLibraryRepo_CreateLibrary(t *testing.T) {
	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name     string // description of this test case
		library  collegemanagementsystem.Library
		willpass bool
	}{
		{
			name: "Valid library_book",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name: "inValid book_name",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid title",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid author",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid available_copies",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clearLibraryTable(db)

			result, err := repo.CreateLibrary(tt.library)
			if tt.willpass {

				if err != nil {
					t.Fatalf("Expected succes got error: %v", err)
				}
				if result.Book_id == 0 {
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

func TestMySQLLibraryRepo_GetAllLibrary(t *testing.T) {

	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name          string // description of this test case
		insertCount   int
		expectedCount int
	}{
		{
			name:          "empty library_book",
			insertCount:   0,
			expectedCount: 0,
		},
		{
			name:          "one library_book",
			insertCount:   1,
			expectedCount: 1,
		},
		{
			name:          "multiple library_book",
			insertCount:   3,
			expectedCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearLibraryTable(db)

			// Verify clear
			var count int
			db.QueryRow("SELECT COUNT(*) FROM libraries").Scan(&count)

			if count != 0 {
				t.Fatalf("table not cleared, has %d rows", count)
			}

			for i := 0; i < tt.insertCount; i++ {
				db.Exec("INSERT INTO libraries(book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)

			}
			library, err := repo.GetAllLibrary()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(library) != tt.expectedCount {
				t.Fatalf("Expected %d library_books got %d", tt.expectedCount, len(library))
			}
		})
	}
}

func TestMySQLLibraryRepo_GetByIDLibrary(t *testing.T) {

	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name     string // description of this test case
		book_id  int
		willpass bool
	}{
		{
			name:     "valid id",
			book_id:  1,
			willpass: true,
		},
		{
			name:     "invalid id",
			book_id:  0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearLibraryTable(db)

			var book_id int
			if tt.willpass {
				res, _ := db.Exec("INSERT INTO libraries(book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)
				lastID, _ := res.LastInsertId()
				book_id = int(lastID)
			} else {
				book_id = 9999
			}
			libraries, err := repo.GetByIDLibrary(book_id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Expected succes got error: %v", err)
				}
				if libraries.Book_id != book_id {
					t.Fatalf("Expected id %d , got %d", book_id, libraries.Book_id)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
			}
		})
	}
}

func TestMySQLLibraryRepo_UpdateLibrary(t *testing.T) {

	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name     string // description of this test case
		library  collegemanagementsystem.Library
		willpass bool
	}{
		{
			name: "Valid library_book",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name: "inValid book_name",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid title",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid author",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid available_copies",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clearLibraryTable(db)

			var book_id int

			if tt.willpass {
				res, _ := db.Exec("INSERT INTO libraries(book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)
				lastID, _ := res.LastInsertId()
				book_id = int(lastID)
			} else {
				book_id = 9999
			}
			tt.library.Book_id = book_id

			err := repo.UpdateLibrary(tt.library)

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

func TestMySQLLibraryRepo_DeleteLibrary(t *testing.T) {

	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name     string // description of this test case
		book_id  int
		willpass bool
	}{
		{
			name:     "valid ID",
			book_id:  1,
			willpass: true,
		},
		{
			name:     "Invalid ID",
			book_id:  0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearLibraryTable(db)

			var book_id int
			if tt.willpass {
				res, _ := db.Exec("INSERT INTO libraries(book_name , title , author , available_copies) VALUES (? , ? , ? , ?)", "comics", "The boys", "abhi", 10)
				lastID, _ := res.LastInsertId()
				book_id = int(lastID)
			} else {
				book_id = 9999
			}
			err := repo.DeleteLibrary(book_id)

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

func TestMongoDBLibraryRepo_CreateLibrary(t *testing.T) {

	libcol, borrowcol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libcol, borrowcol)

	tests := []struct {
		name     string // description of this test case
		library  collegemanagementsystem.Library
		willpass bool
	}{
		{
			name: "Valid library_book",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name: "inValid book_name",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid title",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid author",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid available_copies",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ClearMongoLibrary(libcol, borrowcol)

			res, err := repo.CreateLibrary(tt.library)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if res.Book_id == 0 {
					t.Fatal("expected id , got 0")
				}
			} else {
				if err == nil {
					t.Fatal("Expected error , got nil")
				}
			}

		})
	}
}

func TestMongoDBLibraryRepo_GetAllLibrary(t *testing.T) {

	libcol, borrowcol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libcol, borrowcol)

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

			ClearMongoLibrary(libcol, borrowcol)

			for i := 0; i < tt.insertCount; i++ {

				libcol.InsertOne(context.TODO(),
					bson.M{
						"book_id":          i + 1,
						"book_name":        "Comics",
						"title":            "The Boys",
						"author":           "Abhi",
						"available_copies": 5,
					},
				)
			}

			libs, err := repo.GetAllLibrary()

			if err != nil {
				t.Fatal(err)
			}

			if len(libs) != tt.expectedCount {
				t.Fatalf("expected %d got %d", tt.expectedCount, len(libs))
			}
		})
	}
}

func TestMongoDBLibraryRepo_GetByIDLibrary(t *testing.T) {

	libcol, borrowcol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libcol, borrowcol)

	tests := []struct {
		name     string // description of this test case
		book_id  int
		willpass bool
	}{
		{
			name:     "valid id",
			book_id:  1,
			willpass: true,
		},
		{
			name:     "invalid id",
			book_id:  0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ClearMongoLibrary(libcol, borrowcol)

			id := 1

			if tt.willpass {

				libcol.InsertOne(context.TODO(),
					bson.M{
						"book_id":          id,
						"book_name":        "Comics",
						"title":            "The Boys",
						"author":           "Abhi",
						"available_copies": 5,
					},
				)
			}

			lib, err := repo.GetByIDLibrary(id)

			if tt.willpass {

				if err != nil {
					t.Fatal(err)
				}

				if lib.Book_id != id {
					t.Fatal("wrong id")
				}

			} else {

				if err == nil {
					t.Fatal("expected error , got nil")
				}
			}
		})
	}
}

func TestMongoDBLibraryRepo_UpdateLibrary(t *testing.T) {

	libcol, borrowcol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libcol, borrowcol)

	tests := []struct {
		name     string // description of this test case
		library  collegemanagementsystem.Library
		willpass bool
	}{
		{
			name: "Valid library_book",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: true,
		},
		{
			name: "inValid book_name",
			library: collegemanagementsystem.Library{
				Book_name:        "",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid title",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "",
				Author:           "abhi",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid author",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "",
				Available_copies: 10,
			},
			willpass: false,
		},
		{
			name: "invalid available_copies",
			library: collegemanagementsystem.Library{
				Book_name:        "comics",
				Title:            "The boys",
				Author:           "abhi",
				Available_copies: 0,
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearMongoLibrary(libcol, borrowcol)
			var book_id int
			if tt.willpass {
				libcol.InsertOne(context.TODO(),
					bson.M{
						"book_id":          book_id,
						"book_name":        "Comics",
						"title":            "The Boys",
						"author":           "Abhi",
						"available_copies": 5,
					},
				)
			}
			tt.library.Book_id = book_id

			err := repo.UpdateLibrary(tt.library)

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

func TestMongoDBLibraryRepo_DeleteLibrary(t *testing.T) {

	libcol, borrowcol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libcol, borrowcol)

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
			ClearMongoLibrary(libcol, borrowcol)

			var book_id int
			if tt.willpass {
				libcol.InsertOne(context.TODO(),
					bson.M{
						"book_id":          book_id,
						"book_name":        "Comics",
						"title":            "The Boys",
						"author":           "Abhi",
						"available_copies": 5,
					},
				)
			}
			err := repo.DeleteLibrary(book_id)

			if tt.willpass {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error , got nil")
				}
			}
		})
	}
}

func TestMongoDBLibraryRepo_BorrowBook(t *testing.T) {

	libcol, borrowcol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libcol, borrowcol)

	tests := []struct {
		name     string // description of this test case
		copies   int
		willpass bool
	}{
		{
			name:     "available",
			copies:   2,
			willpass: true,
		},
		{
			name:     " not available",
			copies:   0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ClearMongoLibrary(libcol, borrowcol)

			id := 1

			// Insert library
			libcol.InsertOne(context.TODO(),
				bson.M{
					"book_id":          id,
					"book_name":        "Comics",
					"title":            "The Boys",
					"author":           "Abhi",
					"available_copies": tt.copies,
				},
			)

			info := collegemanagementsystem.BorrowInfo{
				BookID: id,
				UserID: 1,
			}

			err := repo.BorrowBook(info)

			if tt.willpass && err != nil {
				t.Fatal(err)
			}

			if !tt.willpass && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestMongoDBLibraryRepo_ReturnBook(t *testing.T) {

	libcol, borrowcol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libcol, borrowcol)

	tests := []struct {
		name     string // description of this test case
		borrowed bool
		willpass bool
	}{
		{
			name:     "available",
			borrowed: true,
			willpass: true,
		},
		{
			name:     " not available",
			borrowed: false,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearMongoLibrary(libcol, borrowcol)

			id := 1

			// Insert library
			libcol.InsertOne(context.TODO(),
				bson.M{
					"book_id":          id,
					"book_name":        "Comics",
					"title":            "The Boys",
					"author":           "Abhi",
					"available_copies": 1,
				},
			)

			// Insert borrow record if needed
			if tt.borrowed {

				borrowcol.InsertOne(context.TODO(),
					bson.M{
						"borrow_id":   1,
						"book_id":     id,
						"user_id":     1,
						"return_date": "",
						"borrow_date": time.Now().Format(time.RFC3339),
					},
				)
			}

			info := collegemanagementsystem.BorrowInfo{
				BookID: id,
			}

			err := repo.ReturnBook(info)

			if tt.willpass && err != nil {
				t.Fatal(err)
			}

			if !tt.willpass && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestMySQLLibraryRepo_BorrowBook(t *testing.T) {

	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name     string // description of this test case
		copies   int
		willpass bool
	}{
		{
			name:     "available",
			copies:   2,
			willpass: true,
		},
		{
			name:     " not available",
			copies:   0,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clearLibraryTable(db)

			id := 1

			// Insert library
			db.Exec("INSERT INTO libraries(book_id , book_name , title , author , available_copies) VALUES (? , ? , ? , ? , ?)", id, "comics", "The boys", "abhi", tt.copies)

			info := collegemanagementsystem.BorrowInfo{
				BookID: id,
				UserID: 1,
			}

			err := repo.BorrowBook(info)

			if tt.willpass && err != nil {
				t.Fatal(err)
			}

			if !tt.willpass && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestMySQLLibraryRepo_ReturnBook(t *testing.T) {

	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name     string // description of this test case
		borrowed bool
		willpass bool
	}{
		{
			name:     "return valid",
			borrowed: true,
			willpass: true,
		},
		{
			name:     " return invalid",
			borrowed: false,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clearLibraryTable(db)

			bookID := 1
			userID := 1

			// Insert library
			_, err := db.Exec("INSERT INTO libraries(book_id , book_name , title , author , available_copies) VALUES (? , ? , ? , ? , ?)", bookID, "comics", "The boys", "abhi", 1)
			if err != nil {
				t.Fatalf("insert library failed: %v", err)
			}

			// Insert borrow record if needed
			if tt.borrowed {

				_, err := db.Exec(`
					INSERT INTO borrow_records
					SET borrow_id = ?,
						book_id = ?,
						user_id = ?,
						user_type = ?,
						return_date = NULL,
						borrow_date = ?
				`,
					1, bookID, userID, "student", time.Now(),
				)

				if err != nil {
					t.Fatalf("insert borrow failed: %v", err)
				}

			}

			info := collegemanagementsystem.BorrowInfo{
				BookID: bookID,
				UserID: userID,
			}

			err = repo.ReturnBook(info)

			if tt.willpass && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tt.willpass && err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.willpass {

				var copies int
				err := db.QueryRow(
					"SELECT available_copies FROM libraries WHERE book_id=?",
					bookID,
				).Scan(&copies)
				if err != nil {
					t.Fatal(err)
				}

				if copies != 2 {
					t.Fatalf("expected copies 2, got %d", copies)
				}
			}

		})
	}
}

func TestMySQLLibraryRepo_CheckUserExists(t *testing.T) {

	db := setupMySQLLibraryTestDB(t)
	defer db.Close()

	repo := collegemanagementsystem.NewMySQLLibraryRepo(db)

	tests := []struct {
		name     string // description of this test case
		userType string
		insert   bool
		willpass bool
	}{
		{
			name:     "student exists",
			userType: "student",
			insert:   true,
			willpass: true,
		},
		{
			name:     "student not exists",
			userType: "student",
			insert:   false,
			willpass: false,
		},
		{
			name:     "lecturer exists",
			userType: "lecturer",
			insert:   true,
			willpass: true,
		},
		{
			name:     "lecturer not exists",
			userType: "lecturer",
			insert:   false,
			willpass: false,
		},
		{
			name:     "invalid user type",
			userType: "admin",
			insert:   false,
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db.Exec("DELETE FROM students")
			db.Exec("DELETE FROM lecturers")

			userID := 1

			if tt.insert {

				if tt.userType == "student" {
					_, err := db.Exec("INSERT INTO students (id , name , age , email, dept) VALUES (? , ? , ? , ? , ?)", userID, "Akash", 22, "akash@gmail.com", "CSE")
					if err != nil {
						t.Fatal(err)
					}
				}
				if tt.userType == "lecturer" {
					_, err := db.Exec("INSERT INTO lecturers (id , name , age , email, designation) VALUES (? , ? , ? , ? , ?)", userID, "Akash", 40, "akash@gmail.com", "HOD")
					if err != nil {
						t.Fatal(err)
					}
				}

			}

			exists, err := repo.CheckUserExists(userID, tt.userType)

			// invalid type case
			if tt.userType == "admin" {
				if err == nil {
					t.Fatalf("expected error for invalid user type")
				}
				return
			}
			// normal case
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if exists != tt.willpass {
				t.Fatalf("Expected %v , got %v", tt.willpass, exists)
			}
		})
	}
}

func TestMongoDBLibraryRepo_CheckUserExists(t *testing.T) {

	libCol, borrowCol := setupMongoLibraryTestDB(t)

	repo := collegemanagementsystem.NewMongoDBLibraryRepo(libCol, borrowCol)

	// Get DB
	db := libCol.Database()

	// Correct collections
	studentCol := db.Collection("students")
	lecturerCol := db.Collection("lecturers")

	tests := []struct {
		name     string // description of this test case
		userType string
		insert   bool
		willpass bool
	}{
		{
			name:     "student exists",
			userType: "student",
			insert:   true,
			willpass: true,
		},
		{
			name:     "student not exists",
			userType: "student",
			insert:   false,
			willpass: false,
		},
		{
			name:     "lecturer exists",
			userType: "lecturer",
			insert:   true,
			willpass: true,
		},
		{
			name:     "lecturer not exists",
			userType: "lecturer",
			insert:   false,
			willpass: false,
		},
		{
			name:     "invalid user type",
			userType: "admin",
			insert:   false,
			willpass: false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			ClearMongoLibrary(libCol, borrowCol)

			studentCol.DeleteMany(context.TODO(), bson.M{})
			lecturerCol.DeleteMany(context.TODO(), bson.M{})
			userID := 1

			if tt.insert {
				if tt.userType == "student" {
					_, err := studentCol.InsertOne(context.TODO(), bson.M{
						"id":    userID,
						"name":  "Akash",
						"age":   22,
						"email": "akash@gmail.com",
						"dept":  "CSE",
					})
					if err != nil {
						t.Fatal(err)
					}
				}
				if tt.userType == "lecturer" {
					_, err := lecturerCol.InsertOne(context.TODO(), bson.M{
						"id":          userID,
						"name":        "Akash",
						"age":         40,
						"email":       "akash@gmail.com",
						"designation": "HOD",
					})
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			exists, err := repo.CheckUserExists(userID, tt.userType)

			if tt.userType == "admin" {
				if err == nil {
					t.Fatal("Expected error for invalid user_type")
				}
				return
			}
			// Normal cases
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if exists != tt.willpass {
				t.Fatalf("Expected %v , got %v", tt.willpass, exists)
			}

		})
	}
}
