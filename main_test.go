package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
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

	var requests []struct {
		count int
		want  int
	}
	var actual int

	for city, cafes := range cafeList {
		requests = []struct {
			count int
			want  int
		}{
			{0, 0},
			{1, 1},
			{2, 2},
			{100, min(len(cafes), 100)},
		}
		for _, v := range requests {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city="+city+"&count="+strconv.Itoa(v.count), nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			if response.Body.String() != "" {
				reqCafes := strings.Split(response.Body.String(), ",")
				actual = len(reqCafes)
			} else {
				actual = 0
			}
			assert.Equal(t, v.want, actual, "for city=%q, count=%d", city, v.count)
		}
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	var requests = []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/cafe?city=moscow&search="+v.search, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		var got []string
		if response.Body.String() != "" {
			got = strings.Split(response.Body.String(), ",")
		}
		for _, cafe := range got {
			assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(v.search)), "for search=%q", v.search)
		}
		assert.Len(t, got, v.wantCount, "for search=%q", v.search)
	}
}
