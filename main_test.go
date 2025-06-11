package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow" 

	requests := []struct {
		count int 
		expectedCount  int 
	}{
		{0, 0}, 
		{1, 1},  
		{2, 2}, 
		{100, len(cafeList[city])}, 
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?count=%d&city=%s", v.count, city)
		req := httptest.NewRequest("GET", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)
		responseBody := strings.TrimSpace(response.Body.String())

		if v.count == 0 {
            assert.Empty(t, responseBody, "if count=0 response should be empty")
            return
        }

		cafes := strings.Split(responseBody, ",")
		assert.Equal(t, v.expectedCount, len(cafes), 
			"for count=%d expected %d cafes, got %d", v.count, v.expectedCount, len(cafes))
		assert.Equal(t, v.expectedCount, strings.TrimSpace(response.Body.String()))

	}
}

func TestCafeSearch(t *testing.T) {
    handler := http.HandlerFunc(mainHandle)
    city := "moscow"

    testCases := []struct {
        searchQuery string
        minExpected int 
    }{
        {"фасоль", 0},    
        {"кофе", 2},    
        {"вилка", 1},      
    }

    for _, v := range testCases {
        t.Run(fmt.Sprintf("search=%s", v.searchQuery), func(t *testing.T) {
            url := fmt.Sprintf("/cafe?city=%s&search=%s", city, v.searchQuery)
            req := httptest.NewRequest("GET", url, nil)
            response := httptest.NewRecorder()

            handler.ServeHTTP(response, req)

            require.Equal(t, http.StatusOK, response.Code, "status code should be OK")
			responseBody := strings.TrimSpace(response.Body.String())

			if v.searchQuery == "фасоль" {
            assert.Empty(t, responseBody, "if search=фасоль response should be empty")
            return
        	}
			
			cafes := strings.Split(strings.TrimSpace(response.Body.String()), ",")

            assert.GreaterOrEqual(t, len(cafes), v.minExpected, 
                "should return at least %d cafes", v.minExpected)

            for _, cafe := range cafes {
                cafe = strings.TrimSpace(cafe)
                assert.True(t, strings.Contains(strings.ToLower(cafe), v.searchQuery),
                    "cafe '%s' should contain '%s'", cafe, v.searchQuery)
            }
        })
    }
}
