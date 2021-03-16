package mbserver

import (
	"io"
	"log"
	"net"
)

// ListenRTUonTCP starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTUonTCP("0.0.0.0:21000")
func (m *MultiServer) ListenRTUonTCP(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				continue
			}
			go m.acceptRTUTCPRequests(conn)
		}
	}()
	return nil
}

func (m *MultiServer) acceptRTUTCPRequests(conn io.ReadWriteCloser) {
	defer conn.Close()
	for {
		buffer := make([]byte, 512)

		bytesRead, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return
		}

		if bytesRead != 0 {

			// Set the length of the packet to the number of read bytes.
			packet := buffer[:bytesRead]

			frame, err := NewRTUFrame(packet)
			if err != nil {
				log.Printf("bad serial frame error %v\n", err)
				return
			}

			request := &Request{conn, frame}

			m.requestChan <- request
		}
	}
}
