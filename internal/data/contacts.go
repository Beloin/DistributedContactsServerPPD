package data

import "sync"

type Contact struct {
	Name      string
	Number    string
	savedTime uint32
}

// Map structured with `UserName`: { ContactName: Contact{} }
var ContactsMap = new(sync.Map)

// TODO: Make a broadcast when changing a contact
