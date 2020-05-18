package transport

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

const userAgent = `Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36`

type Transport struct {
	upstream http.RoundTripper
	Cookies  http.CookieJar
}

func NewClient() (c *http.Client, err error) {

	scraperTransport, err := NewTransport(http.DefaultTransport)
	if err != nil {
		return
	}

	c = &http.Client{
		Transport: scraperTransport,
		Jar:       scraperTransport.Cookies,
	}

	return
}

func NewTransport(upstream http.RoundTripper) (*Transport, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Transport{upstream, jar}, nil
}

func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", userAgent)
	}

	if r.Header.Get("Referer") == "" {
		r.Header.Set("Referer", r.URL.String())
	}

	resp, err := t.upstream.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	// Check if Cloudflare anti-bot is on
	serverHeader := resp.Header.Get("Server")
	if resp.StatusCode == 503 && (serverHeader == "cloudflare-nginx" || serverHeader == "cloudflare") {
		log.Printf("Solving challenge for %s by calling remote selenium", resp.Request.URL.Hostname())
		resp, err := t.retrieveResponse(resp)

		return resp, err
	}

	return resp, err
}
func (t *Transport) retrieveResponse(resp *http.Response) (*http.Response, error) {
	const (
		port = 4444
	)

	req := resp.Request

	req.Header.Set("User-Agent", resp.Request.Header.Get("User-Agent"))

	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{Args: []string{"-headless"}}

	caps.AddFirefox(firefoxCaps)
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("https://13.59.61.191:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()
	if err := wd.Get(req.URL.String()); err != nil {
		panic(err)
	}
	time.Sleep(7000 * time.Millisecond)
	body, _ := wd.PageSource()
	response := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        make(http.Header, 0),
	}
	if cookies, err := wd.GetCookies(); err != nil {
		t.Cookies.SetCookies(req.URL, cookies)
	}

	return response, nil

}
