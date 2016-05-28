package main

import (
	"log"
	"net"
	"strings"
)

var backends []*net.UDPAddr

func main() {

	log.Println("Resolving backends")

	// Resolve backends
	for _, be := range strings.Split("127.0.0.1:44120,127.0.0.1:44121", ",") {
		beAddr, err := net.ResolveUDPAddr("udp", be)
		if err != nil {
			log.Fatalf("Cannot resolve backend addr %v: %v", be, err)
		}
		backends = append(backends, beAddr)
	}

	log.Println("Listening to UDP")

	// Listen to UDP
	addr, err := net.ResolveUDPAddr("udp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("Listen loop...")

	// Receive packets
	buf := make([]byte, 1500)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error reading from udp on server listen loop: ", err)
			continue
		}

		//log.Printf("received %v from %v", string(buf[:n]), addr)

		go handlePacket(buf[:n], addr, conn)
	}
}

func handlePacket(msg []byte, raddr *net.UDPAddr, conn *net.UDPConn) {

	// Create the listen addr
	laddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatalf("Could not resolve listen addr: %v", err)
	}

	// Listen on port for respones
	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal("Could not listen for backend repsonse:", err)
	}

	go func(raddr *net.UDPAddr, conn, lconn *net.UDPConn) {
		//log.Println("Backend listen loop ready")

		rbuf := make([]byte, 1500)

		// Receive the response
		n, _, err := lconn.ReadFromUDP(rbuf)
		if err != nil {
			log.Fatalln("Could not read from UDP in be listen loop:", err)
		}

		//log.Printf("received %v from %v", string(rbuf[:n]), addr)

		// Send response back to client
		conn.WriteToUDP(rbuf[:n], raddr)

		// Stop listening for other response
		lconn.Close() // TODO pass reference
		return
	}(raddr, conn, lconn)

	// Send the request to all backends
	for _, be := range backends {
		//log.Println("Sending to backend:", be)
		lconn.WriteToUDP(msg, be)
	}
}
