package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type DNSMsg struct {
	header   DNSHeader
	question DNSQn
	record   DNSRecord
}

type DNSHeader struct {
	ID      uint16
	QR      uint8
	Opcode  uint8
	AA      uint8
	TC      uint8
	RD      uint8
	RA      uint8
	Z       uint8
	Rcode   uint8
	Qdcount uint16
	Ancount uint16
	Nscount uint16
	Arcount uint16
}

func NewDNSHeader(id uint16) DNSHeader {
	return DNSHeader{
		ID: id,
		QR: 1,
	}
}

type DNSQn struct {
}

type DNSRecord struct {
}

func (h DNSHeader) Bytes() []byte {
	response := make([]byte, 12)

	binary.BigEndian.PutUint16(response[0:2], h.ID)
	binary.BigEndian.PutUint16(response[2:4], h.PackFlags())
	binary.BigEndian.PutUint16(response[4:6], h.Qdcount)
	binary.BigEndian.PutUint16(response[6:8], h.Ancount)
	binary.BigEndian.PutUint16(response[8:10], h.Nscount)
	binary.BigEndian.PutUint16(response[10:12], h.Arcount)

	return response
}

func (h DNSHeader) PackFlags() uint16 {
	var flags uint16
	flags |= uint16(h.QR&0x1) << 15
	flags |= uint16(h.Opcode&0xF) << 11
	flags |= uint16(h.AA&0x1) << 10
	flags |= uint16(h.TC&0x1) << 9
	flags |= uint16(h.RD&0x1) << 8
	flags |= uint16(h.RA&0x1) << 7
	flags |= uint16(h.Z&0x7) << 4
	flags |= uint16(h.Rcode & 0xF)
	return flags
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		// Create an empty response
		response := []byte{}

		// write header
		requestID := binary.BigEndian.Uint16(buf[0:2])
		header := NewDNSHeader(requestID)
		response = header.Bytes()

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
