package traceroute

import (
	"errors"
	"log"
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
				log.Println("Found IP address: ", ipnet.IP.String())
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
	log.Println("Destination address: ", addr)

	ipAddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return
	}
	copy(destAddr[:], ipAddr.IP.To4())
	return
}

type TracerouteOptions struct {
	Port      int
	MaxHops   int
	TimeoutMs int64
	Retries   int
}

func defaultOptions(options *TracerouteOptions) {
	if options.Port == 0 {
		options.Port = 33434
	}
	if options.MaxHops == 0 {
		options.MaxHops = 30
	}
	if options.TimeoutMs == 0 {
		options.TimeoutMs = 1000
	}
	if options.Retries == 0 {
		options.Retries = 3
	}
}

func Traceroute(dest string, options *TracerouteOptions) (result string, err error) {
	defaultOptions(options)
	destAddr, err := destAddr(dest)
	socketAddr, err := socketAddr()
	tv := syscall.NsecToTimeval(1000 * 1000 * options.TimeoutMs)
	if err != nil {
		return
	}

	ttl := 1
	retry := 0
	for {
		start := time.Now()

		recvSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
		if err != nil {
			return "", err
		}

		sendSocket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			return "", err
		}
		syscall.SetsockoptInt(sendSocket, 0x0, syscall.IP_TTL, ttl)
		syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

		defer syscall.Close(recvSocket)
		defer syscall.Close(sendSocket)

		syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: options.Port, Addr: socketAddr})
		syscall.Sendto(sendSocket, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: options.Port, Addr: destAddr})

		var p = make([]byte, 512)
		n, from, err := syscall.Recvfrom(recvSocket, p, 0)
		elapsed := time.Since(start)
		if err == nil {
			currAddr := from.(*syscall.SockaddrInet4).Addr
			log.Println("Received n=", n, ", from=", currAddr, ", t=", elapsed)

			ttl += 1
			retry = 0

			if ttl > options.MaxHops || currAddr == destAddr {
				return "", nil
			}
		} else {
			retry += 1
			log.Print("* ")
			if retry > options.Retries {
				ttl += 1
				retry = 0
			}
		}

	}
}
