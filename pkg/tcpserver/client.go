package tcpserver

import (
	"distributed_contacts_server/internal/clock"
	"distributed_contacts_server/internal/data"
	"distributed_contacts_server/internal/parser"
	"fmt"
	"net"
	"time"
)

func Connect(host string, port string, servername string) {
	otherName := host + ":" + port

	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Printf("[CONNECT][%s] Could not connect to `%s:%s`: %s \n", servername, host, port, err)
		return
	}
	defer conn.Close()

	realServerName, err := initialConnectionSetup(servername, &conn, otherName)
	fmt.Printf("[CONNECT][%s] Handshake done with %s! Connected with name %s\n", servername, otherName, realServerName)

	data.AddServer(host, port, realServerName)

	// TODO: Send an ListAll request to server
	go initialPingServerLoop(servername, conn, otherName)
	serverLoop(realServerName, conn)
}

func initialConnectionSetup(serverName string, conn *net.Conn, otherName string) (string, error) {
	fmt.Printf("[CONNECT][%s] Handshake start with %s\n", serverName, otherName)
	_, err := (*conn).Write([]byte{1})
	if err != nil {
		return "", err
	}

	var buffer []byte
	// TODO: Validate error
	_ = parser.ParseString(serverName, &buffer)

	fmt.Printf("[CONNECT][%s] Sending my name to %s\n", serverName, otherName)
	_, err = (*conn).Write(buffer)
	if err != nil {
		return "", err
	}

	fmt.Printf("[CONNECT][%s] Reading other name from %s\n", serverName, otherName)
	read, buf, err := readAll(*conn, 256)
	if err != nil || read != 256 {
		fmt.Printf("[CONNECT][%s] Could not get name from %s\n", serverName, otherName)
		return "", nil
	}

	name := parser.ReadTillNull(buf)
	return name, nil
}

// Function to ping server sending heartbeat
func initialPingServerLoop(name string, conn net.Conn, otherName string) {
	retries := 0
	for {
		fmt.Printf("[CONNECT][%s] Pinging %s\n", name, otherName)
		time.Sleep(time.Second * 5)
		n, err := conn.Write([]byte{3})

		if n < 1 || err != nil {
			retries++
			fmt.Printf("[CONNECT][%s] Could not ping server %s\n", name, otherName)

			if retries > 3 {
				return
			}

			continue
		}

		n, err = conn.Write([]byte{1})
		if n < 1 || err != nil {
			retries++
			fmt.Printf("[CONNECT][%s] Could not ping server %s\n", name, otherName)

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
			fmt.Printf("[CONNECT][%s] Could not ping server %s\n", name, otherName)

			if retries > 3 {
				return
			}

		}
	}
}
