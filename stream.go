package gochat

import (
	"fmt"
	"github.com/simonz05/godis"
	"log"
)

type Stream interface {

	// Get the receive channel of messages to send to the client
	Receive() <-chan string

	// Send a string chat message
	Send(msg string) error

	// Close the stream.
	Close()

	// Check whether the stream is closed.
	IsClosed() bool

	// Get a printable name of the area.
	Name() string
}

type streamImpl struct {
	pub    *godis.Client
	sub    *godis.Sub
	rcv    chan string
	area   *Area
	user   *User
	signal chan bool
}

func RegisterStream(area *Area, user *User) (Stream, error) {

	pub := godis.New(Cfg.SubAddr, Cfg.SubDb, Cfg.SubPassword)
	sub := godis.NewSub(Cfg.SubAddr, Cfg.SubDb, Cfg.SubPassword)

	err := sub.Subscribe(area.token("Updates"))

	if err != nil {
		return nil, err
	}

	s := &streamImpl{
		pub,
		sub,
		make(chan string, 0),
		area,
		user,
		make(chan bool, 0),
	}

	go s.rcvGoroutine()

	// Do on-join activities
	s.onJoin()

	return s, nil
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

func (s *streamImpl) Receive() <-chan string {
	return s.rcv
}

func (s *streamImpl) Send(m string) error {
	if s.IsClosed() {
		return throw("Stream is closed!")
	}

	msg := newStreamChatMessage(&Message{s.user, m})
	_, err := s.pub.Publish(s.area.token("Updates"), msg.String())
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

	// Notify others of leaving
	s.onLeave()

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
			s.rcv <- string(redisMsg.Elem.Bytes())
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

func (s *streamImpl) onJoin() error {
	// Notify user-join
	join := newStreamUserMessage(s.user, true)
	_, err := s.pub.Publish(s.area.token("Updates"), join.String())
	if err != nil {
		return err
	}

	// Send join chat message
	msg := newStreamChatMessage(s.area.joinMsg(s.user))
	_, err = s.pub.Publish(s.area.token("Updates"), msg.String())
	if err != nil {
		return err
	}

	return nil
}

func (s *streamImpl) onLeave() error {
	// Notify user-join
	join := newStreamUserMessage(s.user, false) // false means leave
	_, err := s.pub.Publish(s.area.token("Updates"), join.String())
	if err != nil {
		return err
	}

	// Send join chat message
	msg := newStreamChatMessage(s.area.leaveMsg(s.user))
	_, err = s.pub.Publish(s.area.token("Updates"), msg.String())
	if err != nil {
		return err
	}

	return nil
}
