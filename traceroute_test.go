package traceroute

import (
	"fmt"
	"testing"
)

func TestTraceroute(t *testing.T) {
	out, err := Traceroute("google.com", new(TracerouteOptions))
	if err == nil {
		if len(out.Hops) == 0 {
			t.Errorf("TestTraceroute failed. Expected at least one hop")
		}
	} else {
		t.Errorf("TestTraceroute failed due to an error: %v", err)
	}

	for i, hop := range out.Hops {
		addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
		hostOrAddr := addr
		if hop.Host != "" {
			hostOrAddr = hop.Host
		}
		fmt.Printf("%-3d %v (%v)  %v\n", i, hostOrAddr, addr, hop.ElapsedTime)
	}

}
