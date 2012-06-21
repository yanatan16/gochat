package gochat

import (
	"fmt"
	"github.com/simonz05/godis/redis"
	"log"
)

type Stream interface {

	// Get the receive channel of messages
	Receive() <-chan Message

	// Send a string message
	Send(msg string) error

	// Close the stream.
	Close()

	// Check whether the stream is closed.
	IsClosed() bool

	// Get a printable name of the area.
	Name() string
}

type streamImpl struct {
	pub    *redis.Client
	sub    *redis.Sub
	rcv    chan Message
	area   *Area
	user   *User
	signal chan bool
}

func RegisterStream(area *Area, user *User) (s Stream, err error) {

	pub := redis.New(Cfg.SubAddr, Cfg.SubDb, Cfg.SubPassword)
	sub := redis.NewSub(Cfg.SubAddr, Cfg.SubDb, Cfg.SubPassword)

	err = sub.Subscribe(area.token("Updates"))

	if err != nil {
		return nil, err
	}

	si := &streamImpl{
		pub,
		sub,
		make(chan Message, 0),
		area,
		user,
		make(chan bool, 0),
	}

	go si.rcvGoroutine()

	// Send first message
	msg := area.joinMsg(user)
	_, err = si.pub.Publish(si.area.token("Updates"), Serialize(msg))
	if err != nil {
		return nil, err
	}

	return si, nil
}

func (s *streamImpl) IsClosed() bool {
	select {
	case <-s.signal:
		// Closed
		return true
	default:
		// Open!
		return false
	}
	return false
}

func (s *streamImpl) Receive() <-chan Message {
	return s.rcv
}

func (s *streamImpl) Send(m string) error {
	if s.IsClosed() {
		return throw("Stream is closed!")
	}

	msg := Message{s.user, m}
	_, err := s.pub.Publish(s.area.token("Updates"), Serialize(&msg))
	if err != nil {
		return err
	}
	return nil
}

func (s *streamImpl) Close() {
	if s.IsClosed() {
		log.Println("Repeat close on closed stream.")
		return
	}
	// Signal the goroutines
	close(s.signal)

	s.sub.Unsubscribe(s.area.token("Updates"))

	// Send a leaving message
	s.pub.Publish(s.area.token("Updates"), Serialize(s.area.leaveMsg(s.user)))

	// Finish up
	s.sub.Close()
	s.pub.Quit()
}

func (s *streamImpl) rcvGoroutine() {
	// Send any new messages
loop:
	for {
		select {
		case redisMsg, ok := <-s.sub.Messages:
			if !ok {
				// Connection lost!
				break loop
			}
			msg := NewMessage()
			Deserialize(redisMsg.Elem.Bytes(), msg)
			s.rcv <- *msg
		case <-s.signal:
			// Stream Closed!
			break loop
		}
	}

	close(s.rcv)
}

func (s *streamImpl) Name() string {
	return fmt.Sprintf("Area: %s, User: %s", s.area.String(), s.user.String())
}
