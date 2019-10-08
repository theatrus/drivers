package socket

import (
	"bufio"
	"github.com/reef-pi/hal"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout = 5 * time.Second

	commandPins = "list_pins"
)

type socketDriver struct {
	endpoint string
	socket net.Conn
	lock sync.Mutex
}

func New(endpoint string) (hal.Driver, error) {
	sd := &socketDriver{
		endpoint: endpoint,
	}
	err := sd.ensureConnection()
	return sd, err
}

func (s *socketDriver) ensureConnection() error {
	if s.socket != nil {
		s.socket.Close()
		s.socket = nil
	}
	var conn net.Conn
	var err error
	if strings.HasPrefix("/", s.endpoint) {
		conn, err = net.Dial("unix", s.endpoint)
	} else {
		conn, err = net.DialTimeout("tcp", s.endpoint, defaultTimeout)
	}
	if err != nil {
		return err
	}
	s.socket = conn
	return nil
}

// closeSocket closes the underlying socket and reset the socket
// to be re-opened on the next command response pair
func (s *socketDriver) closeSocket() {
	if s.socket != nil {
		s.socket.Close()
	}
	s.socket = nil
}

// commandResponse handles a two way communication session
// with the driver socket, writing a line of output and receiving a line of
// input. The command should not have a terminating newline.
func (s *socketDriver) commandResponse(command string) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.ensureConnection()
	if err != nil {
		return "", err
	}
	s.socket.SetWriteDeadline(time.Now().Add(defaultTimeout))
	w := bufio.NewWriter(s.socket)
	_, err = w.WriteString(command + "\n")
	if err != nil {
		s.closeSocket()
		return "", err
	}
	err = w.Flush()
	if err != nil {
		s.closeSocket()
		return "", err
	}
	r := bufio.NewReader(s.socket)
	response, err := r.ReadString('\n')
	if err != nil {
		s.closeSocket()
		return "", err
	}
	return strings.TrimSpace(response), nil
}

func (s *socketDriver) Close() error { s.closeSocket(); return nil }
func (s *socketDriver) Metadata() hal.Metadata {
	return hal.Metadata{
		Name:         "socket_driver",
		Description:  "Connects to a socket and delegates all driver operations to a remote program",
		Capabilities: []hal.Capability{hal.AnalogInput, hal.DigitalInput, hal.DigitalOutput, hal.PWM},
	}
}
func (s *socketDriver) Pins(capability hal.Capability) ([]hal.Pin, error) {
	_, err := s.commandResponse(commandPins)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}
