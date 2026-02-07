package collegemanagementsystem

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Student struct {
	ID    int    `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Age   int    `json:"age" bson:"age"`
	Email string `json:"email" bson:"email"`
	Dept  string `json:"dept" bson:"dept"`
}

func NewMySQLStudentRepo(db *sql.DB) StudentRepository {
	return &MySQLStudentRepo{DB: db}
}

func NewMongoDBStudentRepo(col *mongo.Collection) StudentRepository {
	return &MongoDBStudentRepo{Collection: col}
}

// Create Students
func (m *MySQLStudentRepo) CreateStudent(s Student) (*Student, error) {
	res, err := m.DB.Exec("INSERT INTO students(name , age , email , dept ) VALUES (? , ? , ? , ?)", s.Name, s.Age, s.Email, s.Dept)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	s.ID = int(id)

	return &s, nil
}

// Get Students
func (m *MySQLStudentRepo) GetAllStudent() ([]Student, error) {

	rows, err := m.DB.Query(
		"SELECT id,name,age,email,dept FROM students",
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var students []Student

	for rows.Next() {

		var s Student

		rows.Scan(&s.ID, &s.Name, &s.Age, &s.Email, &s.Dept)

		students = append(students, s)
	}

	return students, nil
}

/* READ BY ID */
func (m *MySQLStudentRepo) GetByIDStudent(id int) (*Student, error) {

	row := m.DB.QueryRow(
		"SELECT id,name,age,email,dept FROM students WHERE id=?",
		id,
	)

	var s Student

	err := row.Scan(&s.ID, &s.Name, &s.Age, &s.Email, &s.Dept)

	if err != nil {
		return nil, err
	}

	return &s, nil
}

/* UPDATE */
func (m *MySQLStudentRepo) UpdateStudent(s Student) error {

	_, err := m.DB.Exec(
		"UPDATE students SET name=?,age=?,email=?,dept=? WHERE id=?",
		s.Name, s.Age, s.Email, s.Dept, s.ID,
	)

	return err
}

/* DELETE */
func (m *MySQLStudentRepo) DeleteStudent(id int) error {

	_, err := m.DB.Exec(
		"DELETE FROM students WHERE id=?",
		id,
	)

	return err
}

// Students MongoDB CRUD operations

// Create Students
func (m *MongoDBStudentRepo) CreateStudent(s Student) (*Student, error) {

	// Generate ID manually
	count, _ := m.Collection.CountDocuments(context.TODO(), bson.M{})
	s.ID = int(count) + 1

	_, err := m.Collection.InsertOne(context.TODO(), s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Read all students
func (m *MongoDBStudentRepo) GetAllStudent() ([]Student, error) {
	cur, err := m.Collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	var students []Student

	for cur.Next(context.TODO()) {
		var s Student
		cur.Decode(&s)

		students = append(students, s)

	}
	return students, nil
}

// Read students byu ID
func (m *MongoDBStudentRepo) GetByIDStudent(id int) (*Student, error) {
	var s Student
	if err := m.Collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Update students By ID
func (m *MongoDBStudentRepo) UpdateStudent(s Student) error {
	_, err := m.Collection.UpdateOne(context.TODO(), bson.M{"id": s.ID}, bson.M{"$set": s})
	return err
}

// Delete students by ID
func (m *MongoDBStudentRepo) DeleteStudent(id int) error {

	_, err := m.Collection.DeleteOne(
		context.TODO(),
		bson.M{"id": id},
	)

	return err
}
