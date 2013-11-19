package main

import (
	"fmt"
	"github.com/aeden/traceroute"
)

func main() {
	out, err := traceroute.Traceroute("google.com", new(traceroute.TracerouteOptions))
	if err != nil {
		fmt.Printf("Error: ", err)
	}

	for i, hop := range out.Hops {
		addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
		fmt.Printf("%-3d %v (%v)  %v\n", i, addr, addr, hop.ElapsedTime)
	}
}
