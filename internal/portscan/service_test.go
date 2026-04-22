package portscan

import (
	"reflect"
	"testing"
)

func TestParsePorts(t *testing.T) {
	ports, err := ParsePorts("22,80,443,8000-8002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []int{22, 80, 443, 8000, 8001, 8002}
	if !reflect.DeepEqual(ports, expected) {
		t.Fatalf("expected %v, got %v", expected, ports)
	}
}
