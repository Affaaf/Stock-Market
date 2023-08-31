# Stock-Market-Simulation

This Git repository is a comprehensive stock market simulation application. It provides a set of RESTful API endpoints for managing users, stock data, and transactions, with the goal of simulating stock market activities.

## Table of Contents

- [Installation](#installation)
- [User Authentication](#UserAuthentication)
- [Postman Collection](#PostmanCollection)

### Clone Repository
```
git clone https://Affaaf:@github.com/Stock-Market
```

### Install Requirements
```
go get -d -v $(cat dependencies.txt)
```

### Postgres Connectivity
 Create .env and add port and database credentials in .env file 

### Database Migrations
```
go run migrate/migrate.go
``` 

### Run Server
```
./Go_Assignment
```

### Postman Collection

[Postman Collection](https://drive.google.com/file/d/1V3WC91Be8ZUNlEfCTIF2qulH14tT3CKt/view?usp=sharing)