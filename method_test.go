package gin

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/illidaris/logger"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	router *gin.Engine
)

func init() {
	// init log core
	logger.OnlyConsole()
	// init gin
	router = gin.Default()
	router.Use(LoggerHandler())
	router.Use(RecoverHandler())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})
	router.GET("/error", func(c *gin.Context) {
		panic(errors.New("this is an error"))
	})
}

// Get get method to gin mock server
func Get(uri string, router *gin.Engine) (int, []byte) {
	req := httptest.NewRequest("GET", uri, nil)
	const rid = "X-Request-ID"
	req.Header.Set(rid, "123456")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	result := w.Result()
	defer result.Body.Close()
	body, _ := ioutil.ReadAll(result.Body)
	return result.StatusCode, body
}

// TestLoggerHandler test log
func TestLoggerHandler(t *testing.T) {
	uri := "/test"
	code, body := Get(uri, router)
	if code != 200 {
		t.Errorf("reponse code is not 200，code:%d\n", code)
	}
	fmt.Printf("response:%v\n", string(body))
	if string(body) != "success" {
		t.Errorf("reponse msg，body:%v\n", string(body))
	}
}

func ExampleLoggerHandler() {
	router.Use(LoggerHandler())
}

// TestRecoverHandler test error log
func TestRecoverHandler(t *testing.T) {
	uri := "/error"
	code, body := Get(uri, router)
	fmt.Printf("response:%v\n", string(body))
	if code != 500 {
		t.Errorf("reponse code is not 200，code:%d\n", code)
	}
}

func ExampleRecoverHandler() {
	router.Use(RecoverHandler())
}
