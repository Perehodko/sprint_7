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
		want  int 
	}{
		{0, 0}, 
		{1, 1},  
		{2, 2}, 
		{100, len(cafeList["moscow"])}, 
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?count=%d&city=%s", v.count, city)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)
		
		cafes := strings.Split(response.Body.String(), ",")
		assert.Equal(t, v.want, len(cafes), 
			"for count=%d expected %d cafes, got %d", v.count, v.want, len(cafes))
		assert.Equal(t, v.want, strings.TrimSpace(response.Body.String()))

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

    for _, tc := range testCases {
        t.Run(fmt.Sprintf("search=%s", tc.searchQuery), func(t *testing.T) {
            url := fmt.Sprintf("/cafe?city=%s&search=%s", city, tc.searchQuery)
            req := httptest.NewRequest("GET", url, nil)
            w := httptest.NewRecorder()

            handler.ServeHTTP(w, req)

            require.Equal(t, http.StatusOK, w.Code, "status code should be OK")

            cafes := strings.Split(strings.TrimSpace(w.Body.String()), ",")
            assert.GreaterOrEqual(t, len(cafes), tc.minExpected, 
                "should return at least %d cafes", tc.minExpected)

            lowerSearch := strings.ToLower(tc.searchQuery)
            for _, cafe := range cafes {
                cafe = strings.TrimSpace(cafe)
                assert.True(t, strings.Contains(strings.ToLower(cafe), lowerSearch),
                    "cafe '%s' should contain '%s'", cafe, tc.searchQuery)
            }
        })
    }
}
