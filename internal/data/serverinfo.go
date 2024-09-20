package data

import "time"

type ServerInfo struct {
	lastHeartbeat time.Time
	Host          string
	Port          string
}
