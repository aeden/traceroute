package traceroute

import (
	"errors"
	"net"
	"syscall"
	"time"
)

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
	N           int
	ElapsedTime time.Duration
}

type TracerouteResult struct {
	Hops []TracerouteHop
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

		recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
		if err != nil {
			return result, err
		}

		sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			return result, err
		}
		syscall.SetsockoptInt(sendSocket, 0x0, syscall.IP_TTL, ttl)
		syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

		defer syscall.Close(recvSocket)
		defer syscall.Close(sendSocket)

		syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: options.Port, Addr: socketAddr})
		syscall.Sendto(sendSocket, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: options.Port, Addr: destAddr})

		var p = make([]byte, options.PacketSize)
		n, from, err := syscall.Recvfrom(recvSocket, p, 0)
		elapsed := time.Since(start)
		if err == nil {
			currAddr := from.(*syscall.SockaddrInet4).Addr
			result.Hops = append(result.Hops, TracerouteHop{Address: currAddr, N: n, ElapsedTime: elapsed})
			//log.Println("Received n=", n, ", from=", currAddr, ", t=", elapsed)

			ttl += 1
			retry = 0

			if ttl > options.MaxHops || currAddr == destAddr {
				return result, nil
			}
		} else {
			retry += 1
			//log.Print("* ")
			if retry > options.Retries {
				ttl += 1
				retry = 0
			}
		}

	}
}
