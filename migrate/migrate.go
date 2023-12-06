package main

import (
	"Go_Assignment/m/initializers"
	"Go_Assignment/m/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.ConnectToTestDb()

}

func main() {

	initializers.DB.AutoMigrate(&models.User{}, &models.Transaction{}, &models.StockData{})
}
