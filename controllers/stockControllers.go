package controllers

import (
	"Go_Assignment/m/constants"
	"Go_Assignment/m/initializers"
	"Go_Assignment/m/models"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var wg = sync.WaitGroup{}

// Signup godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags Users
// @Accept json
// @Produce json
// @Param username body string true "Username of the user"
// @Param email body string true "Email of the user"
// @Param balance body float64 true "Initial balance of the user"
// @Param password body string true "Password for the user"
// @Success 201 {string} string "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "Username already exists"
// @Failure 500 {object} map[string]string "Failed to create user"
// @Router /signup [post]
type data struct {
	Username string
	Email    string
	Balance  float64
	Password string
}

func Signup(c *gin.Context) {

	var requestBody data
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": constants.InvalidRequest})
		return
	}

	var existingUser models.User
	if err := initializers.DB.Where("username = ?", requestBody.Username).First(&existingUser).Error; err == nil {
		c.JSON(409, gin.H{"error": constants.UsernameExists})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": constants.Failed})
		return
	}

	user := models.User{
		Username: requestBody.Username,
		Email:    requestBody.Email,
		Balance:  requestBody.Balance,
		Password: string(hashedPassword),
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": constants.FailedCreateUser})
		return
	}

	c.JSON(201, gin.H{"message": constants.UserCreatedSuccessfully})
}

// Login godoc
// @Summary Authenticate a user
// @Description Authenticate a user by checking their username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param username formData string true "Username of the user"
// @Param password formData string true "Password for the user"
// @Success 200 {object} map[string]string "Authentication successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /login [post]

var user struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func Login(c *gin.Context) {

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var foundUser models.User
	if err := initializers.DB.Where("username = ?", user.Username).First(&foundUser).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": constants.UserFound})
			return
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": constants.InvalidCredentials})
		return
	}

	claims := models.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
		UserID: foundUser.ID,
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.CouldToken})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// UserData godoc
// @Summary Get user data
// @Description Retrieve user data by username
// @Tags Users
// @Param username path string true "Username of the user"
// @Produce json
// @Success 200 {object} map[string]interface{} "User data"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{username} [get]

func UserData(c *gin.Context) {
	username := c.Param("username")
	cachedUserData, err := initializers.RedisClient.Get(context.Background(), username).Result()
	if err == nil {

		var cachedUser models.User
		if err := json.Unmarshal([]byte(cachedUserData), &cachedUser); err != nil {
			c.JSON(500, gin.H{"error": constants.FailedUnmarshalData})
			return
		}

		c.JSON(200, gin.H{
			"Data": cachedUserData,
		})
		return
	}
	var user models.User
	result := initializers.DB.First(&user, "username = ?", username)
	if result.Error != nil {
		c.JSON(404, gin.H{"error": constants.UserFound})
		return
	}
	err = initializers.RedisClient.Set(context.Background(), username, user, 5*time.Minute).Err()
	if err != nil {
		c.JSON(500, gin.H{"error": constants.FailedCacheUserData})
		return
	}

	c.JSON(200, gin.H{
		"user": user,
	})
}

// IngestStockData godoc
// @Summary Ingest stock data
// @Description Ingest stock data for a specific ticker symbol
// @Tags Stocks
// @Accept json
// @Produce json
// @Param body body struct {
//   Ticker     string  `json:"ticker" binding:"required" example:"AAPL"`
//   OpenPrice  float64 `json:"openPrice" binding:"required" example:"150.0"`
//   ClosePrice float64 `json:"closePrice" binding:"required" example:"152.0"`
//   High       float64 `json:"high" binding:"required" example:"155.0"`
//   Low        float64 `json:"low" binding:"required" example:"148.0"`
//   Volume     int     `json:"volume" binding:"required" example:"10000"`
// }
// @Success 200 {object} map[string]interface{} "Data saved successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Router /stocks [post]

var payload struct {
	Ticker     string
	OpenPrice  float64
	ClosePrice float64
	High       float64
	Low        float64
	Volume     int
}

func IngestStockData(c *gin.Context) {

	if err := c.Bind(&payload); err != nil {
		return
	}
	data := models.StockData{Ticker: payload.Ticker, OpenPrice: payload.OpenPrice, ClosePrice: payload.ClosePrice, High: payload.High, Low: payload.Low, Volume: payload.Volume}
	result := initializers.DB.Create(&data)

	if result.Error != nil {
		c.Status(400)
		return
	}
	c.JSON(200, gin.H{
		"mesage": constants.DataSavedSuccessfully,
		"data":   data,
	})
}

