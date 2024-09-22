package tcpserver

import (
	"bytes"
	"distributed_contacts_server/internal/clock"
	"distributed_contacts_server/internal/data"
	"distributed_contacts_server/internal/parser"
	"fmt"
	"net"
	"os"
	"time"
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
		// Go routine to ping and other to listen
		if identity == 1 {
			go serverLoop(name, conn)
			go pingServerLoop(name, conn)
		} else { // Treat if != 2?
			go clientLoop(name, conn)
		}
	}
}

// Server loop
func serverLoop(name string, conn net.Conn) {
	for {
		n, buff, err := readAll(conn, 1)
		if n < 1 || err != nil {
			fmt.Printf("[SERVER](%s) Lost connection to server\n", name)
			data.Disconnect(name)
			return
		}

		switch buff[0] {
		case 1:
			err = recvServerUpdateCommand(&conn)
		case 2:
			err = recvServerDeleteCommand(&conn)
		case 3:
			err = recvServerPingCommand(name, &conn)
		case 4:
			// TODO: Implement ClockUpdate
		case 5:
			err = recvServerAskForUpdateCommand(&conn)
		default:
			fmt.Printf("[SERVER](%s) Undefined command", name)
		}

		if err != nil {
			fmt.Printf("[SERVER](%s) Lost connection to server\n", name)
			data.Disconnect(name)
			return
		}
	}
}

// Function to ping server sending heartbeat
func pingServerLoop(name string, conn net.Conn) {
	retries := 0
	for {
		time.Sleep(time.Second * 5)
		n, err := conn.Write([]byte{3})

		if n < 1 || err != nil {
			retries++
			fmt.Printf("[SERVER](%s) Could not ping server\n", name)

			if retries > 3 {
				return
			}

			continue
		}

		bts := make([]byte, 4)
		parser.Parse32Bits(clock.CurrentClock.Load(), &bts)
		n, err = conn.Write(bts)

		if n < 1 || err != nil {
			retries++
			fmt.Printf("[SERVER](%s) Could not ping server\n", name)

			if retries > 3 {
				return
			}

		}
	}
}

// Client loop
func clientLoop(name string, conn net.Conn) {
	for {
		n, buff, err := readAll(conn, 1)
		if n < 1 || err != nil {
			fmt.Printf("[CLIENT](%s) Lost connection\n", name)
			return
		}

		command := buff[0]
		switch command {
		case 1:
			err = recvClientUpdateCommand(name, conn)
		case 2:
		case 3:
		default:
			fmt.Printf("[CLIENT](%s) Command not implemented\n", name)
		}
	}
}

func readAll(conn net.Conn, size int) (int, []byte, error) {
	var received int

	// buffer := bytes.NewBuffer(nil)
	buffer := new(bytes.Buffer)
	for {
		// TODO: Test this
		chunk := make([]byte, DEFAULT_BUFFER_RECV)
		read, err := conn.Read(chunk)
		if err != nil {
			return received, buffer.Bytes(), err
		}

		received += read
		buffer.Write(chunk)

		if read == 0 || read < size {
			break
		}
	}

	return received, buffer.Bytes(), nil
}

func recvServerUpdateCommand(conn *net.Conn) error {
	_, buff, err := readAll(*conn, 4)
	if err != nil {
		return err
	}
	serverClock := parser.ParseTo32Bits(buff)

	_, buff, err = readAll(*conn, 256)
	if err != nil {
		return err
	}
	name := parser.ReadTillNull(buff)

	_, buff, err = readAll(*conn, 256)
	if err != nil {
		return err
	}
	contactName := parser.ReadTillNull(buff)

	_, buff, err = readAll(*conn, 10)
	if err != nil {
		return err
	}
	number := parser.ReadTillNull(buff)

	data.CompareAndUpdateContact(name, contactName, number, serverClock)

	return nil
}

func recvServerDeleteCommand(conn *net.Conn) error {
	_, buff, err := readAll(*conn, 4)
	if err != nil {
		return err
	}
	serverClock := parser.ParseTo32Bits(buff)

	_, buff, err = readAll(*conn, 256)
	if err != nil {
		return err
	}
	name := parser.ReadTillNull(buff)

	_, buff, err = readAll(*conn, 256)
	if err != nil {
		return err
	}
	contactName := parser.ReadTillNull(buff)

	data.CompareAndDeleteContact(name, contactName, serverClock)

	return nil
}

func recvServerPingCommand(name string, conn *net.Conn) error {
	_, buff, err := readAll(*conn, 1)
	if err != nil {
		return err
	}
	status := parser.ParseTo32Bits(buff)

	_, buff, err = readAll(*conn, 4)
	if err != nil {
		return err
	}
	serverClock := parser.ParseTo32Bits(buff)

	// TODO: See current clock and update external server if My clock >>
	data.Pong(name, serverClock, status == 1)

	return nil
}

func recvServerAskForUpdateCommand(conn *net.Conn) error {
	// TODO: Implement recvServerAskForUpdateCommand
	return nil
}

func recvClientUpdateCommand(name string, conn *net.Conn) {
}
