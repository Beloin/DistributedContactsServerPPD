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
	CONN_TYPE = "tcp"
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

		fmt.Println("[LISTEN] Sending my name to " + addr)
		var buffer []byte
		_ = parser.ParseString(host, &buffer)
		_, err = conn.Write(buffer)
		if err != nil {
			fmt.Println("[LISTEN] Could not send name to " + addr)
			conn.Close()
			continue
		}

		fmt.Printf("[LISTEN] Handshake done with %s\n", addr)
		name := parser.ReadTillNull(buf)

		// Go routine to ping and other to listen
		if identity == 1 {
			hostAndPort := strings.Split(addr, ":")
			otherHost, otherPort := hostAndPort[0], hostAndPort[1]
			data.AddServer(otherHost, otherPort, name, conn)
			// TODO: Send a response to client so he can know there's new server
			go serverLoop(name, conn)
			go pingServerLoop(name, conn)
		} else { // Treat if != 2?
			data.AddClient(name)
			go clientLoop(name, conn)
		}
	}
}

// Server loop
func serverLoop(name string, conn net.Conn) {
	fmt.Printf("[SERVER] Started server loop %s\n", name)
	for {
		n, buff, err := readAll(conn, 1)

		if n < 1 || err != nil {
			fmt.Printf("[SERVER] Lost connection to server from %s\n", name)
			data.Disconnect(name)
			return
		}

		fmt.Printf("[SERVER] Recv message from %s\n", name)
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
		time.Sleep(time.Second * 5)
		fmt.Printf("[SERVER] Pinging %s\n", otherServer)
		n, err := conn.Write([]byte{3})

		if n < 1 || err != nil {
			retries++
			fmt.Printf("[SERVER] Could not ping server %s\n", otherServer)

			if retries > 3 {
				return
			}

			continue
		}

		n, err = conn.Write([]byte{1})
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
	fmt.Printf("[CLIENT] Started Client loop for %s\n", name)
	for {
		n, buff, err := readAll(conn, 1)
		if n < 1 || err != nil {
			fmt.Printf("[CLIENT] Lost connection with %s\n", name)
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
			fmt.Printf("[SERVER] Lost connection to client from %s\n", name)
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

	_, buff, err = readAll(*conn, 20)
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
	fmt.Printf("[CONNECT] Recv Ping from %s\n", name)
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

	data.Pong(name, serverClock, status == 1)

	return nil
}

func recvServerAskForUpdateCommand(conn *net.Conn) error {
	all, fullAmount := data.ListAll()
	connection := *conn
	bts := make([]byte, 4)
	parser.Parse32Bits(uint32(fullAmount), &bts)
	_, err := connection.Write(bts)
	if err != nil {
		return err
	}

	for _, contactAmount := range all {
		var buffer []byte
		for _, contact := range contactAmount.Contacts {
			// Contact time
			parser.Parse32Bits(contact.SavedTime, &bts)
			_, err := connection.Write(bts)
			if err != nil {
				return err
			}

			// Username
			err = parser.ParseString(contactAmount.Name, &buffer)
			if err != nil {
				return err
			}
			_, err = connection.Write(buffer)
			if err != nil {
				return err
			}

			// Contact Name
			contactName := contact.Name
			err = parser.ParseString(contactName, &buffer)
			if err != nil {
				return err
			}
			_, err = connection.Write(buffer)
			if err != nil {
				return err
			}

			// Number
			number := contact.Number
			err = parser.ParseLenString(number, &buffer, 20)
			if err != nil {
				return err
			}
			_, err = connection.Write(buffer)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

func recvClientUpdateCommand(name string, conn *net.Conn) error {
	_, buff, err := readAll(*conn, 256)
	if err != nil {
		return err
	}
	contactName := parser.ReadTillNull(buff)

	_, buff, err = readAll(*conn, 20)
	if err != nil {
		return err
	}
	number := parser.ReadTillNull(buff)

	newContact := data.AddContact(name, contactName, number)
	data.BroadcastUpdate(name, newContact)

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
	all := data.ListAllByName(name)
	amount := len(all)

	connection := *conn
	bts := make([]byte, 4)
	parser.Parse32Bits(uint32(amount), &bts)
	_, err := connection.Write(bts)
	if err != nil {
		return err
	}

	var buffer []byte
	for _, contact := range all {
		// Contact Name
		contactName := contact.Name
		err = parser.ParseString(contactName, &buffer)
		if err != nil {
			return err
		}
		_, err = connection.Write(buffer)
		if err != nil {
			return err
		}

		// Number
		number := contact.Number
		err = parser.ParseLenString(number, &buffer, 20)
		if err != nil {
			return err
		}
		_, err = connection.Write(buffer)
		if err != nil {
			return err
		}

	}

	return nil
}
