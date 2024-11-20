package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func GetTestGinContext(w *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
	}

	return ctx
}

func TestGetUrlFromBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := GetTestGinContext(w)

	body := make(map[string]string)
	body["bodyURL"] = "https://github.com/example"
	exampleJson, _ := json.Marshal(body)
	exampleBody := strings.NewReader(string(exampleJson))
	r.Request.Body = io.NopCloser(exampleBody)

	e, err := GetUrlFromBody(r)
	if err != nil {
		fmt.Println("Error reading body:", err)
	}

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "https://github.com/example", e)
}

func TestGetButtonIdFromBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := GetTestGinContext(w)

	body := make(map[string]string)
	body["buttonID"] = "increaseButton"
	exampleJson, _ := json.Marshal(body)
	exampleBody := strings.NewReader(string(exampleJson))
	r.Request.Body = io.NopCloser(exampleBody)

	e, err := GetButtonIdFromBody(r)
	if err != nil {
		fmt.Println("Error reading body:", err)
	}

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "increaseButton", e)
}

func TestCheckIfRepoAlreadyCloned(t *testing.T) {
	exampleRepoFolderName := "counterApp"
	isCloned := CheckIfRepoAlreadyCloned(exampleRepoFolderName)
	assert.Equal(t, false, isCloned)
}

func TestGetRepoFolderName(t *testing.T) {
	exampleRepoName := "https://github.com/exampleRepoFolder"
	exampleRepoFolderName := GetRepoFolderName(exampleRepoName)
	assert.Equal(t, "exampleRepoFolder", exampleRepoFolderName)
}
