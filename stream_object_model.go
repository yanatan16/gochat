package gochat

import (
	"encoding/json"
	"log"
)

type streamChatMessageMsg struct {
	Msg string
}
type streamChatMessage struct {
	Id   string
	Msgs []streamChatMessageMsg
}

func newStreamChatMessage(msg *Message) *streamChatMessage {
	return &streamChatMessage{
		"chat",
		[]streamChatMessageMsg{{msg.String()}},
	}
}

type streamUserMessageUser struct {
	Name string
}
type streamUserMessage struct {
	Id    string
	Op    string
	Users []streamUserMessageUser
}

func newStreamUserMessage(u *User, add bool) *streamUserMessage {
	um := &streamUserMessage{
		"user",
		"add",
		[]streamUserMessageUser{{u.String()}},
	}
	if !add {
		um.Op = "rem"
	}
	return um
}

func (m *streamChatMessage) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("Error marshalling stream chat message!", err)
		return ""
	}
	return string(b)
}
func (m *streamUserMessage) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("Error marshalling stream user message!", err)
		return ""
	}
	return string(b)
}

type streamClientMessage struct {
	Msg string
}

func interpretClientMessage(msg string) (*streamClientMessage, error) {
	m := &streamClientMessage{}
	err := json.Unmarshal([]byte(msg), m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
