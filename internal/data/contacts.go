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
	SavedTime uint32
}

// Map structured with `UserName`: { ContactName: *Contact{} }
var ContactsMap = new(sync.Map)

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

// TODO: Make a broadcast when changing a contact
// But create other function, so in `CompareAndUpdateContact` we don't end
// in an infinite loop
// The same for delete etc
// Adds or update contact for given UserName
func AddContact(name string, contactName string, number string) *Contact {
	now := clock.CurrentClock.Add(1)
	m, _ := ContactsMap.Load(name)
	// TODO: Treat when there's no value for `name`

	contactMap := m.(map[string]*Contact)
	var newCon *Contact
	val, ok := contactMap[contactName]
	if ok {
		val.Number = number
	} else {
		newCon = new(Contact)
		*newCon = Contact{
			Name:      contactName,
			Number:    number,
			SavedTime: uint32(now),
		}
		contactMap[contactName] = newCon
	}

	ContactsMap.Store(name, contactMap)

	return newCon
}

func RemoveContact(name string, contactName string) {
	now := clock.CurrentClock.Load()
	storedMap, _ := ContactsMap.Load(name)

	contactMap := storedMap.(map[string]*Contact)
	val, ok := contactMap[contactName]

	if ok {
		if val.SavedTime < uint32(now) {
			delete(contactMap, contactName)
		}
	}

	ContactsMap.Store(name, contactMap)
}

type ContactAmount struct {
	Name     string
	Contacts []*Contact
}

func ListAll() ([]ContactAmount, int) {
	fullAmount := 0
	var contactAmountSlice []ContactAmount
	ContactsMap.Range(func(key, value any) bool {
		name := key.(string)
		cont := ListAllByName(name)
		fullAmount += len(cont)
		contactAmountSlice = append(contactAmountSlice, ContactAmount{Name: name, Contacts: cont})
		return true
	})

	return contactAmountSlice, fullAmount
}

func ListAllByName(name string) []*Contact {
	storedMap, _ := ContactsMap.Load(name)
	contactMap := storedMap.(map[string]*Contact)
	return slices.Collect(maps.Values(contactMap))
}

func CompareAndUpdateContact(name string, contactName string, number string, otherTime uint32) {
	if otherTime > clock.CurrentClock.Load() {
		clock.CurrentClock.Store(otherTime)
	}

	userMap, exists := ContactsMap.Load(name)

	if !exists {
		AddClient(name)
		AddContact(name, contactName, number)
		return
	}

	contactMap := userMap.(map[string]*Contact)
	val, exists := contactMap[contactName]

	if exists {
		if val.SavedTime >= otherTime {
			return
		}
	}

	AddContact(name, contactName, number)
}

func CompareAndDeleteContact(name string, contactName string, otherTime uint32) {
	if otherTime > clock.CurrentClock.Load() {
		clock.CurrentClock.Store(otherTime)
	}
	userMap, exists := ContactsMap.Load(name)

	if !exists {
		AddClient(name)
		return
	}

	innerContacts := userMap.(map[string]*Contact)
	_, exists = innerContacts[contactName]
	if exists {
		delete(innerContacts, contactName)
		ContactsMap.Store(name, innerContacts)
	}
}
