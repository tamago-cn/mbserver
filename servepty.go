package mbserver

import (
	"io"
	"log"
	"os"

	"github.com/creack/pty"
)

// ListenPTY starts the Modbus server listening to a serial device.
// For example:  err := s.ListenPTY()
func (s *Server) ListenPTY() (*os.File, error) {
	master, slave, err := pty.Open()
	if err != nil {
		return nil, err
	}
	s.ptys = append(s.ptys, master)
	s.ttys = append(s.ttys, slave)
	go s.acceptPtyRequests(master)
	return slave, nil
}

func (s *Server) acceptPtyRequests(port io.ReadWriteCloser) {
	for {
		buffer := make([]byte, 512)

		bytesRead, err := port.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return
		}

		if bytesRead != 0 {

			// Set the length of the packet to the number of read bytes.
			packet := buffer[:bytesRead]

			log.Printf("master recv: %v\n", packet)

			frame, err := NewRTUFrame(packet)
			if err != nil {
				log.Printf("bad serial frame error %v\n", err)
				return
			}

			request := &Request{port, frame}

			s.requestChan <- request
		}
	}
}
