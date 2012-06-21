package gochat

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"log"
	"net/http"
)

type StreamService interface {
	InitiateStream(area *Area, user *User) (string, error)

	CloseStream(area *Area, user *User)
}

type streamServiceImpl struct {
	addr    string
	port    int
	streams map[string]Stream
	handler SimpleHandler
}

func NewStreamService() StreamService {
	ss := &streamServiceImpl{
		Cfg.WsAddr,
		Cfg.WsPort,
		make(map[string]Stream),
		NewSimpleHandler(),
	}
	http.Handle("/", ss.handler)
	go http.ListenAndServe(fmt.Sprintf(":%d", Cfg.WsPort), nil)
	return ss
}

func (s *streamServiceImpl) InitiateStream(area *Area, user *User) (sAddr string, err error) {
	stream, err := RegisterStream(area, user)
	if err != nil {
		return "", err
	}
	uuid := area.unique(user)

	s.streams[uuid] = stream

	defer func() {
		if err := recover(); err != nil {
			log.Println("Could not register websocket", err)
		}
	}()

	// Start Server
	s.handler.Add(fmt.Sprintf("/%s", uuid), websocket.Handler(MakeHandler(stream)))

	sAddr = fmt.Sprintf("ws://%s:%d/%s", s.addr, s.port, uuid)

	return sAddr, nil
}

func (s *streamServiceImpl) CloseStream(area *Area, user *User) {
	uuid := area.unique(user)
	if stream, ok := s.streams[uuid]; !ok {
		log.Printf("CloseStream: User %s not registerd to area %s", user.String(), area.String())
	} else {
		s.handler.Rem(fmt.Sprintf("/%s", uuid))
		stream.Close()
	}
}

func MakeHandler(s Stream) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Println("Websocket handler called for stream", s.Name(), "at", ws.RemoteAddr())
		go StreamServer(s, ws)
		StreamClient(s, ws)
	}
}

// StreamServer listens on a websocket and forwards received messages to a text chat stream.
func StreamServer(s Stream, ws *websocket.Conn) {
	for {
		var msg string
		// Receive receives a text message from client, since buf is string.
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Printf("Error receiving on websocket for %s: %s", s.Name(), err.Error())
			break
		}

		err = s.Send(msg)
		if err != nil {
			log.Printf("Error sending message for %s: %s", s.Name(), err.Error())
			break
		}
	}
	ws.Close()
}

// StreamClient listens on received messages and forwards it down a websocket connection. It should be called in a goroutine.
func StreamClient(s Stream, ws *websocket.Conn) {
	rec := s.Receive()
	for {
		msg, ok := <-rec
		if !ok {
			break
		}

		err := websocket.Message.Send(ws, msg.String())
		if err != nil {
			log.Printf("Error on sending message for stream %s!", s.Name())
			break
		}
	}
	ws.Close()
}
