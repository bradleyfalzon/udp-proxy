package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
)

var (
	backends []*net.UDPAddr
)

func main() {
	flagListen := flag.String("listen", "", "host:port of UDP address to listen on, use :port for all interfaces")
	flagBackends := flag.String("backends", "", "CSV of host:port backends")
	flag.Parse()

	if *flagBackends == "" || *flagListen == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Backends:", *flagBackends)

	// Resolve backends
	for _, be := range strings.Split(*flagBackends, ",") {
		beAddr, err := net.ResolveUDPAddr("udp", be)
		if err != nil {
			log.Fatalf("Cannot resolve backend addr %v: %v", be, err)
		}
		backends = append(backends, beAddr)
	}

	log.Println("Listening to UDP:", *flagListen)

	// Listen to UDP
	addr, err := net.ResolveUDPAddr("udp", *flagListen)
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
		go handlePacket(buf[:n], addr, conn)
	}
}

func handlePacket(msg []byte, raddr *net.UDPAddr, conn *net.UDPConn) {
	// Listen on port for respones
	laddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatalf("Could not resolve listen addr: %v", err)
	}
	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal("Could not listen for backend repsonse:", err)
	}

	go func(raddr *net.UDPAddr, conn, lconn *net.UDPConn) {
		rbuf := make([]byte, 1500)

		// Receive the response
		n, _, err := lconn.ReadFromUDP(rbuf)
		if err != nil {
			log.Fatalln("Could not read from UDP in be listen loop:", err)
		}

		// Send response back to client
		conn.WriteToUDP(rbuf[:n], raddr)

		// Stop listening for other responses
		lconn.Close()
	}(raddr, conn, lconn)

	// Send the request to all backends
	for _, be := range backends {
		lconn.WriteToUDP(msg, be)
	}
}
