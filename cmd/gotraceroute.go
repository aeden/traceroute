package main

import (
	"flag"
	"fmt"
	"github.com/aeden/traceroute"
	"net"
)

func printHop(hop traceroute.TracerouteHop) {
	addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
	hostOrAddr := addr
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}
        if hop.Success {
          fmt.Printf("%-3d %v (%v)  %v\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime)
        } else {
          fmt.Printf("%-3d *\n", hop.TTL)
        }
}

func address(address [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
}

func main() {
	flag.Parse()
	host := flag.Arg(0)
        options := traceroute.TracerouteOptions{}
        options.SetOpt("retries", 0)

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return
	}

	fmt.Printf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, ipAddr, options.MaxHops(), options.PacketSize())

	c := make(chan traceroute.TracerouteHop, 0)
	go func() {
		for {
			hop, ok := <-c
			if !ok {
				fmt.Println()
				return
			}
			printHop(hop)
		}
	}()

	_, err = traceroute.Traceroute(host, &options, c)
	if err != nil {
		fmt.Printf("Error: ", err)
	}
}
