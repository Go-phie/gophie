package transport

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

// ChromeDpTransport : structure for chromedp instance
type ChromeDpTransport struct {
	upstream http.RoundTripper
	Ctx      context.Context
	Cancel   context.CancelFunc
}

// NewChromeDpTransport : initialize a new transport
func NewChromeDpTransport(upstream http.RoundTripper) (*ChromeDpTransport, error) {

	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Debugf),
	)

	return &ChromeDpTransport{
		upstream: upstream,
		Ctx:      ctx,
		Cancel:   cancel,
	}, nil
}

// RoundTrip: extends the RoundTrip API for usage as a colly transport
func (t *ChromeDpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var (
		body string
		err  error
	)

	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", userAgent)
	}

	if r.Header.Get("Referer") == "" {
		r.Header.Set("Referer", r.URL.String())
	}

	r.Header.Set("Content-Type", "text/html")

	log.Debug("Set Headers for page ", r.URL.String())

	if err = chromedp.Run(t.Ctx,
		chromedp.Navigate(r.URL.String()),
		//    chromedp.WaitVisible(`#nnj-body`),
		chromedp.OuterHTML("html", &body),
	); err != nil {
		return &http.Response{}, err
	}
	log.Debug("Successfully retrieved body")

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
