package web

import (
    "fmt"
	// "encoding/json"
	"net/http"
	"net/http/httptest"
    "testing"	
	// "github.com/stretchr/testify/assert"
	"github.com/gin-gonic/gin"

	// tools "github.com/ericojonathan/k.digital/src/be/web/tools"
)

func TestGetEmployeeEncr(t *testing.T) {

	fmt.Println("[test_get_employee_encr]")

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/employee_encr", GetEmployeeEncr)

	//Initializes Viper for Config access
	Init()

	//To get a passing test, we need to pass the encrypted key
	// api_key, _ := tools.Encrypt("33a09a853f9b33da731f4a3e839d0c55")

	req, err := http.NewRequest(http.MethodGet, "/employee_encr?job_title=CNC%20Operator&api_key=718d618066ba7b849afbca1eabddc0398294d4f7d982c9062364eb8bed312a9fea01f1148b450600f9feaff2d6a1adb7a750efcb0f9befd8ba02cfd7145d904b", nil)
    if err != nil {
        t.Fatalf("Couldn't create request: %v\n", err)		
    }

	 // Create a response recorder so you can inspect the response
	 w := httptest.NewRecorder()

	 // Perform the request
	 r.ServeHTTP(w, req)
 
	 // Check to see if the response was what you expected
	 if w.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, w.Code)		
	 }	 
}