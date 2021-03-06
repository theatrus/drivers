package tplink

import (
	"testing"
	"time"

	"github.com/reef-pi/hal"
)

type mockConn struct {
	Buffer []byte
}

func (c *mockConn) Close() error { return nil }
func (c *mockConn) Read(buf []byte) (int, error) {
	return len(buf), nil
}
func (c *mockConn) SetDeadline(_ time.Time) error { return nil }
func (c *mockConn) Write(_ []byte) (int, error)   { return 0, nil }
func mockConnFacctory(_, _ string, _ time.Duration) (Conn, error) {
	return &mockConn{}, nil
}

func TestHS103Plug(t *testing.T) {
	p := NewHS103Plug("127.0.0.1:9999")
	p.cnFactory = mockConnFacctory
	if err := p.On(); err != nil {
		t.Error(err)
	}
	if err := p.Off(); err != nil {
		t.Error(err)
	}

	d, err := HS103HALAdapter([]byte(`{"address":"127.0.0.1:3000"}`), nil)
	if err != nil {
		t.Error(err)
	}
	if d.Metadata().Name == "" {
		t.Error("HAL metadata should not have empty name")
	}

	d1 := d.(hal.DigitalOutputDriver)

	if len(d1.DigitalOutputPins()) != 1 {
		t.Error("Expected exactly one output pin")
	}
	pin, err := d1.DigitalOutputPin(0)
	if err != nil {
		t.Error(err)
	}
	if pin.LastState() != false {
		t.Error("Expected initial state to be false")
	}
}
