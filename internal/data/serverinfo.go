package data

import (
	"distributed_contacts_server/internal/clock"
	"distributed_contacts_server/internal/parser"
	"fmt"
	"net"
	"sync"
	"time"
)

// TODO: Maybe use localhost and add server name?
const (
	SERVER_1 = "contact-server-1:9000"
	SERVER_2 = "contact-server-2:9001"
	SERVER_3 = "contact-server-3:9002"
)

type ServerInfo struct {
	lastHeartbeat time.Time
	Host          string
	Port          string
	currentTime   uint32
	status        bool
	conn          net.Conn
}

// Map structured with `ServerName`: *ServerInfo{}
var ServerMap = new(sync.Map)

func UpdateServer(host string, port string, name string, currentTime uint32) {
	storedMap, exists := ServerMap.Load(name)
	if exists {
		serverInfo := storedMap.(*ServerInfo)
		serverInfo.lastHeartbeat = time.Now()
		serverInfo.currentTime = currentTime
	} else {
		serverInfo := new(ServerInfo)
		*serverInfo = ServerInfo{
			lastHeartbeat: time.Now(),
			Host:          host,
			Port:          port,
			currentTime:   currentTime,
			status:        true,
		}
		ServerMap.Store(name, serverInfo)
	}
}

func AddServer(host string, port string, name string, conn net.Conn) {
	storedMap, exists := ServerMap.Load(name)
	if exists {
		serverInfo := storedMap.(*ServerInfo)
		serverInfo.lastHeartbeat = time.Now()
		serverInfo.status = true
		serverInfo.conn = conn
	} else {
		serverInfo := new(ServerInfo)
		*serverInfo = ServerInfo{
			lastHeartbeat: time.Now(),
			Host:          host,
			Port:          port,
			currentTime:   0,
			status:        true,
			conn:          conn,
		}
		ServerMap.Store(name, serverInfo)
	}
}

// TODO: Add client event when server Disconnect
func Disconnect(name string) {
	storedMap, exists := ServerMap.Load(name)
	if exists {
		serverInfo := storedMap.(*ServerInfo)
		serverInfo.status = false
	}
}

func Pong(name string, otherClock uint32, status bool) {
	storedMap, exists := ServerMap.Load(name)
	if exists {
		serverInfo := storedMap.(*ServerInfo)
		serverInfo.lastHeartbeat = time.Now()
		serverInfo.currentTime = otherClock

		if otherClock > clock.CurrentClock.Load() {
			clock.CurrentClock.Store(otherClock)
		}
	}
}

func BroadcastUpdate(userName string, contact *Contact) error {
	fmt.Printf("[SERVER] Broadicasting %s\n", userName)
	var err error = nil
	var buffer []byte
	ServerMap.Range(func(key, value any) bool {
		server := value.(*ServerInfo)
		if !server.status {
			return true
		}
		_, err = server.conn.Write([]byte{1})
		if err != nil {
			return false
		}

		// Clock
		bts := make([]byte, 4)
		parser.Parse32Bits(contact.SavedTime, &bts)
		_, err = server.conn.Write(bts)
		if err != nil {
			return false
		}

		// Username
		err = parser.ParseString(userName, &buffer)
		if err != nil {
			return false
		}
		_, err = server.conn.Write(buffer)
		if err != nil {
			return false
		}

		// Contact Name
		contactName := contact.Name
		err = parser.ParseString(contactName, &buffer)
		if err != nil {
			return false
		}
		_, err = server.conn.Write(buffer)
		if err != nil {
			return false
		}

		// Number
		number := contact.Number
		err = parser.ParseLenString(number, &buffer, 20)
		if err != nil {
			return false
		}
		_, err = server.conn.Write(buffer)
		if err != nil {
			return false
		}

		return true
	})

	return err
}

// TODO: Return null
func BroadcastDelete(username string, contact *Contact) error {
	return nil
}
