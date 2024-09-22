package data

import (
	"distributed_contacts_server/internal/clock"
	"fmt"
	"maps"
	"slices"
	"sync"
)

type Contact struct {
	Name      string
	Number    string
	savedTime uint32
}

// Map structured with `UserName`: { ContactName: Contact{} }
var ContactsMap = new(sync.Map)

// TODO: Make a broadcast when changing a contact

func AddClient(name string) {
	internalMap := make(map[string]*Contact)
	// Explicit not using pointers to map so we can re-store all user data on update
	_, l := ContactsMap.LoadOrStore(name, internalMap)

	if l {
		fmt.Println("Client re-registered: " + name)
	} else {
		fmt.Println("New client added: " + name)
	}
}

func AddContact(name string, contactName string, number string) {
	now := clock.CurrentClock.Add(1)
	m, _ := ContactsMap.Load(name)
	// TODO: Treat when there's no value for `name`

	contactMap := m.(map[string]*Contact)
	val, ok := contactMap[contactName]
	if ok {
		val.Number = number
	} else {
		newCon := new(Contact)
		*newCon = Contact{
			Name:      contactName,
			Number:    number,
			savedTime: uint32(now),
		}
		contactMap[contactName] = newCon
	}

	ContactsMap.Store(name, contactMap)
}

func RemoveContact(name string, contactName string) {
	now := clock.CurrentClock.Load()
	storedMap, _ := ContactsMap.Load(name)

	contactMap := storedMap.(map[string]*Contact)
	val, ok := contactMap[contactName]

	if ok {
		if val.savedTime < uint32(now) {
			delete(contactMap, contactName)
		}
	}

	ContactsMap.Store(name, contactMap)
}

func ListAll(name string) []*Contact {
	storedMap, _ := ContactsMap.Load(name)
	contactMap := storedMap.(map[string]*Contact)
	return slices.Collect(maps.Values(contactMap))
}
