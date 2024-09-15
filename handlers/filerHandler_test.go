package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestGetFiles(t *testing.T) {
    router := gin.Default()
    router.GET("/files", GetFiles)

    req, _ := http.NewRequest("GET", "/files", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "files")
}

func TestGetFileURL(t *testing.T) {
    router := gin.Default()
    router.GET("/file/:file_id", GetFileURL) 

    req, _ := http.NewRequest("GET", "/file/123", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "file_url")
}
