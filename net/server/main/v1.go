package main

import (
	"bufio"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

const (
	crlf      = "\r\n"
	separator = " "
)

func initVersion1() *Header {
	header := new(Header)
	header.Version = 1
	// Command doesn't exist in v1
	header.Command = PROXY
	return header
}

func parseVersion1(reader *bufio.Reader) (*Header, error) {
	buf := make([]byte, 0, 107)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf(ErrCantReadVersion1Header.Error()+": %v", err)
		}
		buf = append(buf, b)
		if b == '\n' {
			// End of header found
			break
		}
		if len(buf) == 107 {
			// No delimiter in first 107 bytes
			return nil, ErrVersion1HeaderTooLong
		}
		if reader.Buffered() == 0 {
			return nil, ErrCantReadVersion1Header
		}
	}

	// Check for CR before LF.
	if len(buf) < 2 || buf[len(buf)-2] != '\r' {
		return nil, ErrLineMustEndWithCrlf
	}

	tokens := strings.Split(string(buf[:len(buf)-2]), separator)

	if len(tokens) < 2 {
		return nil, ErrCantReadAddressFamilyAndProtocol
	}

	// Read address family and protocol
	var transportProtocol AddressFamilyAndProtocol
	switch tokens[1] {
	case "TCP4":
		transportProtocol = TCPv4
	case "TCP6":
		transportProtocol = TCPv6
	case "UNKNOWN":
		transportProtocol = UNSPEC // doesn't exist in v1 but fits UNKNOWN
	default:
		return nil, ErrCantReadAddressFamilyAndProtocol
	}

	// Expect 6 tokens only when UNKNOWN is not present.
	if transportProtocol != UNSPEC && len(tokens) < 6 {
		return nil, ErrCantReadAddressFamilyAndProtocol
	}

	header := initVersion1()

	// Transport protocol has been processed already.
	header.TransportProtocol = transportProtocol

	// When UNKNOWN, set the command to LOCAL and return early
	if header.TransportProtocol == UNSPEC {
		header.Command = LOCAL
		return header, nil
	}

	// Otherwise, continue to read addresses and ports
	sourceIP, err := parseV1IPAddress(header.TransportProtocol, tokens[2])
	if err != nil {
		return nil, err
	}
	destIP, err := parseV1IPAddress(header.TransportProtocol, tokens[3])
	if err != nil {
		return nil, err
	}
	sourcePort, err := parseV1PortNumber(tokens[4])
	if err != nil {
		return nil, err
	}
	destPort, err := parseV1PortNumber(tokens[5])
	if err != nil {
		return nil, err
	}
	header.SourceAddr = &net.TCPAddr{
		IP:   sourceIP,
		Port: sourcePort,
	}
	header.DestinationAddr = &net.TCPAddr{
		IP:   destIP,
		Port: destPort,
	}

	return header, nil
}

func parseV1PortNumber(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		return 0, ErrInvalidPortNumber
	}
	return port, nil
}

func parseV1IPAddress(protocol AddressFamilyAndProtocol, addrStr string) (net.IP, error) {
	addr, err := netip.ParseAddr(addrStr)
	if err != nil {
		return nil, ErrInvalidAddress
	}

	switch protocol {
	case TCPv4:
		if addr.Is4() {
			return addr.AsSlice(), nil
		}
	case TCPv6:
		if addr.Is6() || addr.Is4In6() {
			return addr.AsSlice(), nil
		}
	}

	return nil, ErrInvalidAddress
}
