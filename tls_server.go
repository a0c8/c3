package main

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"net/http"
)

// StartTLSServer Start a simple tls server
func StartTLSServer() {
	log.SetFlags(log.Lshortfile)

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":20443", config)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleTLSConnection(conn)
	}
}

func handleTLSConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		println(msg[:len(msg)-1])

		// get http response
		resp, err := http.Get(msg[:len(msg)-1])
		if err != nil {
			println(err.Error())
			return
		}
		defer resp.Body.Close()
		log.Println("Response status:", resp.Status)

		buf := make([]byte, 60000)
		for {
			rn, readerr := resp.Body.Read(buf)
			// println(string(buf[:rn]))
			if rn > 0 {
				// log.Println("Begin write: ", rn)
				_, writeerr := conn.Write(buf[:rn])
				if writeerr != nil {
					log.Println("Failed to write")
					return
				}
				// log.Println("End write")
			}

			if readerr != nil {
				log.Println(readerr.Error())
				return
			}

		}
	}
}
