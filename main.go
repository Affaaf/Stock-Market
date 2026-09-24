package main

import (
	_ "Go_Assignment/m/docs"

	"Go_Assignment/m/controllers"
	"Go_Assignment/m/initializers"
	"Go_Assignment/m/models"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.RedisConfig()
}

// @title Tag Service API
// @version 2.0
// @description 	A Tag Service Api
// @host localhost:3030
// @BasePath /api

func main() {
	r := gin.Default()
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	// r.Use(AuthMiddleware())
	r.GET("/userdata/:username", controllers.UserData)
	r.POST("/ingeststockdata", controllers.IngestStockData)
	r.GET("/retrieve-stock-data", controllers.RetrieveAllStockData)
	r.GET("/specific-stock-data/:ticker", controllers.SpecificStockData)
	r.GET("/transactions-specific-user/:user_id", controllers.RetrieveTransactionsOfSpecificUser)
	r.POST("/transaction", controllers.Transaction)
	r.GET("/get-transactions-timestemps/:user_id/:start_timestamp/:end_timestamp", controllers.TransactionsTimestemps)

	r.Run()

}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Malformed token"})
			c.Abort()
			return
		}
		tokenString := bearerToken[1]
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("secret"), nil
		})

		if _, ok := token.Claims.(*models.JWTClaims); !ok || !token.Valid {
			var validationError *jwt.ValidationError
			if errors.As(err, &validationError) {
				if validationError.Errors&jwt.ValidationErrorMalformed != 0 {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Malformed token"})
				} else if validationError.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is either expired or not active yet"})
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
			}
			c.Abort()
			return
		}

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		claims := token.Claims.(*models.JWTClaims)
		c.Set("id", claims.UserID)

		c.Next()
	}
}
