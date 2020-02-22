package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Handler))
	defer ts.Close()

	res, _ := http.Get(ts.URL + "/?search=good+boys")
	if res.StatusCode != 200 {
		t.Errorf("Server failing")
	}
}
