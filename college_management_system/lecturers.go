package collegemanagementsystem

import (
	"context"
	"database/sql"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Lecturer represents lecturer entity stored in DB
type Lecturer struct {
	ID          int    `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Age         int    `json:"age" bson:"age"`
	Email       string `json:"email" bson:"email"`
	Designation string `json:"designation" bson:"designation"`
}

// Repository Constructors

// NewMySQLLecturerRepo creates a new MySQLLecturerRepo instance
func NewMySQLLecturerRepo(db *sql.DB) LecturerRepository {
	return &MySQLLecturerRepo{DB: db}
}

// NewMongoDBLecturerRepo creates a new MongoDBLecturerRepo instance
func NewMongoDBLecturerRepo(col *mongo.Collection) LecturerRepository {
	return &MongoDBLecturerRepo{Collection: col}
}

// MySQL Implementation of LecturerRepository

// CreateLecturer inserts a new lecturer record into MySQL
func (m *MySQLLecturerRepo) CreateLecturer(l Lecturer) (*Lecturer, error) {
	res, err := m.DB.Exec("INSERT INTO lecturers(name , age , email , designation ) VALUES (? , ? , ? , ?)", l.Name, l.Age, l.Email, l.Designation)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	l.ID = int(id)

	return &l, nil
}

// GetAllLecturer retrieves all lecturer records from MySQL.
func (m *MySQLLecturerRepo) GetAllLecturer() ([]Lecturer, error) {

	rows, err := m.DB.Query(
		"SELECT id,name,age,email,designation FROM lecturers",
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var lecturers []Lecturer

	for rows.Next() {

		var l Lecturer

		rows.Scan(&l.ID, &l.Name, &l.Age, &l.Email, &l.Designation)

		lecturers = append(lecturers, l)
	}

	return lecturers, nil
}

// GetByIDLecturer retrieves a lecturer record by ID from MySQL.
func (m *MySQLLecturerRepo) GetByIDLecturer(id int) (*Lecturer, error) {

	row := m.DB.QueryRow(
		"SELECT id,name,age,email,designation FROM lecturers WHERE id=?",
		id,
	)

	var l Lecturer

	err := row.Scan(&l.ID, &l.Name, &l.Age, &l.Email, &l.Designation)

	if err != nil {
		return nil, err
	}

	return &l, nil
}

// UpdateLecturer updates an existing lecturer record in MySQL
func (m *MySQLLecturerRepo) UpdateLecturer(l Lecturer) error {

	_, err := m.DB.Exec(
		"UPDATE lecturers SET name=?,age=?,email=?,designation=? WHERE id=?",
		l.Name, l.Age, l.Email, l.Designation, l.ID,
	)

	return err
}

// DeleteLecturer removes a lecturer record from MySQL by ID
func (m *MySQLLecturerRepo) DeleteLecturer(id int) error {

	res, err := m.DB.Exec(
		"DELETE FROM lecturers WHERE id=?",
		id,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return errors.New("Lecturer not found")
	}

	return nil
}

// MongoDB Implementation of LecturerRepository

// CreateLecturer inserts a new lecturer document into MongoDB
func (m *MongoDBLecturerRepo) CreateLecturer(l Lecturer) (*Lecturer, error) {

	// Generate ID manually
	count, _ := m.Collection.CountDocuments(context.TODO(), bson.M{})
	l.ID = int(count) + 1

	_, err := m.Collection.InsertOne(context.TODO(), l)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// GetAllLecturer retrieves all lecturer documents from MongoDB.
func (m *MongoDBLecturerRepo) GetAllLecturer() ([]Lecturer, error) {
	cur, err := m.Collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	var lecturers []Lecturer

	for cur.Next(context.TODO()) {
		var l Lecturer
		cur.Decode(&l)

		lecturers = append(lecturers, l)

	}
	return lecturers, nil
}

// GetByIDLecturer retrieves a lecturer document by ID from MongoDB.
func (m *MongoDBLecturerRepo) GetByIDLecturer(id int) (*Lecturer, error) {
	var l Lecturer
	if err := m.Collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&l); err != nil {
		return nil, err
	}
	return &l, nil
}

// UpdateLecturer updates an existing lecturer document in MongoDB
func (m *MongoDBLecturerRepo) UpdateLecturer(l Lecturer) error {
	_, err := m.Collection.UpdateOne(context.TODO(), bson.M{"id": l.ID}, bson.M{"$set": l})
	return err
}

// DeleteLecturer removes a lecturer document from MongoDB by ID.
func (m *MongoDBLecturerRepo) DeleteLecturer(id int) error {

	res, err := m.Collection.DeleteOne(
		context.TODO(),
		bson.M{"id": id},
	)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("Lecturer not found ")
	}

	return nil
}
