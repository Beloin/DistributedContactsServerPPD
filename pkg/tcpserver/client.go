package tcpserver

import (
	"fmt"
	"net"
)

func Connect() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	// TODO: Make a callback when recv data
}
