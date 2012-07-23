package gochat

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"
)

type auGen int

var au auGen

const (
	timeout time.Duration = time.Second / 100
)

func init() {
	log.Println("Running test for Stream of groundfloor/textchat")
}

func (g auGen) gen() (*Area, *User) {
	g = auGen(int(g) + 1)
	return &Area{fmt.Sprintf("test-area-%d", int(g))},
		&User{fmt.Sprintf("test-user-%d", int(g))}
}

type streamObjectModelMsg struct {
	Msg string
}
type streamObjectModelUser struct {
	Name string
}
type streamObjectModel struct {
	Id, Op string
	Msgs   []streamObjectModelMsg
	Users  []streamObjectModelUser
}

func receiveCheckChat(t *testing.T, s Stream, exp *Message) {
	select {
	case msg, ok := <-s.Receive():
		if !ok {
			t.Fatal("Receive channel closed unexpectedly!")
		}
		obj := new(streamObjectModel)
		err := json.Unmarshal([]byte(msg), &obj)
		if err != nil {
			t.Error("Couldn't unmarshal message", err, msg)
		} else if obj.Id != "chat" {
			t.Error("Message should be a chat message!", obj)
		} else if len(obj.Msgs) != 1 {
			t.Error("Message should contain at least 1 message", obj)
		} else if obj.Msgs[0].Msg != exp.String() {
			t.Errorf("Message contents are not as expected! (exp:%s) (act:%s)", exp.Msg, obj.Msgs[0].Msg)
		}

	case <-time.After(timeout):
		t.Error("Receive channel contained no message when expected!", exp)
	}
}

func receiveCheckUser(t *testing.T, s Stream, exp *User, op string) {
	select {
	case msg, ok := <-s.Receive():
		if !ok {
			t.Fatal("Receive channel closed unexpectedly!")
		}
		obj := new(streamObjectModel)
		err := json.Unmarshal([]byte(msg), &obj)
		if err != nil {
			t.Error("Couldn't unmarshal message.", err, msg)
		} else if obj.Id != "user" {
			t.Error("Message should be a user message!", obj)
		} else if obj.Op != op {
			t.Errorf("User Message op should be %s, but is %s", op, obj.Op)
		} else if len(obj.Users) != 1 {
			t.Error("Message should contain at least 1 user", obj)
		} else if obj.Users[0].Name != exp.Name {
			t.Error("Message contents are not as expected!", obj)
		}

	case <-time.After(timeout):
		t.Error("Receive channel contained no message when expected!", exp)
	}
}

func errReceiveCheck(t *testing.T, s Stream) {
	select {
	case msg, ok := <-s.Receive():
		if ok {
			t.Error("Message received when expected channel to be closed!", msg)
		}
	case <-time.After(timeout):
		t.Error("Channel not closed when expected!")
	}
}

func noReceiveCheck(t *testing.T, s Stream) {
	select {
	case msg, ok := <-s.Receive():
		if ok {
			t.Error("Message received when no message expected!", msg)
		} else {
			t.Error("Channel closed unexpectedly!")
		}
	case <-time.After(timeout):

	}
}

func TestRegisterClose(t *testing.T) {
	a, u := au.gen()

	s, err := RegisterStream(a, u)
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	s.Close()
}

func TestReceive(t *testing.T) {
	a, u := au.gen()

	s, err := RegisterStream(a, u)
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	receiveCheckUser(t, s, u, "add")
	receiveCheckChat(t, s, a.joinMsg(u))

	noReceiveCheck(t, s)

	s.Close()

	errReceiveCheck(t, s)
}

func TestSend(t *testing.T) {
	a, u := au.gen()

	s, err := RegisterStream(a, u)
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	receiveCheckUser(t, s, u, "add")
	receiveCheckChat(t, s, a.joinMsg(u))

	msgstr := "my message!"
	err = s.Send(msgstr)
	if err != nil {
		t.Error("Error on Sending message:", Message{u, msgstr}, err)
	}

	receiveCheckChat(t, s, &Message{u, msgstr})

	s.Close()

	err = s.Send("no message!")
	if err == nil {
		t.Error("No error when sending on a closed Stream!")
	}
}

func TestStreamTwoUsers(t *testing.T) {
	a1, u1 := au.gen()
	_, u2 := au.gen()

	s1, err := RegisterStream(a1, u1)
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	s2, err := RegisterStream(a1, u2)
	if err != nil {
		t.Error("Couldn't create second Stream connection!", err)
	}
	defer s2.Close()

	receiveCheckUser(t, s1, u1, "add")
	receiveCheckChat(t, s1, a1.joinMsg(u1))

	receiveCheckUser(t, s1, u2, "add")
	receiveCheckChat(t, s1, a1.joinMsg(u2))

	receiveCheckUser(t, s2, u2, "add")
	receiveCheckChat(t, s2, a1.joinMsg(u2))

	msgstr := "cross the streams!"
	exp := &Message{u2, msgstr}
	err = s2.Send(msgstr)
	if err != nil {
		t.Error("Error on sending a message!", err)
	}

	receiveCheckChat(t, s1, exp)
	receiveCheckChat(t, s2, exp)

	s1.Close()

	errReceiveCheck(t, s1)
	receiveCheckUser(t, s2, u1, "rem")
	receiveCheckChat(t, s2, a1.leaveMsg(u1))
	noReceiveCheck(t, s2)
}
