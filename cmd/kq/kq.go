package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	pages := flag.Int("p", 0, "search result pages. 0 or 1 return the first page, >0 creates a PIT")
	query := flag.String("q", "", "Elasticsearch query")
	trackHits := flag.Bool("track-hits", false, "Elasticsearch Track Total Hits")
	filterPath := flag.String("f", "", "set filter_path.  be sure to include pit_id if paginating")
	flag.Parse()

	if *pages < 0 {
		fmt.Fprintf(os.Stderr, "page count must be >= 0")
		os.Exit(1)
	}

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

	var paginationQuery map[string]interface{}
	if *pages > 0 {
		if body == nil {
			log.Fatal("no query to paginate over")
		}
		if *index == "" {
			*index = "*"
		}
		rsp, err := es.OpenPointInTime(
			strings.Split(*index, ","),
			es.OpenPointInTime.WithContext(ctx),
			es.OpenPointInTime.WithKeepAlive("1m"))
		if err != nil {
			log.Fatal("while creating PIT: ", err)
		}
		if rsp.IsError() {
			log.Fatal("while creating PIT: ", rsp.String())
		}
		var pit struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(rsp.Body).Decode(&pit); err != nil {
			log.Fatal("while parsing PIT response: ", err)
		}
		defer func() {
			if _, err := es.ClosePointInTime(
				es.ClosePointInTime.WithContext(ctx),
				es.ClosePointInTime.WithBody(strings.NewReader(`{"id":"`+pit.ID+`"}`)),
			); err != nil {
				log.Println("failed to close PIT: ", err)
			}
		}()

		var q map[string]interface{}
		if err := json.NewDecoder(body).Decode(&q); err != nil {
			log.Fatal("while parsing query: ", err)
		}
		if _, ok := q["sort"]; !ok {
			log.Fatal("missing sort in query")
		}
		q["pit"] = map[string]interface{}{"id": pit.ID, "keep_alive": "1m"}
		paginationQuery = q
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(&q); err != nil {
			log.Fatal("while generating query: ", err)
		}
		body = &buf
	}

	searchOptions := []func(*esapi.SearchRequest){
		es.Search.WithContext(ctx),
		es.Search.WithTrackTotalHits(*trackHits),
	}
	if body != nil {
		searchOptions = append(searchOptions, es.Search.WithBody(body))
	}
	if *index != "" && *pages == 0 {
		searchOptions = append(searchOptions, es.Search.WithIndex(*index))
	}
	if *filterPath != "" {
		searchOptions = append(searchOptions, es.Search.WithFilterPath(*filterPath))
	}

	result := search(es, searchOptions)
	for p := 1; p < *pages; p++ {
		var r struct {
			PITID string `json:"pit_id"`
			Hits  struct {
				Hits []struct {
					Sort []interface{} `json:"sort"`
				} `json:"hits"`
			} `json:"hits"`
		}
		if err := json.NewDecoder(result).Decode(&r); err != nil {
			log.Fatal("while decoding search response: ", err)
		}
		hitCount := len(r.Hits.Hits)
		if hitCount == 0 {
			break
		}
		paginationQuery["search_after"] = r.Hits.Hits[hitCount-1].Sort
		if oldPIT := paginationQuery["pit"].(map[string]interface{})["id"]; oldPIT != r.PITID {
			log.Printf("PIT changed, old: %s, new: %s", oldPIT, r.PITID)
		}
		paginationQuery["pit"] = map[string]interface{}{"id": r.PITID, "keep_alive": "1m"}
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(&paginationQuery); err != nil {
			log.Fatal("while generating query: ", err)
		}
		searchOptions = []func(*esapi.SearchRequest){
			es.Search.WithContext(ctx),
			es.Search.WithTrackTotalHits(*trackHits),
			es.Search.WithBody(&buf),
		}
		if *filterPath != "" {
			searchOptions = append(searchOptions, es.Search.WithFilterPath(*filterPath))
		}
		result = search(es, searchOptions)
	}
}

// search is for convenience
func search(es *esv7.Client, opts []func(*esapi.SearchRequest)) *bytes.Buffer {
	res, err := es.Search(opts...)
	if err != nil {
		log.Fatal(err)
	}
	if res.IsError() {
		log.Fatal(res.String())
	}
	var result bytes.Buffer
	if _, err := io.Copy(os.Stdout, io.TeeReader(res.Body, &result)); err != nil {
		log.Fatal(err)
	}
	return &result
}
