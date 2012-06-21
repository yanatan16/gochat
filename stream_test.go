package textchat

import (
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

func receiveCheck(t *testing.T, s Stream, exp *Message) {
	select {
	case msg, ok := <-s.Receive():
		if !ok {
			t.Error("Receive channel closed unexpectedly!")
		} else if msg.String() != exp.String() {
			t.Errorf("Message received from area is incorrect! Expected: %s, Actual: %s", exp.String(), msg.String())
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

	s, err := RegisterStream(a, u, []Message{})
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	s.Close()
}

func TestReceive(t *testing.T) {
	a, u := au.gen()

	s, err := RegisterStream(a, u, []Message{})
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	receiveCheck(t, s, a.joinMsg(u))

	noReceiveCheck(t, s)

	s.Close()

	errReceiveCheck(t, s)
}

func TestSend(t *testing.T) {
	a, u := au.gen()

	s, err := RegisterStream(a, u, []Message{})
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	receiveCheck(t, s, a.joinMsg(u))

	msgstr := "my message!"
	err = s.Send(msgstr)
	if err != nil {
		t.Error("Error on Sending message:", Message{u, msgstr}, err)
	}

	receiveCheck(t, s, &Message{u, msgstr})

	s.Close()

	err = s.Send("no message!")
	if err == nil {
		t.Error("No error when sending on a closed Stream!")
	}
}

func TestBacklog(t *testing.T) {
	a, u := au.gen()
	backlog := []Message{}
	backlog = append(backlog, Message{&User{"testing"}, "testing message"})
	backlog = append(backlog, Message{&User{"iamreal"}, "truth is fiction"})

	s, err := RegisterStream(a, u, backlog)
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}
	defer s.Close()

	receiveCheck(t, s, &backlog[0])
	receiveCheck(t, s, &backlog[1])
	receiveCheck(t, s, a.joinMsg(u))
}

func TestStreamTwoUsers(t *testing.T) {
	a1, u1 := au.gen()
	_, u2 := au.gen()

	s1, err := RegisterStream(a1, u1, []Message{})
	if err != nil {
		t.Fatal("Failed to register a test stream. Is redis there and configured for use?", err)
	}

	s2, err := RegisterStream(a1, u2, []Message{})
	if err != nil {
		t.Error("Couldn't create second Stream connection!", err)
	}
	defer s2.Close()

	receiveCheck(t, s1, a1.joinMsg(u1))
	//	receiveCheck(t, s2, a1.joinMsg(u1)) 
	receiveCheck(t, s1, a1.joinMsg(u2))
	receiveCheck(t, s2, a1.joinMsg(u2))

	msgstr := "cross the streams!"
	exp := &Message{u2, msgstr}
	err = s2.Send(msgstr)
	if err != nil {
		t.Error("Error on sending a message!", err)
	}

	receiveCheck(t, s1, exp)
	receiveCheck(t, s2, exp)

	s1.Close()

	errReceiveCheck(t, s1)
	receiveCheck(t, s2, a1.leaveMsg(u1))
	noReceiveCheck(t, s2)

}
