package data

import (
	"sync"
	"time"
)

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
}

// Map structured with `ServerName`: *ServerInfo{}
var ServerMap = new(sync.Map)

func UpdateServer(host string, port string, currentTime uint32) {
	fullname := host + ":" + port
	storedMap, exists := ContactsMap.Load(fullname)
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
		}
		ContactsMap.Store(fullname, serverInfo)
	}
}
