# College Management system API   
A RESTful College Management system built using **Go (Golang)** with support for:  
- MySQL & MongoDB (Multiple Databases)  
- Redis caching  
- JWT Authentication (Access & Refresh Tokens)  
- Swagger API Documentation  
- CRUD for Students, Lecturers, Library  
- Borrow & Return Books System  
***

## Features  
- User Authentication using JWT  
- Secure APIs with Middleware  
- Database Switching  (MySQL/MongoDB)  
- Redis Cache for Fast Responses  
- Auto-generated Swagger Documentation  
- Modular Architecture (Handler, Repo, Model)  
- RESTful API Design  
***
## Tech Stack  
 
| Technology  | Usage               |  
| ----------- | ------------------- |  
| Go          | Backend Language    |  
| Gorilla Mux | HTTP Router         |  
| MySQL       | Relational Database |  
| MongoDB     | NoSQL Database      |  
| Redis       | Caching             |  
| JWT         | Authentication      |  
| Swagger     | API Documentation   |  
***

## Project Structure  
Project1/  
│  
├── college_management_system/  
│ ├── handlers.go  
│ ├── main.go  
│ ├── jwt.go  
│ ├── student.go  
│ ├── lecturer.go  
│ ├── library.go  
│
├── docs/ (Swagger Generated)  
│
├── .env  
├── go.mod  
├── go.sum  
└── README.md  
***
## Environment Variables (.env File)  
### Create a .env file:  
```.env
#Login Credentials
EMAIL=admin@gmail.com
PASSWORD=Admin@123

#JWT
JWT_SECRET=mysecretkey

# MySQL
MYSQL_DSN=root:root@tcp(localhost:3306)/db_name

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DB=db_name

# Redis 
REDIS_ADDR=localhost:6379   
```
***
## Start databases  
Make sure these are running  
- MySQL  
- MongoDB  
- Redis  
### Example:  
```bash
sudo service mysql start  
sudo service redis start  
mongod
```
Windows: Start from Services / MongoDB Compass)
*** 
## How to Run My Project  
### Step 1: Clone  
```bash
git clone <your-repo-url>
cd <repo_name>
```
### Step 2: Install Packages  
```bash
go mod tidy
```

### step 3: Swagger Setup  
#### Install Swagger Tool (One Time)  
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```
### step 4: Generate Swagger Docs  
#### From Project root:  
```bash
swag init
```
#### This will create: 
```bash
docs/
  ├── docs.go
  ├── swagger.json
  └── swagger.yaml
```
### step 5: Run Server  
```bash
go run main.go  
```
### Server runs on:  
```bash
http://localhost:8080
```
***
## Open Swagger UI  
#### Open in browser:  
```bash
http://localhost:8080/swagger/index.html
```
- Here I can test APIs without Postman.  
***
# Authentication System (JWT)  
### My project uses:   
- Access Token (15 min)  
- Refresh Token (7 days)  
- Stored in Cookies  
 ## Login API
```bash
POST /login
```
### Example:  
```bash
{
  "email": "admin@gmail.com",
  "password": "admin123"
}
```
After login → cookies are set → protected APIs work.  
***
# API Endpoints  
### Authentication  
| Method | URL      | Work      |  
| ------ | -------- | --------- |  
| POST   | /login   | Login     |  
| POST   | /refresh | New Token |  
| POST   | /logout  | Logout    |  

### Students  
| Method | Endpoint           | Description    |  
| ------ | ------------------ | -------------- |  
| POST   | /api/students      | Create Student |  
| GET    | /api/students      | Get All        |  
| GET    | /api/students/{id} | Get By ID      |  
| PUT    | /api/students/{id} | Update         |  
| DELETE | /api/students/{id} | Delete         |  

### Lecturers

| Method | Endpoint            | Description |  
| ------ | ------------------- | ----------- |  
| POST   | /api/lecturers      | Create      |  
| GET    | /api/lecturers      | Get All     |  
| GET    | /api/lecturers/{id} | Get By ID   |  
| PUT    | /api/lecturers/{id} | Update      |  
| DELETE | /api/lecturers/{id} | Delete      |  


### Library  
| Method | Endpoint            | Description |   
| ------ | ------------------- | ----------- |   
| POST   | /api/libraries      | Create      |   
| GET    | /api/libraries      | Get All     |   
| GET    | /api/libraries/{id} | Get By ID   |   
| PUT    | /api/libraries/{id} | Update      |  
| DELETE | /api/libraries/{id} | Delete      |    

### Borrow & Return  
| Method | Endpoint    | Description |  
| ------ | ----------- | ----------- |  
| POST   | /api/borrow | Borrow Book |  
| POST   | /api/return | Return Book |  
***

## Database Selection  
You can switch databases using query param:  
```bash
?db=mysql
?db=mongo
```
#### Example:  
```bash
GET /api/students?db=mongo
```
Default: MyYSQL  
***

## Redis Caching  
I use Redis to cache:  
- Students  
- Lecturers  
- Library  
### Flow:  
```bash
Client → Redis → MySQL(or)MongoDB → Redis → Client
```
TTL(Time to live) = 60 seconds  
This improves speed.  
***

# Logging & Audit  
#### My project logs actions:  
### Example:  
```bash
[LOG] CREATE_STUDENT
[AUDIT] CREATE STUDENT 5
```
This helps in debugging.  
***

## Testing with Curl  
### JWT Authentication And Authorization
#### Login  
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"email\":\"admin@gmail.com\",\"password\":\"admin123\"}" ^
 http://localhost:8080/login -c cookies.txt
```
#### Refresh 
```bash 
curl -X POST http://localhost:8080/refresh -b cookies.txt
```
#### Logout  
```bash
curl -X POST http://localhost:8080/logout -b cookies.txt
```

