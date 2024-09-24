package tcpserver

import (
	"distributed_contacts_server/internal/clock"
	"distributed_contacts_server/internal/parser"
	"fmt"
	"net"
	"time"
)

// TODO: Create a connection to other servers
func Connect(host string, port string, servername string) {
	otherName := host + ":" + port

	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Printf("[CONNECT][%s] Could not connect to `%s:%s`: %s \n", servername, host, port, err)
		return
	}
	defer conn.Close()

	initialConnectionSetup(servername, &conn, otherName)
	fmt.Printf("[CONNECT][%s] Handshake done with %s! Connected\n", servername, otherName)

	initialPingServerLoop(servername, conn, otherName)
}

// TODO: Send an ListAll request to server
func initialConnectionSetup(serverName string, conn *net.Conn, otherName string) error {
	fmt.Printf("[CONNECT][%s] Handshake start with %s\n", serverName, otherName)
	_, err := (*conn).Write([]byte{1})
	if err != nil {
		return err
	}

	var buffer []byte
	// TODO: Validate error
	_ = parser.ParseString(serverName, &buffer)

	fmt.Printf("[CONNECT][%s] Sending my name to %s\n", serverName, otherName)
	_, err = (*conn).Write(buffer)
	if err != nil {
		return err
	}

	return nil
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
