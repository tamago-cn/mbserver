// Package mbserver implments a Modbus server (slave).
package mbserver

import (
	"net"
	"os"

	"github.com/goburrow/serial"
)

// MultiServer is a milti Modbus slave with allocated memory for discrete inputs, coils, etc.
type MultiServer struct {
	servers     [256]*Server
	requestChan chan *Request

	listeners []net.Listener
	ports     []serial.Port
	ptys      []*os.File
	ttys      []*os.File
}

// NewMultiServer new multi server
func NewMultiServer() *MultiServer {
	m := &MultiServer{}

	for i := 0; i < 256; i++ {
		m.servers[i] = NewServer()
	}
	m.requestChan = make(chan *Request)

	go m.handler()

	return m
}

func (m *MultiServer) handle(request *Request) Framer {
	var exception *Exception
	var data []byte

	response := request.frame.Copy()

	addr := request.frame.GetAddr()
	s := m.servers[int(addr)]

	if s != nil {
		function := request.frame.GetFunction()
		if s.function[function] != nil {
			data, exception = s.function[function](s, request.frame)
			response.SetData(data)
		} else {
			exception = &IllegalFunction
		}

	} else {
		exception = &GatewayTargetDeviceFailedtoRespond
	}
	if exception != &Success {
		response.SetException(exception)
	}

	return response
}

// All requests are handled synchronously to prevent modbus memory corruption.
func (m *MultiServer) handler() {
	for {
		request := <-m.requestChan
		response := m.handle(request)
		request.conn.Write(response.Bytes())
	}
}

// Remove stops listening to TCP/IP ports and closes serial ports.
func (m *MultiServer) Remove(addr uint8) {
	if m.servers[int(addr)] != nil {
		m.servers[int(addr)].Close()
	}
	m.servers[int(addr)] = nil
}

// Close stops listening to TCP/IP ports and closes serial ports.
func (m *MultiServer) Close() {
	for _, s := range m.servers {
		s.Close()
	}
}
