package traceroute

import (
	"testing"
)

func TestTraceroute(t *testing.T) {
	out, err := Traceroute("google.com", new(TracerouteOptions))
	if err == nil {
		expected := ""
		if out != expected {
			t.Errorf("TestTraceroute failed. Expected %v, got %v", expected, out)
		}
	} else {
		t.Errorf("TestTraceroute failed due to an error: %v", err)
	}

}
