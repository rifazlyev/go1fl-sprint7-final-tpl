package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		req := httptest.NewRequest(http.MethodGet, v.request, nil)
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
	requests := []struct {
		count     int
		wantCount int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(100, len(cafeList["moscow"]))},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		targetURL := fmt.Sprintf("/cafe?city=moscow&count=%d", v.count)
		req := httptest.NewRequest(http.MethodGet, targetURL, nil)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		var resultCafeList []string
		if responseBody := response.Body.String(); responseBody != "" {
			resultCafeList = strings.Split(responseBody, ",")
		}
		assert.Len(t, resultCafeList, v.wantCount)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
		{"", len(cafeList["moscow"])},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		targetURL := fmt.Sprintf("/cafe?city=moscow&search=%s", v.search)
		req := httptest.NewRequest(http.MethodGet, targetURL, nil)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		var resultCafeList []string
		if responseBody := response.Body.String(); responseBody != "" {
			resultCafeList = strings.Split(responseBody, ",")
		}
		if v.search != "" {
			searchLower := strings.ToLower(v.search)
			for _, cafe := range resultCafeList {
				cafeLower := strings.ToLower(cafe)
				assert.Contains(t, cafeLower, searchLower)
			}
		}
		assert.Len(t, resultCafeList, v.wantCount)
	}
}
