package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

// ChromeDpTransport : structure for chromedp instance
type ChromeDpTransport struct {
	upstream          http.RoundTripper
	Ctx               context.Context
	Cancel            context.CancelFunc
	RemoteAllocCancel context.CancelFunc
}

var (
	baseCtx context.Context
)

func getDebugURL() string {
	remoteUrl := fmt.Sprintf("%s/json/version", os.Getenv("GOPHIE_CHROMEDP_URL"))
	resp, err := http.Get(remoteUrl)
	if err != nil {
		log.Fatal(err)
	}

	var result map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	return result["webSocketDebuggerUrl"].(string)
}

// NewChromeDpTransport : initialize a new transport
func NewChromeDpTransport(upstream http.RoundTripper) (*ChromeDpTransport, error) {

	//  devToolWsUrl := getDebugURL()
	//  // create allocator context for use with creating a browser context later
	//  allocatorContext, cancel := chromedp.NewRemoteAllocator(context.Background(), devToolWsUrl)
	//  baseCtx, cancel = chromedp.NewContext(
	//    allocatorContext,
	//  )
	// for local exec
	baseCtx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Debugf),
	)

	// start the browser without a timeout
	if err := chromedp.Run(baseCtx); err != nil {
		panic(err)
	}
	return &ChromeDpTransport{
		upstream:          upstream,
		Ctx:               baseCtx,
		Cancel:            cancel,
		RemoteAllocCancel: cancel,
	}, nil
}

type LogAction struct {
	Message string
}

func (a LogAction) Do(ctx context.Context) error {
	log.Debug(a.Message)
	return nil
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

	// Check if CloudFlare blocker exists
	resp, err := http.Get(r.URL.String())
	if err != nil {
		return &http.Response{}, err
	}

	resp.Header.Set("Content-Type", "text/html")

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &http.Response{}, err
	}
	body = string(respBody)

	r.Header.Set("Content-Type", "text/html")

	log.Debug("Set Headers for page ", r.URL.String())

	if strings.Contains(strings.ToLower(string(respBody)), "wait a moment") {
		log.Debug("Using ChromeDP driver")
		// reuse baseCtx to create new context to reuse browser
		ctx, cancel := chromedp.NewContext(baseCtx)
		defer cancel()
		if err = chromedp.Run(ctx,
			chromedp.Navigate(r.URL.String()),
			LogAction{Message: "Finished navigation"},
			chromedp.WaitVisible(`#nnj-body, #top`),
			chromedp.OuterHTML("html", &body),
		); err != nil {
			log.Fatal(err)
			return &http.Response{}, err
		}
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
