package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchHandler))
	defer ts.Close()

	res, _ := http.Get(ts.URL + "?query=good+boys&engine=netnaija")
	if res.StatusCode != 200 {
		t.Errorf("Server failing")
	}
}

func TestListAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(ListHandler))
	defer ts.Close()

	res, _ := http.Get(ts.URL + "?page=1&engine=fzmovies")
	if res.StatusCode != 200 {
		t.Errorf("Server failing")
	}
}
