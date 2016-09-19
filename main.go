package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strings"
)

var (
	listenAddr = flag.String("l", "localhost:8080", "local address to listen on")
)

type Replacer func([]byte) []byte

var HostRe = regexp.MustCompile(`Host: (.*)\b`)

func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("listening: %v", err)
	}
	proxy(ln, func(r []byte) (w []byte) {
		return bytes.Replace(r, []byte("hello"), []byte("?????"), -1)
	})
}

func proxy(ln net.Listener, replacer Replacer) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		log.Printf("connected: %v", conn.RemoteAddr())
		go handle(conn, replacer)
	}
}

func handle(conn net.Conn, replacer Replacer) {
	defer conn.Close()

	reqData, err := ioutil.ReadAll(conn)
	if err != nil {
		panic(err)
	}

	matched := HostRe.FindSubmatch(reqData)
	if len(matched) != 2 {
		return
	}

	remoteAddr := func() string {
		if strings.Contains(string(matched[1]), ":") {
			return string(matched[1])
		}
		return string(matched[1]) + ":80"
	}()

	// TODO.
	// Accept된 Connection에서 Host를 추출한 뒤
	// Connection을 다시 Writable한 상태로 변경하기.
	rconn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("dialing remote: %v", err)
		return
	}
	defer rconn.Close()

	go io.Copy(rconn, bytes.NewReader(reqData))

	// TODO.
	// 지나친 응답 딜레이 원인 찾기
	received, _ := ioutil.ReadAll(rconn)
	bytes.NewReader(replacer(received)).WriteTo(conn)
}
