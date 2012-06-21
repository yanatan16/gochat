// Package textchat forms the backend of a text chatting system using a memcached data storage layer and channels for asynchronous messaging.
package textchat

import (
	"fmt"
	"strings"
)

type Message struct {
	Usr *User
	Msg string
}

type Area struct {
	Name string
}

type User struct {
	Name string
}

type Serializer interface {
	Store() string
	Parse(string)
}

type TextChatError string

func (a *Area) token(key string) string {
	return fmt.Sprintf("textchat:%s:%s", a.Name, key)
}
func (a *Area) unique(user *User) string {
	return a.String() + "/" + user.String()
	//TODO use a hash
}

func (a *Area) welcome() *Message {
	return &Message{&User{""}, "Welcome to " + a.Name + "!"}
}
func (a *Area) joinMsg(user *User) *Message {
	return &Message{&User{""}, fmt.Sprintf("User %s has joined %s!", user.String(), a.String())}
}
func (a *Area) leaveMsg(user *User) *Message {
	return &Message{&User{""}, fmt.Sprintf("User %s has left %s!", user.String(), a.String())}
}

func (a *Area) String() string {
	return a.Name
}

func (u *User) String() string {
	return u.Name
}

func (m *Message) String() string {
	if len(m.Usr.String()) > 0 {
		return fmt.Sprintf("%s: %s", m.Usr.String(), m.Msg)
	}
	return m.Msg
}

func (s *TextChatError) Error() string {
	return string(*s)
}

func (u *User) Store() string {
	return u.Name
}
func (u *User) Parse(s string) {
	u.Name = s
}

func (a *Area) Store() string {
	return a.Name
}
func (a *Area) Parse(s string) {
	a.Name = s
}

func (m *Message) Store() string {
	return fmt.Sprintf("%s|%s", m.Usr.Store(), m.Msg)
}
func (m *Message) Parse(s string) {
	ss := strings.SplitN(s, "|", 2)
	if len(ss) == 1 {
		m.Usr.Parse("")
		m.Msg = strings.TrimLeft(ss[0], "|")
	} else {
		m.Usr.Parse(ss[0])
		m.Msg = strings.TrimLeft(ss[1], "|")
	}
}
func NewMessage() *Message {
	msg := new(Message)
	msg.Usr = new(User)
	return msg
}
