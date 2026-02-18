package collegemanagementsystem_test

import (
	collegemanagementsystem "college_management_system/college_management_system"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestConnectMySQL(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		dsn      string
		willpass bool
	}{
		{
			name:     "valid dsn",
			dsn:      "root:root@tcp(localhost:3306)/management_system",
			willpass: true,
		},
		{
			name:     "invalid dsn",
			dsn:      "wrong:wrong@tcp(localhost:3306)/fake_db",
			willpass: false,
		},
		{
			name:     "empty dsn",
			dsn:      "",
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("MYSQL_DSN", tt.dsn)

			db, err := collegemanagementsystem.ConnectMySQL()
			if err != nil {
				if tt.willpass {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}
			if !tt.willpass {
				t.Fatal("Expected error , got nil")
			}
			if db == nil {
				t.Fatal("db is nil")
			}
			db.Close()
		})
	}
}

func TestConnectMongo(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		uri      string
		dbName   string
		willpass bool
	}{
		{
			name:     "valid mongo connection",
			uri:      "mongodb://localhost:27017",
			dbName:   "test_db",
			willpass: true,
		},
		{
			name:     "invalid uri",
			uri:      "mongodb://wronghost:27017",
			dbName:   "test_db",
			willpass: false,
		},
		{
			name:     "empty uri",
			uri:      "",
			dbName:   "test_db",
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("MONGO_URI", tt.uri)
			os.Setenv("MONGO_DB", tt.dbName)

			db, err := collegemanagementsystem.ConnectMongo()

			if err != nil {
				if tt.willpass {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}
			if !tt.willpass {
				t.Fatal("Expected error , got nil")
			}
			if db == nil {
				t.Fatal("db is nil")
			}
			if db.Name() != tt.dbName {
				t.Fatalf("Expected db %s , got %s", tt.dbName, db.Name())
			}
		})
	}
}

func TestConnectRedis(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		addr     string
		willpass bool
	}{
		{
			name:     "valid redis",
			addr:     "localhost:6379",
			willpass: true,
		},
		{
			name:     "invalid redis",
			addr:     "wronghost:6379",
			willpass: false,
		},
		{
			name:     "empty addr",
			addr:     "",
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("REDIS_ADDR", tt.addr)
			client, err := collegemanagementsystem.ConnectRedis()
			if err != nil {
				if tt.willpass {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}
			if !tt.willpass {
				t.Fatal("Expected error , got nil")
			}
			if client == nil {
				t.Fatal("client is nil")
			}
			client.Close()
		})
	}
}

func TestStudentHandler_GetRepo(t *testing.T) {

	// Setup env
	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")

	// Connect MySQL
	mysqlDB, err := collegemanagementsystem.ConnectMySQL()
	if err != nil {
		t.Fatalf("mysql connect failed: %v", err)
	}

	// Connect Mongo
	mongoDB, err := collegemanagementsystem.ConnectMongo()
	if err != nil {
		t.Fatalf("mongo connect failed: %v", err)
	}

	// Create real repos
	mysqlRepo := collegemanagementsystem.NewMySQLStudentRepo(mysqlDB)
	mongoRepo := collegemanagementsystem.NewMongoDBStudentRepo(
		mongoDB.Collection("students"),
	)

	handler := &collegemanagementsystem.StudentHandler{
		MySQLRepo: mysqlRepo,
		MongoRepo: mongoRepo,
	}
	tests := []struct {
		name   string // description of this test case
		query  string
		expect string
	}{
		{
			name:   "use mongo",
			query:  "?db=mongo",
			expect: "mongo",
		},
		{
			name:   "use mysql",
			query:  "?db=mysql",
			expect: "mysql",
		},
		{
			name:   "default mysql",
			query:  "",
			expect: "mysql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(
				http.MethodGet,
				"/students"+tt.query,
				nil,
			)

			repo := handler.GetRepo(r)

			switch tt.expect {

			case "mongo":
				if repo != mongoRepo {
					t.Fatal("expected mongo repo")
				}

			case "mysql":
				if repo != mysqlRepo {
					t.Fatal("expected mysql repo")
				}
			}
		})
	}
}

func TestLecturerHandler_GetRepo(t *testing.T) {

	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")

	mysqlDB, _ := collegemanagementsystem.ConnectMySQL()
	mongoDB, _ := collegemanagementsystem.ConnectMongo()

	mysqlRepo := collegemanagementsystem.NewMySQLLecturerRepo(mysqlDB)
	mongoRepo := collegemanagementsystem.NewMongoDBLecturerRepo(
		mongoDB.Collection("lecturers"),
	)

	handler := &collegemanagementsystem.LecturerHandler{
		MySQLRepo: mysqlRepo,
		MongoRepo: mongoRepo,
	}

	req1 := httptest.NewRequest("GET", "/lecturers?db=mongo", nil)
	req2 := httptest.NewRequest("GET", "/lecturers", nil)

	if handler.GetRepo(req1) != mongoRepo {
		t.Fatal("expected mongo repo")
	}

	if handler.GetRepo(req2) != mysqlRepo {
		t.Fatal("expected mysql repo")
	}

	tests := []struct {
		name   string // description of this test case
		query  string
		expect string
	}{
		{
			name:   "use mongo",
			query:  "?db=mongo",
			expect: "mongo",
		},
		{
			name:   "use mysql",
			query:  "?db=mysql",
			expect: "mysql",
		},
		{
			name:   "default mysql",
			query:  "",
			expect: "mysql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(
				http.MethodGet,
				"/students"+tt.query,
				nil,
			)

			repo := handler.GetRepo(r)

			switch tt.expect {

			case "mongo":
				if repo != mongoRepo {
					t.Fatal("expected mongo repo")
				}

			case "mysql":
				if repo != mysqlRepo {
					t.Fatal("expected mysql repo")
				}
			}

		})
	}
}

func TestLibraryHandler_GetRepo(t *testing.T) {

	os.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/management_system")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "college_db")

	mysqlDB, _ := collegemanagementsystem.ConnectMySQL()
	mongoDB, _ := collegemanagementsystem.ConnectMongo()

	mysqlRepo := collegemanagementsystem.NewMySQLLibraryRepo(mysqlDB)

	mongoRepo := collegemanagementsystem.NewMongoDBLibraryRepo(
		mongoDB.Collection("libraries"),
		mongoDB.Collection("borrow_records"),
	)

	handler := &collegemanagementsystem.LibraryHandler{
		MySQLRepo: mysqlRepo,
		MongoRepo: mongoRepo,
	}

	req := httptest.NewRequest("GET", "/library?db=mongo", nil)

	if handler.GetRepo(req) != mongoRepo {
		t.Fatal("expected mongo repo")
	}
	tests := []struct {
		name   string // description of this test case
		query  string
		expect string
	}{
		{
			name:   "use mongo",
			query:  "?db=mongo",
			expect: "mongo",
		},
		{
			name:   "use mysql",
			query:  "?db=mysql",
			expect: "mysql",
		},
		{
			name:   "default mysql",
			query:  "",
			expect: "mysql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(
				http.MethodGet,
				"/students"+tt.query,
				nil,
			)

			repo := handler.GetRepo(r)

			switch tt.expect {

			case "mongo":
				if repo != mongoRepo {
					t.Fatal("expected mongo repo")
				}

			case "mysql":
				if repo != mysqlRepo {
					t.Fatal("expected mysql repo")
				}
			}

		})
	}
}
