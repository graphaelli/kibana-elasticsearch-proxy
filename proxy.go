package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"go.elastic.co/apm/module/apmhttp"

	"github.com/graphaelli/kibana-elasticsearch-proxy/transport"
)

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	cookies := flag.String("c", "", "cookies")
	debug := flag.Bool("D", false, "debug")
	kibana := flag.String("u", "", "Kibana URL")
	addr := flag.String("addr", "localhost:9222", "listen address")
	flag.Parse()

	var transportOptions []transport.Option
	if *kibana != "" {
		kibanaURL, err := url.Parse(*kibana)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse kibana URL %q: %s", *kibana, err)
			os.Exit(1)
		}
		transportOptions = append(transportOptions, transport.WithURL(kibanaURL))
	}
	if *debug {
		transportOptions = append(transportOptions, transport.WithDebug())
	}
	if *cookies != "" {
		h := http.Header{}
		h.Add("Cookie", *cookies)
		transportOptions = append(transportOptions, transport.WithHeaders(h))
	}

	// dummy URL - httputil.DumpRequestOut complains otherwise
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "proxy"})
	proxy.Transport = transport.New(transportOptions...)

	server := http.Server{
		Addr:    *addr,
		Handler: apmhttp.Wrap(logHandler(proxy)),
	}
	log.Printf("starting on http://%s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
