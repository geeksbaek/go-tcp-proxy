package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	listenPort = flag.String("p", ":8080", "local address to listen on")
)

func main() {
	flag.Parse()
	http.HandleFunc("/", proxy)
	http.ListenAndServe(*listenPort, nil)
}

func replacer(r io.Reader) io.Reader {
	d, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	dd := bytes.Replace(d, []byte("hello"), []byte("?????"), -1)
	return bytes.NewReader(dd)
}

func proxy(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	io.Copy(w, replacer(resp.Body))
}
