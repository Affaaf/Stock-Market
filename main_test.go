package main

import (
	"Go_Assignment/m/controllers"
	"Go_Assignment/m/initializers"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToTestDb()
}

func TestSignup(t *testing.T) {
	router := gin.Default()
	router.POST("/users", controllers.Signup)
	payload := `{"username": "testatre", "Balance": 56767, "Email": "abc@gmail.com", "password": "pswd12"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	res := w.Body
	body, _ := ioutil.ReadAll(res)
	bodyString := string(body)
	var result struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(bodyString), &result); err != nil {
		return
	} else {
		assert.Equal(t, http.StatusCreated, w.Code)
	}
}

func TestInvalidSignup(t *testing.T) {
	router := gin.Default()
	router.POST("/users", controllers.Signup)
	payload := `{"username": "test", "Balance": 56767, "Email": "abc@gmail.com", "password": "pswd12"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	res := w.Body
	body, _ := ioutil.ReadAll(res)
	bodyString := string(body)
	var result struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(bodyString), &result); err != nil {
		return
	}
	if result.Error != "" {
		assert.Equal(t, http.StatusConflict, w.Code)
	}

}

func TestLogin(t *testing.T) {
	router := gin.Default()
	router.POST("/login", controllers.Login)

	payload := `{"username": "testa", "password": "pswd12"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	res := w.Body
	body, _ := ioutil.ReadAll(res)
	bodyString := string(body)

	var result struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal([]byte(bodyString), &result); err != nil {
		return
	} else {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, result.Token)
	}
}

func TestInvalidLogin(t *testing.T) {
	router := gin.Default()
	router.POST("/login", controllers.Login)

	payload := `{"username": "testa1332", "password": "pswd12"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	res := w.Body
	body, _ := ioutil.ReadAll(res)
	bodyString := string(body)

	var result struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal([]byte(bodyString), &result); err != nil {
		return
	} else {
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}
