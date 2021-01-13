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
		//buffer := make([]byte, 512)

		//bytesRead, err := port.Read(buffer)
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

		//if bytesRead != 0 {

		// Set the length of the packet to the number of read bytes.
		//packet := buffer[:bytesRead]

		frame, err := NewRTUFrame(packet)
		if err != nil {
			log.Printf("bad serial frame error %v\n", err)
			continue
		}

		request := &Request{port, frame}

		s.requestChan <- request
		//}
	}
}

func readRTU(port serial.Port) ([]byte, error) {
	buffer := make([]byte, 512)
	bytesRead := 0

	n, err := readBytes(port, buffer, bytesRead, 2)
	if err != nil {
		return nil, err
	}
	bytesRead += n

	cmdID := buffer[1]

	switch cmdID {
	case byte(1), byte(2), byte(3), byte(4), byte(5), byte(6):
		n, err = readBytes(port, buffer, bytesRead, 6)
		if err != nil {
			return nil, err
		}
		bytesRead += n
		return buffer, nil
	case byte(15), byte(16):
		n, err = readBytes(port, buffer, bytesRead, 5)
		if err != nil {
			return nil, err
		}
		bytesRead += n
		n, err = readBytes(port, buffer, bytesRead, int(buffer[bytesRead-1])+2)
		if err != nil {
			return nil, err
		}
		bytesRead += n
		return buffer[:bytesRead], nil
	}

	return buffer[0:bytesRead], nil
}

func readBytes(port serial.Port, buffer []byte, base int, expect int) (int, error) {
	n := 0
	for i := 0; i < expect; i++ {
		buf := make([]byte, 1)
		x, err := port.Read(buf)
		if err != nil {
			return 0, err
		}
		if x == 0 {
			return 0, errors.New("read error")
		}
		buffer[base+n] = buf[0]
		n++
	}
	return n, nil
}
