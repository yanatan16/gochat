// Package gochat is a simple chat server written in go. It is designed to be used with a Redis backend for multi-instance deployments. 
//
// Make sure to update the config.json file for non-default options.
//
package gochat

import (
	"fmt"
)

type Message struct {
	User *User
	Msg  string
}

type Area struct {
	Name string
}

type User struct {
	Name string
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
	if len(m.User.String()) > 0 {
		return fmt.Sprintf("%s: %s", m.User.String(), m.Msg)
	}
	return m.Msg
}

func (s *TextChatError) Error() string {
	return string(*s)
}
func NewMessage() *Message {
	msg := new(Message)
	msg.User = new(User)
	return msg
}
