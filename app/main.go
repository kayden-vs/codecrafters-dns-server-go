package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

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

		response := []byte{}

		// write header section
		requestID := binary.BigEndian.Uint16(buf[0:2])
		opCode := uint8((flags >> 11) & 0x0F)
		rd := uint8((flags >> 8) & 0x01)

		header := NewDNSHeader(requestID, opCode, rd)
		header.Qdcount = 1
		header.Ancount = 1
		response = header.Bytes()

		// write question section
		qn := NewDNSQn("codecrafters.io")
		response = append(response, qn.Bytes()...)

		// write answer section
		ans := NewDNSRecord("codecrafters.io", "8.8.8.8")
		response = append(response, ans.Bytes()...)

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}

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

func NewDNSHeader(id uint16, opcode uint8, rd uint8) DNSHeader {
	var rcode uint8
	if opcode != 0 {
		rcode = 4
	}

	return DNSHeader{
		ID:     id,
		QR:     1,
		Opcode: opcode,
		RD:     rd,
		Rcode:  rcode,
	}
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

type DNSQn struct {
	Name  string
	Type  uint16
	Class uint16
}

func NewDNSQn(name string) DNSQn {
	return DNSQn{
		Name:  name,
		Type:  1,
		Class: 1,
	}
}

func (q DNSQn) Bytes() []byte {
	var buf []byte

	// 1. Encode each label as [length][characters]
	// Trim trailing dot just in case (e.g., "google.com.")
	labels := strings.Split(strings.TrimSuffix(q.Name, "."), ".")
	for _, label := range labels {
		buf = append(buf, byte(len(label)))
		buf = append(buf, label...)
	}

	// 2. Terminate the domain name with a null byte (0x00)
	buf = append(buf, 0x00)

	// 3. Append Type (2 bytes) and Class (2 bytes) in Big Endian
	buf = binary.BigEndian.AppendUint16(buf, q.Type)
	buf = binary.BigEndian.AppendUint16(buf, q.Class)

	return buf
}

type DNSRecord struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

func NewDNSRecord(name string, ipAddr string) DNSRecord {
	ip := net.ParseIP(ipAddr).To4()
	return DNSRecord{
		Name:     name,
		Type:     1,
		Class:    1,
		TTL:      60,
		RDLength: uint16(len(ip)),
		RData:    ip,
	}
}

func (a DNSRecord) Bytes() []byte {
	var buf []byte

	// 1. Encode Name (same label encoding as DNSQn)
	labels := strings.Split(strings.TrimSuffix(a.Name, "."), ".")
	for _, label := range labels {
		buf = append(buf, byte(len(label)))
		buf = append(buf, label...)
	}
	buf = append(buf, 0x00)

	// 2. Append Type (2B), Class (2B), TTL (4B), and RDLength (2B)
	buf = binary.BigEndian.AppendUint16(buf, a.Type)
	buf = binary.BigEndian.AppendUint16(buf, a.Class)
	buf = binary.BigEndian.AppendUint32(buf, a.TTL)
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(a.RData)))

	// 3. Append raw RData bytes
	buf = append(buf, a.RData...)

	return buf
}
