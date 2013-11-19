package main

import (
	"flag"
	"fmt"
	"github.com/aeden/traceroute"
)

func address(address [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
}

func main() {
	flag.Parse()
	host := flag.Arg(0)
	options := traceroute.TracerouteOptions{}

	result, err := traceroute.Traceroute(host, &options)

	fmt.Printf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, address(result.DestinationAddress), options.MaxHops, options.PacketSize)

	if err != nil {
		fmt.Printf("Error: ", err)
	}

	for i, hop := range result.Hops {
		addr := address(hop.Address)
		fmt.Printf("%-3d %v (%v)  %v\n", i, addr, addr, hop.ElapsedTime)
	}
}
