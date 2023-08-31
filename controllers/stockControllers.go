package controllers

import (
	"Go_Assignment/m/initializers"
	"Go_Assignment/m/models"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var wg = sync.WaitGroup{}

func Signup(c *gin.Context) {

	type data struct {
		Username string
		Email    string
		Balance  float64
		Password string
	}

	var requestBody data
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		// Handle the error (e.g., return an error response)
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Check if username already exists
	var existingUser models.User
	if err := initializers.DB.Where("username = ?", requestBody.Username).First(&existingUser).Error; err == nil {
		c.JSON(409, gin.H{"error": "Username already exists"})
		return
	}

	// Create the user in the database
	user := models.User{Username: requestBody.Username, Email: requestBody.Email, Balance: requestBody.Balance, Password: requestBody.Password}

	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(201, gin.H{"message": "User created successfully"})
}

func Login(c *gin.Context) {
	var user struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var foundUser models.User
	initializers.DB.Where("username = ?", user.Username).First(&foundUser)

	claims := models.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
		UserID: foundUser.ID,
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func UserData(c *gin.Context) {
	username := c.Param("username")
	// Check if the user data is cached in Redis
	cachedUserData, err := initializers.RedisClient.Get(context.Background(), username).Result()
	if err == nil {
		// Data found in cache, return it

		var cachedUser models.User
		// Unmarshal the cached user data from JSON
		if err := json.Unmarshal([]byte(cachedUserData), &cachedUser); err != nil {
			// Handle unmarshaling error
			c.JSON(500, gin.H{
				"error": "Failed to unmarshal cached user data",
			})
			return
		}

		c.JSON(200, gin.H{
			"Data": cachedUserData,
		})
		return
	}
	// Data not found in cache, fetch from the database
	var user models.User
	result := initializers.DB.First(&user, "username = ?", username)
	if result.Error != nil {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	}
	// Cache the user data in Redis for future requests
	err = initializers.RedisClient.Set(context.Background(), username, user, 5*time.Minute).Err()
	if err != nil {
		// Handle Redis cache error
		c.JSON(500, gin.H{
			"error": "Failed to cache user data",
		})
		return
	}

	c.JSON(200, gin.H{
		"user": user,
	})
}

func IngestStockData(c *gin.Context) {
	var body struct {
		Ticker     string
		OpenPrice  float64
		ClosePrice float64
		High       float64
		Low        float64
		Volume     int
	}

	if err := c.Bind(&body); err != nil {
		// Handle the error (e.g., return an error response)
		return
	}
	data := models.StockData{Ticker: body.Ticker, OpenPrice: body.OpenPrice, ClosePrice: body.ClosePrice, High: body.High, Low: body.Low, Volume: body.Volume}
	result := initializers.DB.Create(&data)

	if result.Error != nil {
		c.Status(400)
		return
	}
	c.JSON(200, gin.H{
		"mesage": "data inserted successfully",
		"data":   data,
	})
}
func RetrieveAllStockData(c *gin.Context) {
	cachedStockData, err := initializers.RedisClient.Get(context.Background(), "all_stock_data").Result()
	if err == nil {
		// Data found in cache, return it as JSON

		var cachedData []models.StockData
		// Unmarshal the cached stock data from JSON
		if err := json.Unmarshal([]byte(cachedStockData), &cachedData); err != nil {
			// Handle unmarshaling error
			c.JSON(500, gin.H{
				"error": "Failed to unmarshal cached stock data",
			})
			return
		}

		c.JSON(200, gin.H{
			"data": cachedData,
		})
		return
	}

	var data []models.StockData
	result := initializers.DB.Find(&data)

	if result.Error != nil {
		c.JSON(404, gin.H{
			"error": "Stock Data not found",
		})
		return
	}

	// Marshal the stock data to JSON before caching
	serializedStockData, err := json.Marshal(data)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to marshal stock data",
		})
		return
	}

	// Cache the stock data in Redis for future requests
	err = initializers.RedisClient.Set(context.Background(), "all_stock_data", string(serializedStockData), 5*time.Minute).Err()
	if err != nil {
		// Handle Redis cache error
		c.JSON(500, gin.H{
			"error": "Failed to cache stock data",
		})
		return
	}

	c.JSON(200, gin.H{
		"data": data,
	})
}

func SpecificStockData(c *gin.Context) {
	ticker := c.Param("ticker")
	var stock []models.StockData
	result := initializers.DB.Find(&stock, "ticker = ?", ticker)

	if result.Error != nil {
		c.JSON(404, gin.H{
			"error": "Stock not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"ticker": stock,
	})
}

func RetrieveTransactionsOfSpecificUser(c *gin.Context) {
	user_id := c.Param("user_id")
	var transaction []models.Transaction
	result := initializers.DB.Find(&transaction, "user_id = ?", user_id)

	if result.Error != nil {
		c.JSON(404, gin.H{
			"error": "Transactions data not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"transaction": transaction,
	})
}

func Transaction(c *gin.Context) {
	var body struct {
		UserID            uint
		Ticker            string
		TransactionType   string
		TransactionVolume int
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	var stock models.StockData
	result := initializers.DB.First(&stock, "ticker = ?", body.Ticker)

	if result.Error != nil {
		c.JSON(404, gin.H{"error": "No record found for the provided ticker"})
		return
	}
	wg.Add(1)
	go func() {
		time.Sleep(10 * time.Second)
		transcation_type := body.TransactionType
		transaction_price := 0.0 // Initialize transaction_price

		if transcation_type == "sell" {
			high_price := stock.High
			transaction_price = high_price * float64(body.TransactionVolume)
		} else {
			low_price := stock.Low
			transaction_price = low_price * float64(body.TransactionVolume)
		}

		id := body.UserID
		var user models.User
		initializers.DB.First(&user, id)

		if user.Balance < transaction_price {
			c.JSON(400, gin.H{"error": "Your balance is less than your transaction"})
			return
		}

		balance := user.Balance - transaction_price
		initializers.DB.Model(&user).Updates(models.User{Balance: balance})

		transaction := models.Transaction{
			UserID:            body.UserID,
			Ticker:            body.Ticker,
			TransactionType:   body.TransactionType,
			TransactionVolume: body.TransactionVolume,
			TransactionPrice:  transaction_price,
		}

		if err := initializers.DB.Create(&transaction).Error; err != nil {
			c.JSON(400, gin.H{"error": "Transaction error"})
			return
		}
		wg.Done()
		// Log transaction success or failure
	}()

	c.JSON(200, gin.H{
		"message": "Transaction processing started in the background",
	})
	wg.Wait()
}

func TransactionsTimestemps(c *gin.Context) {
	userID := c.Param("user_id")
	startTimestamp := c.Param("start_timestamp")
	endTimestamp := c.Param("end_timestamp")

	// Convert the timestamps to time.Time objects
	input_layout := "2006-01-02"
	startTime, err := time.Parse(input_layout, startTimestamp)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid start timestamp"})
		return
	}

	endTime, err := time.Parse(input_layout, endTimestamp)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid end timestamp"})
		return
	}

	var transactions []models.Transaction
	result := initializers.DB.Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startTime, endTime).Find(&transactions)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	c.JSON(200, gin.H{"transactions": transactions})
}
