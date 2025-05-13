package gnssgo

import (
	"testing"
)

func TestSimple(t *testing.T) {
	// A simple test to verify that Go can find and run tests in this directory
	if VER_GNSSGO == "" {
		t.Error("VER_GNSSGO should not be empty")
	}
}
