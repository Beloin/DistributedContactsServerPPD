package data

import (
	"distributed_contacts_server/internal/clock"
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
}

// Map structured with `ServerName`: *ServerInfo{}
var ServerMap = new(sync.Map)

func UpdateServer(host string, port string, name string, currentTime uint32) {
	storedMap, exists := ContactsMap.Load(name)
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

func AddServer(host string, port string, name string) {
	storedMap, exists := ContactsMap.Load(name)
	if exists {
		serverInfo := storedMap.(*ServerInfo)
		serverInfo.lastHeartbeat = time.Now()
	} else {
		serverInfo := new(ServerInfo)
		*serverInfo = ServerInfo{
			lastHeartbeat: time.Now(),
			Host:          host,
			Port:          port,
			currentTime:   0,
			status:        true,
		}
    ServerMap.Store(name, serverInfo)
	}
}


// TODO: Add client event when server Disconnect
func Disconnect(name string) {
	storedMap, exists := ContactsMap.Load(name)
	if exists {
		serverInfo := storedMap.(*ServerInfo)
		serverInfo.status = false
	}
}

func Pong(name string, otherClock uint32, status bool) {
	storedMap, exists := ContactsMap.Load(name)
	if exists {
		serverInfo := storedMap.(*ServerInfo)
		serverInfo.lastHeartbeat = time.Now()
		serverInfo.currentTime = otherClock

		if otherClock > clock.CurrentClock.Load() {
			clock.CurrentClock.Store(otherClock)
		}
	}
}

func Broadcast(userName string, contact *Contact) {
}
