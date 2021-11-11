package transport

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type roundTripper struct {
	url    url.URL
	header http.Header
	rt     http.RoundTripper
	debug  bool
	logger *log.Logger
}

// RoundTrip makes this an http.RoundTripper
func (r *roundTripper) RoundTrip(reqOrig *http.Request) (*http.Response, error) {
	// RoundTrip is not supposed to mutate req, so copy req and set the custom headers only in the copy.
	reqCopy := *reqOrig
	reqCopy.Header = make(http.Header, len(reqOrig.Header))
	for k, v := range reqOrig.Header {
		reqCopy.Header[k] = v
	}
	req := &reqCopy
	req.Header.Add("kbn-xsrf", "kibana-elasticsearch-proxy")
	for k, vs := range r.header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	req.Method = http.MethodPost

	u := r.url
	req.URL = &u
	req.URL.RawQuery = url.Values{
		"method": []string{reqOrig.Method},
		"path":   []string{reqOrig.URL.Path},
	}.Encode()

	if r.debug {
		if out, err := httputil.DumpRequestOut(reqOrig, false); err == nil {
			r.logger.Println("original", string(out))
		}
		if out, err := httputil.DumpRequestOut(req, true); err == nil {
			r.logger.Println("outgoing", string(out))
		}
	}
	return r.rt.RoundTrip(req)
}

const proxyPath = "/api/console/proxy"

// Option sets options
type Option func(*roundTripper)

// New builds a new Kibana Proxy Transport
func New(o ...Option) http.RoundTripper {
	rt := &roundTripper{
		rt: http.DefaultTransport,
		url: url.URL{
			Scheme: "http",
			Host:   "localhost:5601",
			Path:   proxyPath,
		},
		debug:  false,
		logger: log.Default(),
	}

	for _, o := range o {
		o(rt)
	}
	return rt
}

// WithDebug enables transport debug logging
func WithDebug() Option {
	return func(rt *roundTripper) {
		rt.debug = true
	}
}

// WithHeaders enables request header customization
func WithHeaders(h http.Header) Option {
	return func(rt *roundTripper) {
		rt.header = h
	}
}

// WithLogger enables transport logging customization
func WithLogger(logger *log.Logger) Option {
	return func(rt *roundTripper) {
		rt.logger = logger
	}
}

// WithRoundTripper customizes the RoundTripper to wrap
func WithRoundTripper(wrt http.RoundTripper) Option {
	if wrt == nil {
		panic("wrapped RoundTripper is nil")
	}

	return func(rt *roundTripper) {
		rt.rt = wrt
	}
}

// WithURL sets the base URL for kibana communication
func WithURL(u *url.URL) Option {
	copyU := *u
	copyU.Path = strings.TrimSuffix(u.Path, "/") + "/api/console/proxy"
	return func(rt *roundTripper) {
		rt.url = copyU
	}
}
