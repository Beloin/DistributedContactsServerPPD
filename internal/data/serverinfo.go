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

// Map structured with `ServerName`: ServerInfo{}
var ServerMap = new(sync.Map)

func UpdateServer(host string, port string, currentTime int32) {
}
