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

func Signup(c *gin.Context) {
	type data struct {
		Username string
		Email    string
		Balance  float64
		Password string
	}

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

func UserData(c *gin.Context) {
	username := c.Param("username")
	cachedUserData, err := initializers.RedisClient.Get(context.Background(), username).Result()
	if err == nil {

		var cachedUser models.User
		if err := json.Unmarshal([]byte(cachedUserData), &cachedUser); err != nil {
			c.JSON(500, gin.H{
				"error": constants.FailedUnmarshalData,
			})
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
		c.JSON(404, gin.H{
			"error": constants.UserFound,
		})
		return
	}
	err = initializers.RedisClient.Set(context.Background(), username, user, 5*time.Minute).Err()
	if err != nil {
		c.JSON(500, gin.H{
			"error": constants.FailedCacheUserData,
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
		return
	}
	data := models.StockData{Ticker: body.Ticker, OpenPrice: body.OpenPrice, ClosePrice: body.ClosePrice, High: body.High, Low: body.Low, Volume: body.Volume}
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
			c.JSON(500, gin.H{
				"error": constants.FailedUnmarshal,
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
			"error": constants.StockDatanotFound,
		})
		return
	}

	serializedStockData, err := json.Marshal(data)
	if err != nil {
		c.JSON(500, gin.H{
			"error": constants.FailedMarshalData,
		})
		return
	}

	err = initializers.RedisClient.Set(context.Background(), "all_stock_data", string(serializedStockData), 5*time.Minute).Err()
	if err != nil {
		c.JSON(500, gin.H{
			"error": constants.FailedCacheData,
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
			"error": constants.StocknotFound,
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
			"error": constants.TransactionsFound,
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
		c.JSON(400, gin.H{"error": constants.InvalidRequest})
		return
	}

	var stock models.StockData
	result := initializers.DB.First(&stock, "ticker = ?", body.Ticker)

	if result.Error != nil {
		c.JSON(404, gin.H{"error": constants.RecordfoundProvided})
		return
	}
	wg.Add(1)
	go func() {
		time.Sleep(10 * time.Second)
		transaction_price := 0.0
		transcation_t := body.TransactionType
		var transcation_type models.TransactionType
		if transcation_t == "sell" {
			transcation_type = models.Sell

			high_price := stock.High
			transaction_price = high_price * float64(body.TransactionVolume)
		} else {
			transcation_type = models.Buy
			low_price := stock.Low
			transaction_price = low_price * float64(body.TransactionVolume)
		}

		id := body.UserID
		var user models.User
		initializers.DB.First(&user, id)

		if user.Balance < transaction_price {
			c.JSON(400, gin.H{"error": constants.LessbalanceTransaction})
			return
		}

		balance := user.Balance - transaction_price
		initializers.DB.Model(&user).Updates(models.User{Balance: balance})

		transaction := models.Transaction{
			UserID:            body.UserID,
			Ticker:            body.Ticker,
			TransactionType:   transcation_type,
			TransactionVolume: body.TransactionVolume,
			TransactionPrice:  transaction_price,
		}

		if err := initializers.DB.Create(&transaction).Error; err != nil {
			c.JSON(400, gin.H{"error": constants.TransactionError})
			return
		}
		wg.Done()
	}()

	c.JSON(200, gin.H{
		"message": constants.TransactionProcessing,
	})
	wg.Wait()
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
