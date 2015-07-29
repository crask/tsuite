package main

import (
	"flag"
	"fmt"
	"html"
	//	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

var (
	port int
	path string
	help bool
)

func init() {
	flag.IntVar(&port, "p", 8080, "set http port, default 8080")
	flag.StringVar(&path, "s", "/foo", "set handler url, default \"/foo\"")
	flag.BoolVar(&help, "h", false, "show help")
}

func main() {

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
		var b []byte
		b, _ = httputil.DumpRequest(r, true)
		fmt.Println(string(b))

		//	b, _ = ioutil.ReadAll(r.Body)
		//		fmt.Println(string(b))
	})

	if err := http.ListenAndServe(":"+fmt.Sprintf("%d", port), nil); err != nil {
		fmt.Println(err.Error())
	}
}
