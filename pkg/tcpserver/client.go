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
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Println("[CONNECT] Could not connect to ", host, ": ", err)
		return
	}
	defer conn.Close()
	initialConnectionSetup(servername, &conn)
	pingServerLoop_(servername, conn)
}

func initialConnectionSetup(serverName string, conn *net.Conn) error {
	_, err := (*conn).Write([]byte{1})
	if err != nil {
		return err
	}


  return nil
}

// Function to ping server sending heartbeat
func pingServerLoop_(name string, conn net.Conn) {
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
