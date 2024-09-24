package main

import (
	"distributed_contacts_server/internal/data"
	"distributed_contacts_server/pkg/tcpserver"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func remove(slice []string, item string) ([]string, bool) {
	index := -1
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			index = i
			break
		}
	}

	if index == -1 {
		return nil, false
	}

	return append(slice[:index], slice[index+1:]...), false
}

var otherServers = []string{data.SERVER_1, data.SERVER_2, data.SERVER_3}

func usageAndExit() {
	fmt.Println("Usage: server server={1,2,3}")
	os.Exit(1)
}

// Retrieve host and port from string
// Example: localhost:8989 -> (localhost, 8989)
// As colateral, removes from the `otherServers` variable the current server
func getHostAndPort(serverName string) (string, string) {
	otherServers, _ = remove(otherServers, serverName)

	hostAndPort := strings.Split(serverName, ":")
	host := hostAndPort[0]
	port := hostAndPort[1]

	return host, port
}

func main() {
	if len(os.Args) < 2 {
		usageAndExit()
	}

	serverOption := os.Args[1]
	server, err := strconv.Atoi(strings.Split(serverOption, "=")[1])
	if err != nil {
		usageAndExit()
	}

	var host string
	var port string
	if server == 1 {
		host, port = getHostAndPort(data.SERVER_1)
	} else if server == 2 {
		host, port = getHostAndPort(data.SERVER_2)
	} else if server == 3 {
		host, port = getHostAndPort(data.SERVER_3)
	}

	for _, serv := range otherServers {
		hostAndPort := strings.Split(serv, ":")
		otherHost := hostAndPort[0]
		otherPort := hostAndPort[1]
		go tcpserver.Connect(otherHost, otherPort, host)
	}

	tcpserver.Listen(host, port)
}
