package main

import (
	"fmt"
	"golang.org/x/net/http2"
	"html"
	"log"
	"net/http"
)

func main() {
	var srv http.Server
	http2.VerboseLogs = true
	srv.Addr = ":8080"

	// This enables http2 support
	http2.ConfigureServer(&srv, nil)

	// Plain text test handler
	// Open https://localhost:8080/randomtest
	// in your Chrome Canary browser
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi tester %q\n", html.EscapeString(r.URL.Path))
		ShowRequestInfoHandler(w, r)
	})

	// Listen as https ssl server
	// NOTE: WITHOUT SSL IT WONT WORK!!
	// To self generate a test ssl cert/key you could go to
	// http://www.selfsignedcertificate.com/
	// or read the openssl manual
	log.Fatal(srv.ListenAndServeTLS("localhost.cert", "localhost.key"))
}

func ShowRequestInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "Protocol: %s\n", r.Proto)
	fmt.Fprintf(w, "Host: %s\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "RequestURI: %q\n", r.RequestURI)
	fmt.Fprintf(w, "URL: %#v\n", r.URL)
	fmt.Fprintf(w, "Body.ContentLength: %d (-1 means unknown)\n", r.ContentLength)
	fmt.Fprintf(w, "Close: %v (relevant for HTTP/1 only)\n", r.Close)
	fmt.Fprintf(w, "TLS: %#v\n", r.TLS)
	fmt.Fprintf(w, "\nHeaders:\n")
	r.Header.Write(w)
}
