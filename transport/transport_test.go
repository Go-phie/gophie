package transport

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestTransport(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Get("https://thenetnaija.com/videos/movies")
	fmt.Println(c.Jar)

	if err != nil {
		t.Fatal(err)
	}

	_, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
}
