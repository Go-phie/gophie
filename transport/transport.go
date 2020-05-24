package transport

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

//const userAgent = `Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36`
const userAgent = `User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:75.0) Gecko/20100101 Firefox/75.0`

type Transport struct {
	upstream         http.RoundTripper
	Cookies          http.CookieJar
	webDriverCookies []selenium.Cookie
	Header           http.Header
	cookieURL        *url.URL // Too Access the Cookies you need the URL used to store them
}

var dummyURL = "https://netnaija.com"

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

	// Just a Dummy URL to use for cookieURL Validation
	blankURL, _ := url.Parse(dummyURL)
	return &Transport{
		upstream:         upstream,
		Cookies:          jar,
		webDriverCookies: []selenium.Cookie{},
		Header:           make(http.Header, 0),
		cookieURL:        blankURL}, nil
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if len(t.Header) != 0 {
		r.Header = t.Header
	} else {
		if r.Header.Get("User-Agent") == "" {
			r.Header.Set("User-Agent", userAgent)
		}

		if r.Header.Get("Referer") == "" {
			r.Header.Set("Referer", r.URL.String())
		}
		r.Header.Set("Upgrade-Insecure-Requests", "1")
		r.Header.Set("Connection", "keep-alive")
		r.Header.Set("Accept-Language", "en-US,en;q=0.5")
		r.Header.Set("TE", "Trailers")
		r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	}

	log.Info("\n\nCookie Key -----> ", r.URL, " Cookie key ----> ", t.cookieURL)
	// Cookies should be added only when they are available
	if t.cookieURL.String() != dummyURL {
		for _, cookie := range t.Cookies.Cookies(t.cookieURL) {
			r.AddCookie(cookie)
		}
	}

	resp, err := t.upstream.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	errBody, _ := ioutil.ReadAll(resp.Body)
	log.Info("Reponse --> ", string(errBody))

	log.Error("\n\n Request Headers ----> ", r.Header)

	// Check if Cloudflare anti-bot is on
	serverHeader := resp.Header.Get("Server")
	if resp.StatusCode == 503 && (serverHeader == "cloudflare-nginx" || serverHeader == "cloudflare") {
		resp.Body.Close()
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

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "text/html")

	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{Args: []string{"-headless"}}

	caps.AddFirefox(firefoxCaps)
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(t.webDriverCookies, []selenium.Cookie{}) {
		log.Debug("Found cookies, setting remote selenium cookies")
		for _, cookie := range t.webDriverCookies {
			wd.AddCookie(&cookie)
		}
	}
	defer wd.Quit()
	if err := wd.Get(req.URL.String()); err != nil {
		panic(err)
	}

	//  if reflect.DeepEqual(t.webDriverCookies, []selenium.Cookie{}) {
	time.Sleep(10000 * time.Millisecond)
	//  }
	body, _ := wd.PageSource()
	currentURL, _ := wd.CurrentURL()

	if !strings.HasPrefix(t.Header.Get("Referer"), "__cf") {
		// Referrer is usually the URL
		req.Header.Set("Referer", currentURL)
	}

	response := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        req.Header,
	}

	if cookies, err := wd.GetCookies(); err == nil {
		httpCookies := []*http.Cookie{}
		// Empty Existing Cookies before setting new ones
		jar, _ := cookiejar.New(nil)
		t.Cookies = jar
		for _, cookie := range cookies {
			httpCookies = append(httpCookies, &http.Cookie{
				Name:    cookie.Name,
				Value:   cookie.Value,
				Path:    cookie.Path,
				Domain:  cookie.Domain,
				Secure:  cookie.Secure,
				Expires: time.Unix(int64(cookie.Expiry), 0),
			})
		}
		t.Cookies.SetCookies(req.URL, httpCookies)
		log.Info("\n\nCookie Key -----> ", req.URL)
		t.webDriverCookies = cookies
		t.cookieURL = req.URL
	}
	t.Header = req.Header

	return response, nil

}
