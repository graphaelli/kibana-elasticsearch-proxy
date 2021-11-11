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

func main() {
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

	// dummy URL - httputil.DumpRequestOut compplains otherwise
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "proxy"})
	proxy.Transport = transport.New(transportOptions...)

	server := http.Server{
		Addr:    *addr,
		Handler: apmhttp.Wrap(proxy),
	}
	log.Printf("starting on http://%s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}