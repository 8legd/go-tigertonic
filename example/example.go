package main

import (
	"flag"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-tigertonic"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	cert   = flag.String("cert", "", "certificate pathname")
	key    = flag.String("key", "", "private key pathname")
	listen = flag.String("listen", "127.0.0.1:8000", "listen address")

	hMux tigertonic.HostServeMux
	mux, nsMux *tigertonic.TrieServeMux
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: example [-cert=<cert>] [-key=<key>] [-listen=<listen>]")
		flag.PrintDefaults()
	}
	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Lshortfile)

	mux = tigertonic.NewTrieServeMux()
	mux.Handle("POST", "/stuff", tigertonic.Timed(tigertonic.Marshaled(create), "POST-stuff", nil))
	mux.Handle("GET", "/stuff/{id}", tigertonic.Timed(tigertonic.Marshaled(get), "GET-stuff-id", nil))
	mux.Handle("POST", "/stuff/{id}", tigertonic.Timed(tigertonic.Marshaled(update), "POST-stuff-id", nil))
	nsMux = tigertonic.NewTrieServeMux()
	nsMux.HandleNamespace("", mux)
	nsMux.HandleNamespace("/1.0", mux)
	hMux = tigertonic.NewHostServeMux()
	hMux.Handle("example.com", nsMux)
}

func main() {
	flag.Parse()
	go metrics.Log(metrics.DefaultRegistry, 60e9, log.New(os.Stderr, "metrics ", log.Lmicroseconds))
	server := tigertonic.NewServer(*listen, tigertonic.Logged(hMux, func(s string) string {
		return strings.Replace(s, "SECRET", "REDACTED", -1)
	}))
	var err error
	if "" != *cert && "" != *key {
		err = server.ListenAndServeTLS(*cert, *key)
	} else {
		err = server.ListenAndServe()
	}
	if nil != err {
		log.Fatalln(err)
	}
}

// POST /stuff
func create(u *url.URL, h http.Header, rq *MyRequest) (int, http.Header, *MyResponse, error) {
	return http.StatusCreated, http.Header{
		"Content-Location": {fmt.Sprintf(
			"%s://%s/1.0/stuff/%s", // TODO Don't hard-code this.
			u.Scheme,
			u.Host,
			rq.ID,
		)},
	}, &MyResponse{rq.ID, rq.Stuff}, nil
}

// GET /stuff/{id}
func get(u *url.URL, h http.Header, _ interface{}) (int, http.Header, *MyResponse, error) {
	return http.StatusOK, nil, &MyResponse{u.Query().Get("id"), "STUFF"}, nil
}

// POST /stuff/{id}
func update(u *url.URL, h http.Header, rq *MyRequest) (int, http.Header, *MyResponse, error) {
	return http.StatusAccepted, nil, &MyResponse{u.Query().Get("id"), "STUFF"}, nil
}
