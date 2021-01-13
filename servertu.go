package mbserver

import (
	"errors"
	"io"
	"log"
	"strings"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Printf("failed to open %s: %v\n", serialConfig.Address, err)
		return err
	}
	s.ports = append(s.ports, port)
	go s.acceptSerialRequests(port)
	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) {
	for {
		packet, err := readRTU(port)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				continue
			}
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return
		}

		frame, err := NewRTUFrame(packet)
		if err != nil {
			log.Printf("bad serial frame error %v\n", err)
			continue
		}

		request := &Request{port, frame}

		s.requestChan <- request
	}
}

func readRTU(port serial.Port) ([]byte, error) {
	buffer := make([]byte, 0, 512)
	bytesRead := 0
	for {
		addr := byte(0)
		{
			buf := make([]byte, 1)
			n, err := port.Read(buf)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				return nil, errors.New("read addr error")
			}
			addr = buf[0]
			buffer = append(buffer, addr)
			bytesRead++
		}
		cmdID := byte(0)
		{
			buf := make([]byte, 1)
			n, err := port.Read(buf)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				return nil, errors.New("read cmd_id error")
			}
			cmdID = buf[0]
			buffer = append(buffer, cmdID)
			bytesRead++
		}
		if cmdID >= byte(1) && cmdID <= byte(6) {
			for bytesRead < 8 {
				buf := make([]byte, 1)
				n, err := port.Read(buf)
				if err != nil {
					return nil, err
				}
				if n == 0 {
					return nil, errors.New("read data error")
				}
				buffer = append(buffer, buf[0])
				bytesRead++
			}
			return buffer[:bytesRead], nil
		}

		for bytesRead < 7 {
			buf := make([]byte, 1)
			n, err := port.Read(buf)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				return nil, errors.New("read data error")
			}
			buffer = append(buffer, buf[0])
			bytesRead++
		}

		byteCount := int(buffer[bytesRead-1])
		for bytesRead < 9+byteCount {
			buf := make([]byte, 1)
			n, err := port.Read(buf)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				return nil, errors.New("read data error")
			}
			buffer = append(buffer, buf[0])
			bytesRead++
		}
		return buffer[:bytesRead], nil
	}
}
