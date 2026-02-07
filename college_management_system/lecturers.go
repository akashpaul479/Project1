package collegemanagementsystem

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Lecturer struct {
	ID          int    `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Age         int    `json:"age" bson:"age"`
	Email       string `json:"email" bson:"email"`
	Designation string `json:"designation" bson:"designation"`
}

func NewMySQLLecturerRepo(db *sql.DB) LecturerRepository {
	return &MySQLLecturerRepo{DB: db}
}

func NewMongoDBLecturerRepo(col *mongo.Collection) LecturerRepository {
	return &MongoDBLecturerRepo{Collection: col}
}

// Create Students
func (m *MySQLLecturerRepo) CreateLecturer(l Lecturer) (*Lecturer, error) {
	res, err := m.DB.Exec("INSERT INTO lecturers(name , age , email , designation ) VALUES (? , ? , ? , ?)", l.Name, l.Age, l.Email, l.Designation)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	l.ID = int(id)

	return &l, nil
}

// Get Students
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

/* READ BY ID */
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

/* UPDATE */
func (m *MySQLLecturerRepo) UpdateLecturer(l Lecturer) error {

	_, err := m.DB.Exec(
		"UPDATE lecturers SET name=?,age=?,email=?,designation=? WHERE id=?",
		l.Name, l.Age, l.Email, l.Designation, l.ID,
	)

	return err
}

/* DELETE */
func (m *MySQLLecturerRepo) DeleteLecturer(id int) error {

	_, err := m.DB.Exec(
		"DELETE FROM lecturers WHERE id=?",
		id,
	)

	return err
}

// Students MongoDB CRUD operations

// Create Students
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

// Read all students
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

// Read students byu ID
func (m *MongoDBLecturerRepo) GetByIDLecturer(id int) (*Lecturer, error) {
	var l Lecturer
	if err := m.Collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&l); err != nil {
		return nil, err
	}
	return &l, nil
}

// Update students By ID
func (m *MongoDBLecturerRepo) UpdateLecturer(l Lecturer) error {
	_, err := m.Collection.UpdateOne(context.TODO(), bson.M{"id": l.ID}, bson.M{"$set": l})
	return err
}

// Delete students by ID
func (m *MongoDBLecturerRepo) DeleteLecturer(id int) error {

	_, err := m.Collection.DeleteOne(
		context.TODO(),
		bson.M{"id": id},
	)

	return err
}
