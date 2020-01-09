package main

import (
	"crypto/tls"
	"log"
)

// TLSClient a simple tls client
func TLSClient(tlsServer string, url string) {
	log.SetFlags(log.Lshortfile)

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", tlsServer, conf)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	n, err := conn.Write([]byte(url + "\n"))
	if err != nil {
		log.Println(n, err)
		return
	}

	buf := make([]byte, 60000)
	for {
		n, err = conn.Read(buf)
		if n > 0 {
			println(string(buf[:n]))
		}
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}
