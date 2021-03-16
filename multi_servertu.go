package mbserver

import (
	"io"
	"log"
	"strings"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (m *MultiServer) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Printf("failed to open %s: %v\n", serialConfig.Address, err)
		return err
	}
	m.ports = append(m.ports, port)
	go m.acceptSerialRequests(port)
	return err
}

func (m *MultiServer) acceptSerialRequests(port serial.Port) {
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

		m.requestChan <- request
		//}
	}
}
