package mbserver

import (
	"io"
	"log"
	"net"
	"strings"
)

func (m *MultiServer) accept(listen net.Listener) error {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Printf("Unable to accept connections: %#v\n", err)
			return err
		}

		go func(conn net.Conn) {
			defer conn.Close()

			for {
				packet := make([]byte, 512)
				bytesRead, err := conn.Read(packet)
				if err != nil {
					if err != io.EOF {
						log.Printf("read error %v\n", err)
					}
					return
				}
				// Set the length of the packet to the number of read bytes.
				packet = packet[:bytesRead]

				frame, err := NewTCPFrame(packet)
				if err != nil {
					log.Printf("bad packet error %v\n", err)
					return
				}

				request := &Request{conn, frame}

				m.requestChan <- request
			}
		}(conn)
	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (m *MultiServer) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Printf("Failed to Listen: %v\n", err)
		return err
	}
	m.listeners = append(m.listeners, listen)
	go m.accept(listen)
	return err
}
