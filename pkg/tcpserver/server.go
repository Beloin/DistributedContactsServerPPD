package tcpserver

import (
	"bytes"
	"fmt"
	"net"
	"os"
  "distributed_contacts_server/internal/parser"
)

const (
	CONN_TYPE           = "tcp"
	DEFAULT_BUFFER_RECV = 256
)

func Listen(host string, port string) {
	addr := host + ":" + port
	l, err := net.Listen(CONN_TYPE, addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	fmt.Println("Listening on " + addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// TODO: Handle in another thread?
		// How to handle here and save data here?
		// We can create a map and add the host + port as key
		// then in the request it should return a name of this server
		// Maybe even read the first information regarding the SERVERxCLIENT thing
		// listenClient(clientName);
		// listenServer(serverName);
		// Add callbacks too?
		addr := conn.LocalAddr().String()
		read, buf, err := readAll(conn, 1)
		if err != nil || read < 1 {
			fmt.Println("Could not get identity from" + addr)
			conn.Close()
			continue
		}

		identity := buf[0]
		read, buf, err = readAll(conn, 256)
		if err != nil || read != 256 {
			fmt.Println("Could not get name from" + addr)
			conn.Close()
			continue
		}

    name := parser.ReadTillNull(buf)
		if identity == 1 {
		} else { // Treat if != 2?

		}

		go handleRequest(conn)
	}
}

// Server loop
func handleServer(name string) {
	for {
	}
}

// Client loop
func handleClient(name string) {
	for {
	}
}

// Handles incoming requests.
// TODO: Add here validation for CLIENT and SERVER
// Add param entry to be the entry of the connection map?
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)

	reqLen, bytes, err := readAll(conn)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		conn.Close()
	}
	conn.Write([]byte("Message received."))
	conn.Close()
}

func readAll(conn net.Conn, len int) (int, []byte, error) {
	var received int

	// buffer := bytes.NewBuffer(nil)
	buffer := new(bytes.Buffer)
	for {
		chunk := make([]byte, DEFAULT_BUFFER_RECV)
		read, err := conn.Read(chunk)
		if err != nil {
			return received, buffer.Bytes(), err
		}

		received += read
		buffer.Write(chunk)

		if read == 0 || read < DEFAULT_BUFFER_RECV {
			break
		}
	}

	return received, buffer.Bytes(), nil
}
