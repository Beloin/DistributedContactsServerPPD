package tcpserver

import (
	"bytes"
	"distributed_contacts_server/internal/clock"
	"distributed_contacts_server/internal/data"
	"distributed_contacts_server/internal/parser"
	"fmt"
	"net"
	"os"
	"strings"
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
	fmt.Println("[LISTEN] Listening on " + addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("[LISTEN] Error accepting: ", err.Error())
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
		addr := conn.RemoteAddr().String()
		fmt.Printf("[LISTEN] Recieved connection from %s\n", addr)

		read, buf, err := readAll(conn, 1)
		if err != nil || read < 1 {
			fmt.Println("[LISTEN] Could not get identity from " + addr)
			conn.Close()
			continue
		}

		identity := buf[0]
		read, buf, err = readAll(conn, 256)
		if err != nil || read != 256 {
			fmt.Println("[LISTEN] Could not get name from " + addr)
			conn.Close()
			continue
		}

		fmt.Printf("[LISTEN] Handshake done with %s\n", addr)
		name := parser.ReadTillNull(buf)
		// Go routine to ping and other to listen
		if identity == 1 {
			hostAndPort := strings.Split(addr, ":")
			otherHost, otherPort := hostAndPort[0], hostAndPort[1]
			data.AddServer(otherHost, otherPort, name)
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
			fmt.Printf("[SERVER] Lost connection to server from %s\n", name)
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
			fmt.Printf("[SERVER] Undefined command from %s\n", name)
		}

		if err != nil {
			fmt.Printf("[SERVER] Lost connection to server from %s\n", name)
			data.Disconnect(name)
			return
		}
	}
}

// Function to ping server sending heartbeat
func pingServerLoop(otherServer string, conn net.Conn) {
	retries := 0
	for {
		fmt.Printf("[CONNECT] Recv Ping from %s\n", otherServer)
		time.Sleep(time.Second * 5)
		n, err := conn.Write([]byte{3})

		if n < 1 || err != nil {
			retries++
			fmt.Printf("[SERVER] Could not ping server %s\n", otherServer)

			if retries > 3 {
				return
			}

			continue
		}

		bts := make([]byte, 4)
		currentClock := clock.CurrentClock.Load()
		parser.Parse32Bits(currentClock, &bts)
		n, err = conn.Write(bts)

		if n < 1 || err != nil {
			retries++
			fmt.Printf("[SERVER] Could not ping server %s\n", otherServer)

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

		// TODO: After recv each, send for other servers the same data
		command := buff[0]
		switch command {
		case 1:
			err = recvClientUpdateCommand(name, &conn)
		case 2:
			err = recvClientDeleteCommand(name, &conn)
		case 3:
			err = recvClientListAllCommand(name, &conn)
		default:
			fmt.Printf("[CLIENT] Command not implemented from %s\n", name)
		}

		if err != nil {
			fmt.Printf("[SERVER] Lost connection to server from %s\n", name)
			data.Disconnect(name)
			return
		}
	}
}

func readAll(conn net.Conn, size int) (int, []byte, error) {
	var received int

	buffer := new(bytes.Buffer)
	for {
		chunk := make([]byte, size)
		read, err := conn.Read(chunk)
		if err != nil {
			return received, buffer.Bytes(), err
		}

		received += read
		buffer.Write(chunk)

		if read == 0 || received >= size {
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
	status := buff[0]

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

func recvClientUpdateCommand(name string, conn *net.Conn) error {
	_, buff, err := readAll(*conn, 256)
	if err != nil {
		return err
	}
	contactName := parser.ReadTillNull(buff)

	_, buff, err = readAll(*conn, 10)
	if err != nil {
		return err
	}
	number := parser.ReadTillNull(buff)

	data.AddContact(name, contactName, number)

	return nil
}

func recvClientDeleteCommand(name string, conn *net.Conn) error {
	_, buff, err := readAll(*conn, 256)
	if err != nil {
		return err
	}
	contactName := parser.ReadTillNull(buff)

	data.RemoveContact(name, contactName)

	return nil
}

func recvClientListAllCommand(name string, conn *net.Conn) error {
	// TODO: Return the lis of clients
	return nil
}
