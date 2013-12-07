// Package traceroute provides functions for executing a tracroute to a remote
// host.
package traceroute

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"
)

// Return the first non-loopback address as a 4 byte IP address. This address
// is used for sending packets out.
func socketAddr() (addr [4]byte, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if len(ipnet.IP.To4()) == net.IPv4len {
				copy(addr[:], ipnet.IP.To4())
				return
			}
		}
	}
	err = errors.New("You do not appear to be connected to the Internet")
	return
}

// Given a host name convert it to a 4 byte IP address.
func destAddr(dest string) (destAddr [4]byte, err error) {
	addrs, err := net.LookupHost(dest)
	if err != nil {
		return
	}
	addr := addrs[0]

	ipAddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return
	}
	copy(destAddr[:], ipAddr.IP.To4())
	return
}

type TracerouteOptions struct {
	Port       int
	MaxHops    int
	TimeoutMs  int64
	Retries    int
	PacketSize int
}

type TracerouteHop struct {
	Address     [4]byte
	Host        string
	N           int
	ElapsedTime time.Duration
}

func (hop *TracerouteHop) AddressString() string {
	return fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
}

type TracerouteResult struct {
	DestinationAddress [4]byte
	Hops               []TracerouteHop
}

func defaultOptions(options *TracerouteOptions) {
	if options.Port == 0 {
		options.Port = 33434
	}
	if options.MaxHops == 0 {
		options.MaxHops = 64
	}
	if options.TimeoutMs == 0 {
		options.TimeoutMs = 1000
	}
	if options.Retries == 0 {
		options.Retries = 3
	}
	if options.PacketSize == 0 {
		options.PacketSize = 52
	}
}

// Traceroute uses the given dest (hostname) and options to execute a traceroute
// from your machine to the remote host.
//
// Outbound packets are UDP packets and inbound packets are ICMP.
//
// Returns a TracerouteResult which contains an array of hops. Each hop includes
// the elapsed time and its IP address.
func Traceroute(dest string, options *TracerouteOptions) (result TracerouteResult, err error) {
	result.Hops = []TracerouteHop{}
	defaultOptions(options)
	destAddr, err := destAddr(dest)
	result.DestinationAddress = destAddr
	socketAddr, err := socketAddr()
	if err != nil {
		return
	}

	tv := syscall.NsecToTimeval(1000 * 1000 * options.TimeoutMs)

	ttl := 1
	retry := 0
	for {
		start := time.Now()

		// Set up the socket to receive inbound packets
		recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
		if err != nil {
			return result, err
		}

		// Set up the socket to send packets out.
		sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			return result, err
		}
		// This sets the current hop TTL
		syscall.SetsockoptInt(sendSocket, 0x0, syscall.IP_TTL, ttl)
		// This sets the timeout to wait for a response from the remote host
		syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

		defer syscall.Close(recvSocket)
		defer syscall.Close(sendSocket)

		// Bind to the local socket to listen for ICMP packets
		syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: options.Port, Addr: socketAddr})

		// Send a single null byte UDP packet
		syscall.Sendto(sendSocket, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: options.Port, Addr: destAddr})

		var p = make([]byte, options.PacketSize)
		n, from, err := syscall.Recvfrom(recvSocket, p, 0)
		elapsed := time.Since(start)
		if err == nil {
			currAddr := from.(*syscall.SockaddrInet4).Addr
			hop := TracerouteHop{Address: currAddr, N: n, ElapsedTime: elapsed}

			currHost, err := net.LookupAddr(hop.AddressString())
			if err == nil {
				hop.Host = currHost[0]
			}

			result.Hops = append(result.Hops, hop)
			//log.Println("Received n=", n, ", from=", currAddr, ", t=", elapsed)

			ttl += 1
			retry = 0

			if ttl > options.MaxHops || currAddr == destAddr {
				return result, nil
			}
		} else {
			retry += 1
			if retry > options.Retries {
				ttl += 1
				retry = 0
			}
		}

	}
}
