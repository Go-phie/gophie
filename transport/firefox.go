package transport

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

type SeleniumTransport struct {
	upstream  http.RoundTripper
	WebDriver selenium.WebDriver
}

func NewClient() (c *http.Client, err error) {
	seleniumURL := fmt.Sprintf("%s/wd/hub", os.Getenv("GOPHIE_SELENIUM_URL"))
	fmt.Println("selenium url " + seleniumURL)
	scraperTransport, err := NewSeleniumTransport(http.DefaultTransport, seleniumURL)
	if err != nil {
		return
	}

	c = &http.Client{
		Transport: scraperTransport,
	}

	return
}

func NewSeleniumTransport(upstream http.RoundTripper, seleniumURL string) (*SeleniumTransport, error) {

	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{Args: []string{"-headless"}}
	caps.AddFirefox(firefoxCaps)
	wd, err := selenium.NewRemote(caps, seleniumURL)

	if err != nil {
		return &SeleniumTransport{}, err
	}
	return &SeleniumTransport{
		upstream:  upstream,
		WebDriver: wd,
	}, nil
}

func (t *SeleniumTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var (
		title string
		body  string
		err   error
	)

	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", userAgent)
	}

	if r.Header.Get("Referer") == "" {
		r.Header.Set("Referer", r.URL.String())
	}

	r.Header.Set("Content-Type", "text/html")

	if err = t.WebDriver.Get(r.URL.String()); err != nil {
		return &http.Response{}, err
	}

	// Check when it's blocked by cloudflare and Retry after 2seconds
	for {
		title, err = t.WebDriver.Title()
		if err != nil {
			return &http.Response{}, err
		}
		if !strings.HasPrefix(strings.ToLower(title), "just a moment") {
			break
		}

		time.Sleep(2000 * time.Millisecond)
	}

	body, err = t.WebDriver.PageSource()
	if err != nil {
		return &http.Response{}, err
	}

	response := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       r,
		Header:        r.Header,
	}

	return response, nil
}
