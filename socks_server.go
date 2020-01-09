package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strconv"
)

// StartSocksServer Start a simple socks 5 server
func StartSocksServer() {
	log.SetFlags(log.Lshortfile)
	ln, err := net.Listen("tcp", ":20443")
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
		go handleSocksConnection(conn)
	}
}

func handleSocksConnection(conn net.Conn) {
	// https://tools.ietf.org/html/rfc1928
	defer conn.Close()
	r := bufio.NewReader(conn)
	/* The client connects to the server, and sends a version
	    identifier/method selection message:

	                   +----+----------+----------+
	                   |VER | NMETHODS | METHODS  |
	                   +----+----------+----------+
	                   | 1  |    1     | 1 to 255 |
					   +----+----------+----------+
	*/
	ver, err := r.ReadByte()
	if ver != 0x05 || err != nil {
		return
	}
	mcnt, err := r.ReadByte()
	if mcnt <= 0 || err != nil {
		return
	}
	var methods = make([]byte, mcnt)
	_, err = io.ReadFull(r, methods)
	if err != nil {
		return
	}
	/* The server selects from one of the methods given in METHODS, and
		sends a METHOD selection message:
		X'00' NO AUTHENTICATION REQUIRED
	                         +----+--------+
	                         |VER | METHOD |
	                         +----+--------+
	                         | 1  |   1    |
							 +----+--------+
	*/
	conn.Write([]byte{0x05, 0x00})

	/*The SOCKS request is formed as follows:
	     +----+-----+-------+------+----------+----------+
	     |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	     +----+-----+-------+------+----------+----------+
	     | 1  |  1  | X'00' |  1   | Variable |    2     |
	     +----+-----+-------+------+----------+----------+
	  Where:
	       o  VER    protocol version: X'05'
	       o  CMD
	          o  CONNECT X'01'
	          o  BIND X'02'
	          o  UDP ASSOCIATE X'03'
	       o  RSV    RESERVED
	       o  ATYP   address type of following address
	          o  IP V4 address: X'01'
	          o  DOMAINNAME: X'03'
	          o  IP V6 address: X'04'
	       o  DST.ADDR       desired destination address
	       o  DST.PORT desired destination port in network octet
	          order
	*/
	ver, err = r.ReadByte()
	if ver != 0x05 || err != nil {
		return
	}
	cmd, err := r.ReadByte()
	if err != nil {
		return
	}
	_, _ = r.ReadByte() // RSV
	atyp, err := r.ReadByte()
	if err != nil {
		return
	}
	if atyp != 0x01 && atyp != 0x03 {
		return
	}

	addrv4 := make([]byte, 4)
	domainName := ""
	if atyp == 0x01 {
		// the address is a version-4 IP address, with a length of 4 octets
		_, err = io.ReadFull(r, addrv4)
		if err != nil {
			return
		}
		// log.Println("IPv4:", addrv4)
	} else if atyp == 0x03 {
		// the address field contains a fully-qualified domain name.
		// The first octet of the address field contains the number of octets of name that follow, there is no terminating NUL octet.
		addrlen, err := r.ReadByte()
		if addrlen <= 0 || err != nil {
			return
		}
		dmName := make([]byte, addrlen)
		io.ReadFull(r, dmName)
		domainName = string(dmName)
		// log.Println("Domain Name:", domainName)
	}

	portArray := make([]byte, 2)
	_, err = io.ReadFull(r, portArray)
	if err != nil {
		return
	}
	port := binary.BigEndian.Uint16(portArray)
	// log.Println("Port:", port)

	if cmd == 0x01 {
		// CONNECT
		fulladdr := ""
		if atyp == 0x01 {
			fulladdr = strconv.Itoa(int(addrv4[0])) + "." + strconv.Itoa(int(addrv4[1])) + "." + strconv.Itoa(int(addrv4[2])) + "." + strconv.Itoa(int(addrv4[3]))
		} else if atyp == 0x03 {
			fulladdr = domainName
		}
		fulladdr += ":" + strconv.Itoa(int(port))
		target, err := net.Dial("tcp", fulladdr)
		if err != nil {
			return
		}
		defer target.Close()
		/* The SOCKS request information is sent by the client as soon as it has
		established a connection to the SOCKS server, and completed the
		authentication negotiations.  The server evaluates the request, and
		returns a reply formed as follows:
		     +----+-----+-------+------+----------+----------+
		     |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		     +----+-----+-------+------+----------+----------+
		     | 1  |  1  | X'00' |  1   | Variable |    2     |
		     +----+-----+-------+------+----------+----------+
		*/
		conn.Write([]byte{0x05})
		if err != nil {
			conn.Write([]byte{0x01}) // X'01' general SOCKS server failure
		} else {
			conn.Write([]byte{0x00})
		}
		conn.Write([]byte{0x00}) // rsv
		conn.Write([]byte{0x01})
		localaddr := target.LocalAddr().(*net.TCPAddr)
		conn.Write(localaddr.IP.To4())
		localport := make([]byte, 2)
		binary.BigEndian.PutUint16(localport, uint16(localaddr.Port))
		conn.Write(localport)

		// Start proxying
		errCh := make(chan error, 2)
		go proxy(target, conn, errCh)
		go proxy(conn, target, errCh)

		// Wait
		for i := 0; i < 2; i++ {
			e := <-errCh
			if e != nil {
				return
			}
		}
	} else if cmd == 0x02 {
		// BIND
	} else if cmd == 0x03 {
		// UDP
	}
}

func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err
}
