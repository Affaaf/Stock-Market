package initializers

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var TDB *gorm.DB

func ConnectToTestDb() {
	var err error
	dsn := os.Getenv("TESTING_DB_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to test databse ")
	}

}
