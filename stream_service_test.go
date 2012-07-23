package gochat

import (
	"code.google.com/p/go.net/websocket"

	"log"
	"net/http"
	"testing"
	"time"
)

var (
	srv *http.Server
	mux *http.ServeMux
	ss  StreamService
)

func init() {
	log.Printf("Running Test for Stream Service of groundfloor/textchat")
	Cfg.WsAddr = "127.0.0.1"
	Cfg.WsPort = 21234
	ss = NewStreamService()
}

func TestStreamService(t *testing.T) {
	area, user := au.gen()
	//area, user := Area{"streamServiceArea"}, User{"superuser"}

	msgs := []Message{*area.joinMsg(user), Message{user, "testmsg"}}
	mch := make(chan string, 2)

	cAddr, err := ss.InitiateStream(area, user)
	if err != nil {
		t.Error("Error on InitiateStream!", err)
	}
	defer ss.CloseStream(area, user)

	client, err := websocket.Dial(cAddr, "", "http://localhost:21235/")
	if err != nil {
		t.Error("Error on opening return websocket!", err)
	}
	defer client.Close()

	go func() {
		for {
			var msg string
			err := websocket.Message.Receive(client, &msg)
			if err != nil {
				log.Println("Error receiving messages", err)
				break
			}
			mch <- msg
		}
		close(mch)
	}()

	// Send a message
	err = websocket.Message.Send(client, "{\"msg\":\"testmsg\"}")
	if err != nil {
		t.Error("Error on sending message through client websocket!")
	}

	select {
	case usermsg := <-mch:
		if usermsg != newStreamUserMessage(user, true).String() {
			t.Fatalf("User message isn't what was expected! (exp:%s) (msg:%s)",
				user.String(), usermsg)
		}
	case <-time.After(time.Second / 2):
		t.Fatal("No messages received after timeout!")
	}

	for j := range msgs {
		select {
		case msg := <-mch:
			if msg != newStreamChatMessage(&msgs[j]).String() {
				t.Errorf("Message isn't what was expected! Exp: %s, Actual: %s", msgs[j].String(), msg)
			}
		case <-time.After(time.Second):
			t.Error("No messages received after timeout!", j)
		}
	}

}