func RetrieveAllStockData(c *gin.Context) {
	cachedStockData, err := initializers.RedisClient.Get(context.Background(), "all_stock_data").Result()
	if err == nil {

		var cachedData []models.StockData
		if err := json.Unmarshal([]byte(cachedStockData), &cachedData); err != nil {
			c.JSON(500, gin.H{"error": constants.FailedUnmarshal})
			return
		}

		c.JSON(200, gin.H{"data": cachedData})
		return
	}

	var data []models.StockData
	result := initializers.DB.Find(&data)

	if result.Error != nil {
		c.JSON(404, gin.H{"error": constants.StockDatanotFound})
		return
	}

	serializedStockData, err := json.Marshal(data)
	if err != nil {
		c.JSON(500, gin.H{"error": constants.FailedMarshalData})
		return
	}

	err = initializers.RedisClient.Set(context.Background(), "all_stock_data", string(serializedStockData), 5*time.Minute).Err()
	if err != nil {
		c.JSON(500, gin.H{"error": constants.FailedCacheData})
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
		c.JSON(404, gin.H{"error": constants.StocknotFound})
		return
	}

	c.JSON(200, gin.H{"ticker": stock})
}

func RetrieveTransactionsOfSpecificUser(c *gin.Context) {
	user_id := c.Param("user_id")
	var transaction []models.Transaction
	result := initializers.DB.Find(&transaction, "user_id = ?", user_id)

	if result.Error != nil {
		c.JSON(404, gin.H{"error": constants.TransactionsFound})
		return
	}

	c.JSON(200, gin.H{"transaction": transaction})
}

var transaction_data struct {
	UserID            uint
	Ticker            string
	TransactionType   string
	TransactionVolume int
}

func Transaction(c *gin.Context) {
	if err := c.ShouldBindJSON(&transaction_data); err != nil {
		c.JSON(400, gin.H{"error": constants.InvalidRequest})
		return
	}

	var stock models.StockData
	result := initializers.DB.First(&stock, "ticker = ?", transaction_data.Ticker)

	if result.Error != nil {
		c.JSON(404, gin.H{"error": constants.RecordfoundProvided})
		return
	}

	go func() {
		time.Sleep(10 * time.Second)
		transaction_price := 0.0
		transcation_t := transaction_data.TransactionType
		var transcation_type models.TransactionType
		if transcation_t == "sell" {
			transcation_type = models.Sell
			high_price := stock.High
			transaction_price = high_price * float64(transaction_data.TransactionVolume)
		} else {
			transcation_type = models.Buy
			low_price := stock.Low
			transaction_price = low_price * float64(transaction_data.TransactionVolume)
		}

		id := transaction_data.UserID
		var user models.User
		initializers.DB.First(&user, id)

		if user.Balance < transaction_price {
			c.JSON(400, gin.H{"error": constants.LessbalanceTransaction})
			return
		}

		balance := user.Balance - transaction_price
		initializers.DB.Model(&user).Updates(models.User{Balance: balance})

		transaction := models.Transaction{
			UserID:            transaction_data.UserID,
			Ticker:            transaction_data.Ticker,
			TransactionType:   transcation_type,
			TransactionVolume: transaction_data.TransactionVolume,
			TransactionPrice:  transaction_price,
		}

		if err := initializers.DB.Create(&transaction).Error; err != nil {
			c.JSON(400, gin.H{"error": constants.TransactionError})
			return
		}
	}()

	c.JSON(200, gin.H{"message": constants.TransactionProcessing})
}

func TransactionsTimestemps(c *gin.Context) {
	userID := c.Param("user_id")
	startTimestamp := c.Param("start_timestamp")
	endTimestamp := c.Param("end_timestamp")

	input_layout := "2006-01-02"
	startTime, err := time.Parse(input_layout, startTimestamp)
	if err != nil {
		c.JSON(400, gin.H{"error": constants.InvalidsTime})
		return
	}

	endTime, err := time.Parse(input_layout, endTimestamp)
	if err != nil {
		c.JSON(400, gin.H{"error": constants.InvalidTime})
		return
	}

	var transactions []models.Transaction
	result := initializers.DB.Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startTime, endTime).Find(&transactions)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": constants.Error})
		return
	}

	c.JSON(200, gin.H{"transactions": transactions})
}
