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

	db.Exec("SET FOREIGN_KEY_CHECKS=0")

	_, err := db.Exec("TRUNCATE TABLE libraries")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("TRUNCATE TABLE borrow_records")
	if err != nil {
		panic(err)
	}

	db.Exec("SET FOREIGN_KEY_CHECKS=1")
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