## Students CRUD Operations
### Create Students  
```bash
curl -X POST -H "Content-Type: application/json" ^
 -d "{\"name\":\"example\",\"age\":45,\"email\":\"example@gmail.com\",\"dept\":\"CSE\"}" ^  
 http://localhost:8080/api/students?db=mysql -b cookies.txt  
```
### Get all Students  
```bash
curl http://localhost:8080/api/students?db=mysql -b cookies.txt
```
### Get students by ID  
```bash
curl http://localhost:8080/api/students/1?db=mysql -b cookies.txt  
```
### Update Students  
```bash
curl -X PUT -H "Content-Type: application/json" ^
-d "{\"name\":\"john\",\"age\":50,\"email\":\"john@gmail.com\",\"dept\":\"ECE\"}" ^
http://localhost:8080/api/students/1?db=mysql -b cookies.txt
```
### Delete Students  
```bash
curl -X DELETE http://localhost:8080/api/students/1?db=mysql -b cookies.txt
```
- To perform the same CRUD operations on MongoDB, replace db=mysql with db=mongo in the request URL.  

## Lecturers CRUD Operations
### Create Lecturers 
```bash
curl -X POST -H "Content-Type: application/json" ^
 -d "{\"name\":\"example\",\"age\":45,\"email\":\"example@gmail.com\",\"designation\":\"Professor\"}" ^  
 http://localhost:8080/api/lecturers?db=mysql -b cookies.txt  
```
### Get all Lecturers  
```bash
curl http://localhost:8080/api/lecturers?db=mysql  -b cookies.txt
```
### Get Lecturers by ID  
```bash
curl http://localhost:8080/api/lecturers/1?db=mysql -b cookies.txt  
```
### Update Lecturers 
```bash
curl -X PUT -H "Content-Type: application/json" ^
-d "{\"name\":\"john\",\"age\":50,\"email\":\"john@gmail.com\",\"designation\":\"HOD\"}" ^
http://localhost:8080/api/lecturers/1?db=mysql -b cookies.txt
```
### Delete Lecturers  
```bash
curl -X DELETE http://localhost:8080/api/lecturers/1?db=mysql -b cookies.txt
```
- To perform the same CRUD operations on MongoDB, replace db=mysql with db=mongo in the request URL.


## Library CRUD Operations
### Create Library book
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"book_name\":\"The Guide\",\"title\":\"tourist guide\",\"author\":\" R.K. Narayan\",\"available_copies\":10}" ^
http://localhost:8080/api/libraries?db=mysql -b cookies.txt
```
### Get all Library books 
```bash
curl http://localhost:8080/api/libraries?db=mysql -b cookies.txt
```
### Get Library book by ID  
```bash
curl http://localhost:8080/api/libraries/1?db=mysql -b cookies.txt  
```
### Update Library book
```bash
curl -X PUT -H "Content-Type: application/json" ^
-d "{\"book_name\":\"The boys\",\"title\":\"boys\",\"author\":\" john\",\"available_copies\":15}" ^
http://localhost:8080/api/libraries/1?db=mysql -b cookies.txt
```
### Delete Library 
```bash
curl -X DELETE http://localhost:8080/api/libraries/1?db=mysql -b cookies.txt
```
## Borrow_Records
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"user_id\":1,\"user_type\":\"student\",\"book_id\":1}" ^
http://localhost:8080/api/borrow?db=mysql -b cookies.txt
```
## Return_Records  
```bash
curl -X POST -H "Content-Type: application/json" ^
-d "{\"user_id\":1,\"user_type\":\"student\",\"book_id\":1}" ^
http://localhost:8080/api/return?db=mysql -b cookies.txt
```
- To perform the same CRUD operations on MongoDB, replace db=mysql with db=mongo in the request URL.
***
## Status Code   
| Range | Meaning         | Example     |  
| ----- | --------------- | ----------- |  
| 1xx   | Info            | Rare        |  
| 2xx   | Success         | 200, 201    |  
| 3xx   | Redirect        | Rare in API |  
| 4xx   | Client Error    | 400, 401    |  
| 5xx   | Server Error    | 500         |  
***
# Contributions  
Contributions are Welcome!  
- Fork the repository  
- Create a future branch  
- Commit changes  
- Push and open a pull Request  
***
# License  
This project is licensed under MIT License.  
***
# Thanks  
Thank you for visiting the College_Management_System repository. Feel free to reach out it you want help setting up, extending, or deploying this project!
