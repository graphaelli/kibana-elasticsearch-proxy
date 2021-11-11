package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	esv7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/graphaelli/kibana-elasticsearch-proxy/transport"
)

func main() {
	cookies := flag.String("c", "", "cookies")
	debug := flag.Bool("D", false, "debug")
	kibana := flag.String("u", "", "Kibana URL")
	index := flag.String("index", "", "Elasticsearch Index")
	query := flag.String("q", "", "Elasticsearch query")
	trackHits := flag.Bool("track-hits", false, "Elasticsearch Track Total Hits")
	flag.Parse()

	var esConfig esv7.Config
	var transportOptions []transport.Option
	if *kibana != "" {
		kibanaURL, err := url.Parse(*kibana)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse kibana URL %q: %s", *kibana, err)
			os.Exit(1)
		}
		transportOptions = append(transportOptions, transport.WithURL(kibanaURL))

		if kibanaURL.User != nil {
			esConfig.Username = kibanaURL.User.Username()
			esConfig.Password, _ = kibanaURL.User.Password()
		}
	}
	if *cookies != "" {
		h := http.Header{}
		h.Add("Cookie", *cookies)
		transportOptions = append(transportOptions, transport.WithHeaders(h))
	}
	if *debug {
		transportOptions = append(transportOptions, transport.WithDebug())
	}
	esConfig.Transport = transport.New(transportOptions...)

	es, err := esv7.NewClient(esConfig)
	if err != nil {
		log.Fatal(err)
	}

	var body io.Reader
	if *query == "-" {
		body = os.Stdin
	} else if *query != "" {
		body = strings.NewReader(*query)
	}

	ctx := context.Background()
	searchOptions := []func(*esapi.SearchRequest){
		es.Search.WithContext(ctx),
	}
	if body != nil {
		searchOptions = append(searchOptions, es.Search.WithBody(body))
	}
	if *index != "" {
		searchOptions = append(searchOptions, es.Search.WithIndex(*index))
	}
	if *trackHits {
		searchOptions = append(searchOptions, es.Search.WithTrackTotalHits(true))
	}

	res, err := es.Search(searchOptions...)
	if err != nil {
		log.Fatal(err)
	}
	if res.IsError() {
		log.Fatal(res.String())
	}
	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		log.Fatal(err)
	}
}
