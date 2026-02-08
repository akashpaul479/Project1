package collegemanagementsystem

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// library represents a library entity
type Library struct {
	Book_id          int    `json:"book_id" bson:"book_id"`
	Book_name        string `json:"book_name" bson:"book_name"`
	Title            string `json:"title" bson:"title"`
	Author           string `json:"author" bson:"author"`
	Available_copies int    `json:"available_copies" bson:"available_copies"`
}

// Repository Constructors

// NewMySQLStudentRepo creates a new MySQLStudentRepo instance
func NewMySQLLibraryRepo(db *sql.DB) LibraryRepository {
	return &MySQLLibraryRepo{DB: db}
}

// NewMongoDBStudentRepo creates a new MongoDBStudentRepo instance
func NewMongoDBLibraryRepo(col *mongo.Collection) LibraryRepository {
	return &MongoDBLibraryRepo{Collection: col}
}

// CreateLibrary inserts library into MySQL database
func (m *MySQLLibraryRepo) CreateLibrary(l Library) (*Library, error) {

	// Execute SQL insert query
	res, err := m.DB.Exec("INSERT INTO libraries(book_name , title , author , available_copies) VALUES ( ? , ? , ? , ?)", l.Book_name, l.Title, l.Author, l.Available_copies)
	if err != nil {
		return nil, err
	}

	// Get auto-generated ID
	id, _ := res.LastInsertId()
	l.Book_id = int(id)

	return &l, nil
}

// GetAllLibrary returns all libraries from DB or cache
func (m *MySQLLibraryRepo) GetAllLibrary() ([]Library, error) {

	rows, err := m.DB.Query(
		"SELECT book_id , book_name , title , author , available_copies FROM libraries",
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var libraries []Library

	for rows.Next() {

		var l Library

		err := rows.Scan(&l.Book_id, &l.Book_name, &l.Title, &l.Author, &l.Available_copies)
		if err != nil {
			return nil, err
		}

		libraries = append(libraries, l)
	}

	return libraries, nil
}

// GetByIDLibrary retrieves a Libraries record by ID from MySQL
func (m *MySQLLibraryRepo) GetByIDLibrary(id int) (*Library, error) {

	row := m.DB.QueryRow(
		"SELECT book_id , book_name , title , author , available_copies FROM libraries WHERE book_id=?",
		id,
	)

	var l Library

	err := row.Scan(&l.Book_id, &l.Book_name, &l.Title, &l.Author, &l.Available_copies)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

// UpdateLibrary updates an existing libraries record in MySQL
func (m *MySQLLibraryRepo) UpdateLibrary(l Library) error {

	_, err := m.DB.Exec(
		"UPDATE libraries SET book_name=?,title=?,author=?,available_copies=? WHERE book_id=?",
		l.Book_name, l.Title, l.Author, l.Available_copies, l.Book_id,
	)

	return err
}

// DeleteLibrary removes a Libraries record from MySQL by ID.
func (m *MySQLLibraryRepo) DeleteLibrary(id int) error {

	_, err := m.DB.Exec(
		"DELETE FROM libraries WHERE book_id=?",
		id,
	)

	return err
}

// MongoDB Implementation of LibraryRepository

// CreateLibrary inserts a new library document into MongoDB.
func (m *MongoDBLibraryRepo) CreateLibrary(l Library) (*Library, error) {

	// Generate ID manually
	count, _ := m.Collection.CountDocuments(context.TODO(), bson.M{})
	l.Book_id = int(count) + 1

	_, err := m.Collection.InsertOne(context.TODO(), l)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// GetAllLibrary retrieves all libraries documents from MongoDB
func (m *MongoDBLibraryRepo) GetAllLibrary() ([]Library, error) {
	cur, err := m.Collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	var libraries []Library

	for cur.Next(context.TODO()) {
		var l Library
		cur.Decode(&l)

		libraries = append(libraries, l)

	}
	return libraries, nil
}

// GetByIDLibrary retrieves a Libraries document by ID from MongoDB
func (m *MongoDBLibraryRepo) GetByIDLibrary(id int) (*Library, error) {
	var l Library
	if err := m.Collection.FindOne(context.TODO(), bson.M{"book_id": id}).Decode(&l); err != nil {
		return nil, err
	}
	return &l, nil
}

// UpdateLibrary updates an existing libraries document in MongoDB
func (m *MongoDBLibraryRepo) UpdateLibrary(l Library) error {
	_, err := m.Collection.UpdateOne(context.TODO(), bson.M{"book_id": l.Book_id}, bson.M{"$set": l})
	return err
}

// DeleteLibrary removes a libraries document from MongoDB by ID
func (m *MongoDBLibraryRepo) DeleteLibrary(id int) error {

	_, err := m.Collection.DeleteOne(
		context.TODO(),
		bson.M{"book_id": id},
	)

	return err
}
