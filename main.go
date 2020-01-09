package main

import "flag"

import "log"

func main() {
	serverPtr := flag.Bool("s", false, "start as server")
	// urlPtr := flag.String("url", "http://www.baidu.com", "url for getting")
	// tlsServerPtr := flag.String("addr", "127.0.0.1:20443", "tls server address")
	flag.Parse()
	if *serverPtr {
		log.Println("Server Mode -->")
		// StartTLSServer()
		StartSocksServer()
	} else {
		log.Println("Cient Mode -->")
		// TLSClient(*tlsServerPtr, *urlPtr)
	}
}
