package collegemanagementsystem

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Student represents student entity stored in DB
type Student struct {
	ID    int    `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Age   int    `json:"age" bson:"age"`
	Email string `json:"email" bson:"email"`
	Dept  string `json:"dept" bson:"dept"`
}

// Repository Constructors

// NewMySQLStudentRepo creates a new MySQLStudentRepo instance
func NewMySQLStudentRepo(db *sql.DB) StudentRepository {
	return &MySQLStudentRepo{DB: db}
}

// NewMongoDBStudentRepo creates a new MongoDBStudentRepo instance
func NewMongoDBStudentRepo(col *mongo.Collection) StudentRepository {
	return &MongoDBStudentRepo{Collection: col}
}

// CreateStudent inserts student into MySQL database
func (m *MySQLStudentRepo) CreateStudent(s Student) (*Student, error) {

	// Execute SQL insert query
	res, err := m.DB.Exec("INSERT INTO students(name , age , email , dept ) VALUES (? , ? , ? , ?)", s.Name, s.Age, s.Email, s.Dept)
	if err != nil {
		return nil, err
	}

	// Get auto-generated ID
	id, _ := res.LastInsertId()
	s.ID = int(id)

	return &s, nil
}

// GetAllStudent returns all students from DB or cache
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

// GetByIDStudent retrieves a student record by ID from MySQL
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

// UpdateStudent updates an existing student record in MySQL
func (m *MySQLStudentRepo) UpdateStudent(s Student) error {

	_, err := m.DB.Exec(
		"UPDATE students SET name=?,age=?,email=?,dept=? WHERE id=?",
		s.Name, s.Age, s.Email, s.Dept, s.ID,
	)

	return err
}

// DeleteStudent removes a student record from MySQL by ID.
func (m *MySQLStudentRepo) DeleteStudent(id int) error {

	_, err := m.DB.Exec(
		"DELETE FROM students WHERE id=?",
		id,
	)

	return err
}

// MongoDB Implementation of StudentRepository

// CreateStudent inserts a new student document into MongoDB.
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

// GetAllStudent retrieves all student documents from MongoDB
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

// GetByIDStudent retrieves a student document by ID from MongoDB
func (m *MongoDBStudentRepo) GetByIDStudent(id int) (*Student, error) {
	var s Student
	if err := m.Collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateStudent updates an existing student document in MongoDB
func (m *MongoDBStudentRepo) UpdateStudent(s Student) error {
	_, err := m.Collection.UpdateOne(context.TODO(), bson.M{"id": s.ID}, bson.M{"$set": s})
	return err
}

// DeleteStudent removes a student document from MongoDB by ID
func (m *MongoDBStudentRepo) DeleteStudent(id int) error {

	_, err := m.Collection.DeleteOne(
		context.TODO(),
		bson.M{"id": id},
	)

	return err
}
