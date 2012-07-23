package gochat

import (
	"github.com/simonz05/godis"
	"log"
)

// Server is the main interface to perform the backend functions of text chatting.
type Server interface {
	// Close the interface
	Quit()

	// CreateArea creates a Chat Area.
	CreateArea(area *Area) (err error)

	// DeleteArea deletes a Chat Area.
	DeleteArea(area *Area)

	// JoinArea lets a user join an area.
	JoinArea(area *Area, user *User) (err error)

	// LeaveArea lets a user leave an area
	LeaveArea(area *Area, user *User) (err error)

	// ListAreas will list all available areas.
	ListAreas() (areas []Area, err error)

	// ListUsers will list the users in the area given.
	ListUsers(area *Area) (users []User, err error)

	// ListMessages will list the current messages in an area.
	ListMessages(area *Area) (msgs []Message, err error)
}

type tcRedis struct {
	db *godis.Client
}

var areasToken string = "textchat:Areas"

func NewServer() Server {
	var t tcRedis
	t.db = godis.New(Cfg.DbAddr, Cfg.DbDb, Cfg.DbPassword)
	return &t
}

func (t *tcRedis) Quit() {
	t.db.Quit()
}

func (t *tcRedis) CreateArea(area *Area) (err error) {
	s, err := Serialize(area)
	if err != nil {
		return err
	}

	unique, err := t.db.Sadd(areasToken, s)
	if err != nil {
		return throw("Area could not be created!", area.String(), err.Error())
	}

	if !unique {
		log.Println("Non-unique area created.", area.String())
		return nil
	}

	msg := area.welcome()
	s, err = Serialize(msg)
	if err != nil {
		return err
	}

	_, err = t.db.Publish(area.token("Updates"), s)
	if err != nil {
		return err
	}

	return nil
}

func (t *tcRedis) DeleteArea(area *Area) {
	t.db.Del(area.token("Users"), area.token("Messages"), area.token("Updates"))

	s, err := Serialize(area)
	if err != nil {
		log.Println("Error serializing area for removal!", err)
	}
	t.db.Srem(areasToken, s)
}

func (t *tcRedis) JoinArea(area *Area, user *User) error {
	s, err := Serialize(user)
	if err != nil {
		return err
	}

	unique, err := t.db.Sadd(area.token("Users"), s)
	if err != nil {
		return err
	}
	if !unique {
		return throw("JoinArea: User %s is not unique to area %s.", user.String(), area.String())
	}
	return nil
}

func (t *tcRedis) LeaveArea(area *Area, user *User) error {
	s, err := Serialize(user)
	if err != nil {
		return err
	}

	_, err = t.db.Srem(area.token("Users"), s)
	if err != nil {
		return throw("LeaveArea: User %s is not part of Area %s.", user.String(), area.String())
	}
	return nil
}

func (t *tcRedis) ListAreas() (areas []Area, err error) {
	reply, err := t.db.Smembers(areasToken)
	if err != nil {
		return nil, err
	}
	areasBytes := reply.BytesArray()
	areas = make([]Area, len(areasBytes))
	for i, a := range areasBytes {
		area := &areas[i]
		Deserialize(a, area)
	}
	return areas, nil
}

func (t *tcRedis) ListUsers(area *Area) (users []User, err error) {
	reply, err := t.db.Smembers(area.token("Users"))
	if err != nil {
		return nil, err
	}
	usersBytes := reply.BytesArray()
	users = make([]User, len(usersBytes))
	for i, u := range usersBytes {
		user := &users[i]
		Deserialize(u, user)
	}
	return users, nil
}

func (t *tcRedis) ListMessages(area *Area) (msgs []Message, err error) {
	rep, err := t.db.Lrange(area.token("Messages"), 0, -1)
	if err != nil {
		return nil, err
	}

	msgsBytes := rep.BytesArray()
	msgs = make([]Message, len(msgsBytes))
	for i, m := range msgsBytes {
		msg := &msgs[i]
		Deserialize(m, msg)
		msgs[i].User = new(User)
	}
	return msgs, nil
}
