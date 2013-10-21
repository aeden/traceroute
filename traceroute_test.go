package traceroute

import (
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

}
