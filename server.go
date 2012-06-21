package textchat

import (
	"github.com/simonz05/godis/redis"
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
	db *redis.Client
}

var areasToken string = "textchat:Areas"

func NewServer() Server {
	var t tcRedis
	t.db = redis.New(Cfg.DbAddr, Cfg.DbDb, Cfg.DbPassword)
	return &t
}

func (t *tcRedis) Quit() {
	t.db.Quit()
}

func (t *tcRedis) CreateArea(area *Area) (err error) {
	unique, err := t.db.Sadd(areasToken, Serialize(area))
	if err != nil {
		return throw("Area could not be created!", area.String(), err.Error())
	}

	if !unique {
		log.Println("Non-unique area created.", area.String())
		return nil
	}

	msg := area.welcome()
	_, err = t.db.Publish(area.token("Updates"), Serialize(msg))
	if err != nil {
		return err
	}

	return nil
}

func (t *tcRedis) DeleteArea(area *Area) {
	t.db.Del(area.token("Users"), area.token("Messages"), area.token("Updates"))

	t.db.Srem(areasToken, Serialize(area))
}

func (t *tcRedis) JoinArea(area *Area, user *User) error {
	unique, err := t.db.Sadd(area.token("Users"), Serialize(user))
	if err != nil {
		return err
	}
	if !unique {
		return throw("JoinArea: User %s is not unique to area %s.", user.String(), area.String())
	}
	return nil
}

func (t *tcRedis) LeaveArea(area *Area, user *User) error {
	_, err := t.db.Srem(area.token("Users"), Serialize(user))
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
	for i := range areasBytes {
		Deserialize(areasBytes[i], &areas[i])
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
	for i := range usersBytes {
		Deserialize(usersBytes[i], &users[i])
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
	for i := range msgsBytes {
		msgs[i].Usr = new(User)
		Deserialize(msgsBytes[i], &msgs[i])
	}
	return msgs, nil
}
