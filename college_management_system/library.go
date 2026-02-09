package collegemanagementsystem

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// library represents a library entity
type Library struct {
	Book_id          int    `json:"book_id" bson:"book_id"`
	Book_name        string `json:"book_name" bson:"book_name"`
	Title            string `json:"title" bson:"title"`
	Author           string `json:"author" bson:"author"`
	Available_copies int    `json:"available_copies" bson:"available_copies"`
}

// Create struct to store one borrow record
type BorrowInfo struct {
	BorrowID   int    `json:"borrow_id" bson:"borrow_id"`
	UserID     int    `json:"user_id" bson:"user_id"`
	UserType   string `json:"user_type" bson:"user_type"`
	BookID     int    `json:"book_id" bson:"book_id"`
	BorrowDate string `json:"borrow_date" bson:"borrow_date"`
	ReturnDate string `json:"return_date" bson:"return_date"`
}

// Repository Constructors

// NewMySQLLibraryRepo creates a new MySQLLibraryRepo instance
func NewMySQLLibraryRepo(db *sql.DB) LibraryRepository {
	return &MySQLLibraryRepo{DB: db}
}

// NewMongoDBLibraryRepo creates a new MongoDBLibraryRepo instance
func NewMongoDBLibraryRepo(col *mongo.Collection, borrowcol *mongo.Collection) LibraryRepository {
	return &MongoDBLibraryRepo{Collection: col, BorrowCollection: borrowcol}
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

// BorrowBook borrow books from libraries
func (m *MySQLLibraryRepo) BorrowBook(info BorrowInfo) error {

	// Check available copies
	var copies int

	err := m.DB.QueryRow("SELECT available_copies FROM libraries WHERE book_id=?", info.BookID).Scan(&copies)
	if err != nil {
		return err
	}
	if copies <= 0 {
		return errors.New("book not available")
	}

	// Insert borrow records
	_, err = m.DB.Exec("INSERT INTO borrow_records(user_id , user_type, book_id , borrow_date) VALUES (? , ? , ? , NOW())", info.UserID, info.UserType, info.BookID)
	if err != nil {
		return err
	}

	//  Reduce available copies
	_, err = m.DB.Exec(
		"UPDATE libraries SET available_copies = available_copies - 1 WHERE book_id=?",
		info.BookID,
	)
	return err
}

// ReturnBook handles returning a borrowed book in MySQl
func (m *MySQLLibraryRepo) ReturnBook(info BorrowInfo) error {

	// Update borrow record
	res, err := m.DB.Exec("UPDATE borrow_records SET return_date=NOW() WHERE user_id=? AND book_id=? AND return_date IS NULL", info.UserID, info.BookID)
	if err != nil {
		return err
	}

	// Check how many rows are affected
	rows, _ := res.RowsAffected()

	if rows == 0 {
		return errors.New("Book not borrowed or already returned")
	}

	// Increase available copies in libraries
	_, err = m.DB.Exec("UPDATE libraries SET available_copies = available_copies + 1 WHERE book_id=?",
		info.BookID)
	return err
}

// BorrowBook handles borrowing a book in MongoDB
func (m *MongoDBLibraryRepo) BorrowBook(info BorrowInfo) error {
	ctx := context.TODO()

	// Fetch library document to check available copies
	var lib Library
	err := m.Collection.FindOne(ctx, bson.M{"book_id": info.BookID}).Decode(&lib)
	if err != nil {
		return err
	}

	// If no copies available → stop
	if lib.Available_copies <= 0 {
		return errors.New("book not available")
	}

	// Insert borrow records
	// Generate ID
	borrowID, err := m.getNextBorrowID()
	if err != nil {
		return err
	}

	info.BorrowID = borrowID
	info.BorrowDate = time.Now().Format(time.RFC3339)

	// Insert borrow record into borrow collection
	_, err = m.BorrowCollection.InsertOne(ctx, info)
	if err != nil {
		return err
	}

	// Decrease available copies in library
	_, err = m.Collection.UpdateOne(ctx, bson.M{"book_id": info.BookID}, bson.M{"$inc": bson.M{"available_copies": -1}})

	return err
}

// ReturnBook handles returning a borrowed book in MongoDB
func (m *MongoDBLibraryRepo) ReturnBook(info BorrowInfo) error {
	ctx := context.TODO()

	// Find borrow record
	filter := bson.M{"book_id": info.BookID, "return_date": ""}

	//  Update return_date with current timestamp
	update := bson.M{"$set": bson.M{"return_date": time.Now().Format(time.RFC3339)}}
	res, err := m.BorrowCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// If no document matched → already returned or never borrowed
	if res.MatchedCount == 0 {
		return errors.New("Book not borrowed or already returned")
	}

	// Increase available copies in library
	_, err = m.Collection.UpdateOne(ctx, bson.M{"book_id": info.BookID}, bson.M{"$inc": bson.M{"available_copies": 1}})

	return err
}

// getNextBorrowID generates a sequential borrow_id for MongoDB
func (m *MongoDBLibraryRepo) getNextBorrowID() (int, error) {

	ctx := context.TODO()

	// Identify the counter document
	filter := bson.M{"_id": "borrow_id"}

	// Increment sequence by 1
	update := bson.M{
		"$inc": bson.M{"seq": 1},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	//  Store updated counter value
	var result struct {
		Seq int `bson:"seq"`
	}

	err := m.BorrowCollection.
		Database().
		Collection("counters").
		FindOneAndUpdate(ctx, filter, update, opts).
		Decode(&result)

	if err != nil {
		return 0, err
	}

	// Return next borrow_id
	return result.Seq, nil
}
