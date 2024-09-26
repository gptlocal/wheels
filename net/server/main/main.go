package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

type Policy int

const (
	USE Policy = iota
	REJECT
	REQUIRE
	SKIP
)

type ProtocolVersionAndCommand byte

const (
	// LOCAL represents the LOCAL command in v2 or UNKNOWN transport in v1, in which case no address information is expected.
	LOCAL ProtocolVersionAndCommand = '\x20'
	// PROXY represents the PROXY command in v2 or transport is not UNKNOWN in v1, in which case valid local/remote address and port information is expected.
	PROXY ProtocolVersionAndCommand = '\x21'
)

var supportedCommand = map[ProtocolVersionAndCommand]bool{
	LOCAL: true,
	PROXY: true,
}

func (pvc ProtocolVersionAndCommand) IsLocal() bool {
	return LOCAL == pvc
}

type AddressFamilyAndProtocol byte

const (
	UNSPEC       AddressFamilyAndProtocol = '\x00'
	TCPv4        AddressFamilyAndProtocol = '\x11'
	UDPv4        AddressFamilyAndProtocol = '\x12'
	TCPv6        AddressFamilyAndProtocol = '\x21'
	UDPv6        AddressFamilyAndProtocol = '\x22'
	UnixStream   AddressFamilyAndProtocol = '\x31'
	UnixDatagram AddressFamilyAndProtocol = '\x32'
)

// IsIPv4 returns true if the address family is IPv4 (AF_INET4), false otherwise.
func (ap AddressFamilyAndProtocol) IsIPv4() bool {
	return ap&0xF0 == 0x10
}

// IsIPv6 returns true if the address family is IPv6 (AF_INET6), false otherwise.
func (ap AddressFamilyAndProtocol) IsIPv6() bool {
	return ap&0xF0 == 0x20
}

// IsUnix returns true if the address family is UNIX (AF_UNIX), false otherwise.
func (ap AddressFamilyAndProtocol) IsUnix() bool {
	return ap&0xF0 == 0x30
}

// IsStream returns true if the transport protocol is TCP or STREAM (SOCK_STREAM), false otherwise.
func (ap AddressFamilyAndProtocol) IsStream() bool {
	return ap&0x0F == 0x01
}

// IsDatagram returns true if the transport protocol is UDP or DGRAM (SOCK_DGRAM), false otherwise.
func (ap AddressFamilyAndProtocol) IsDatagram() bool {
	return ap&0x0F == 0x02
}

// IsUnspec returns true if the transport protocol or address family is unspecified, false otherwise.
func (ap AddressFamilyAndProtocol) IsUnspec() bool {
	return (ap&0xF0 == 0x00) || (ap&0x0F == 0x00)
}

type Conn struct {
	net.Conn
	readErr   error
	once      sync.Once
	header    *Header
	bufReader *bufio.Reader
}

type Header struct {
	Version           byte
	Command           ProtocolVersionAndCommand
	TransportProtocol AddressFamilyAndProtocol
	SourceAddr        net.Addr
	DestinationAddr   net.Addr
	rawTLVs           []byte
}

var (
	SIGV1 = []byte{'\x50', '\x52', '\x4F', '\x58', '\x59'}
	SIGV2 = []byte{'\x0D', '\x0A', '\x0D', '\x0A', '\x00', '\x0D', '\x0A', '\x51', '\x55', '\x49', '\x54', '\x0A'}

	ErrCantReadVersion1Header               = errors.New("proxyproto: can't read version 1 header")
	ErrVersion1HeaderTooLong                = errors.New("proxyproto: version 1 header must be 107 bytes or less")
	ErrLineMustEndWithCrlf                  = errors.New("proxyproto: version 1 header is invalid, must end with \\r\\n")
	ErrCantReadProtocolVersionAndCommand    = errors.New("proxyproto: can't read proxy protocol version and command")
	ErrCantReadAddressFamilyAndProtocol     = errors.New("proxyproto: can't read address family or protocol")
	ErrCantReadLength                       = errors.New("proxyproto: can't read length")
	ErrCantResolveSourceUnixAddress         = errors.New("proxyproto: can't resolve source Unix address")
	ErrCantResolveDestinationUnixAddress    = errors.New("proxyproto: can't resolve destination Unix address")
	ErrNoProxyProtocol                      = errors.New("proxyproto: proxy protocol signature not present")
	ErrUnknownProxyProtocolVersion          = errors.New("proxyproto: unknown proxy protocol version")
	ErrUnsupportedProtocolVersionAndCommand = errors.New("proxyproto: unsupported proxy protocol version and command")
	ErrUnsupportedAddressFamilyAndProtocol  = errors.New("proxyproto: unsupported address family and protocol")
	ErrInvalidLength                        = errors.New("proxyproto: invalid length")
	ErrInvalidAddress                       = errors.New("proxyproto: invalid address")
	ErrInvalidPortNumber                    = errors.New("proxyproto: invalid port number")
	ErrSuperfluousProxyHeader               = errors.New("proxyproto: upstream connection sent PROXY header but isn't allowed to send one")
)

func main() {
	addr := "localhost:9876"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("couldn't listen to %q: %q\n", addr, err.Error())
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("couldn't accept %q: %q\n", conn, err.Error())
	}
	newConn := &Conn{
		Conn:      conn,
		bufReader: bufio.NewReader(conn),
	}
	defer newConn.Close()

	// Print connection details
	if newConn.LocalAddr() == nil {
		log.Fatal("couldn't retrieve local address")
	}
	log.Printf("local address: %q", newConn.LocalAddr().String())

	if newConn.RemoteAddr() == nil {
		log.Fatal("couldn't retrieve remote address")
	}
	log.Printf("remote address: %q", newConn.RemoteAddr().String())
}

func (p *Conn) LocalAddr() net.Addr {
	p.once.Do(func() { p.readErr = p.readHeader() })
	if p.header == nil || p.header.Command.IsLocal() || p.readErr != nil {
		return p.Conn.LocalAddr()
	}

	return p.header.DestinationAddr
}

func (p *Conn) RemoteAddr() net.Addr {
	p.once.Do(func() { p.readErr = p.readHeader() })
	if p.header == nil || p.header.Command.IsLocal() || p.readErr != nil {
		return p.Conn.RemoteAddr()
	}

	return p.header.SourceAddr
}

func (p *Conn) readHeader() error {
	header, err := Read(p.bufReader)

	p.header = header
	return err
}

func Read(reader *bufio.Reader) (*Header, error) {
	b1, err := reader.Peek(1)
	log.Printf("[Read Head] b1: %q, err: %v\n", b1, err)

	if err != nil {
		if err == io.EOF {
			return nil, ErrNoProxyProtocol
		}
		return nil, err
	}

	if bytes.Equal(b1[:1], SIGV1[:1]) || bytes.Equal(b1[:1], SIGV2[:1]) {
		signature, err := reader.Peek(5)
		if err != nil {
			if err == io.EOF {
				return nil, ErrNoProxyProtocol
			}
			return nil, err
		}
		if bytes.Equal(signature[:5], SIGV1) {
			return parseVersion1(reader)
		}

		signature, err = reader.Peek(12)
		if err != nil {
			if err == io.EOF {
				return nil, ErrNoProxyProtocol
			}
			return nil, err
		}
		if bytes.Equal(signature[:12], SIGV2) {
			return parseVersion2(reader)
		}
	}

	return nil, ErrNoProxyProtocol
}
