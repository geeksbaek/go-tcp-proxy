package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"regexp"
)

var (
	listenAddr = flag.String("l", "localhost:8080", "local address to listen on")
	remoteAddr = flag.String("r", "test.gilgil.net:80", "remote address to dial")
)

type Replacer func([]byte) []byte

var HostRe = regexp.MustCompile(`Host: (.*)\b`)

func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("listening: %v", err)
	}
	proxy(ln, *remoteAddr, func(r []byte) (w []byte) {
		return bytes.Replace(r, []byte("hello"), []byte("?????"), -1)
	})
}

func proxy(ln net.Listener, remoteAddr string, replacer Replacer) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		log.Printf("connected: %v", conn.RemoteAddr())
		go handle(conn, remoteAddr, replacer)
	}
}

func handle(conn net.Conn, remoteAddr string, replacer Replacer) {
	defer conn.Close()

	// TODO.
	// Accept된 Connection에서 Host를 추출한 뒤
	// Connection을 다시 Writable한 상태로 변경하기.
	rconn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("dialing remote: %v", err)
		return
	}
	defer rconn.Close()

	copy(conn, rconn, replacer)
}

func copy(sender, receiver io.ReadWriter, replacer Replacer) {
	go io.Copy(receiver, sender)

	// TODO.
	// 지나친 응답 딜레이 원인 찾기
	received, _ := ioutil.ReadAll(receiver)
	bytes.NewReader(replacer(received)).WriteTo(sender)
}
