package gochat

import (
	"encoding/json"
	"fmt"
)

func throw(s ...interface{}) error {
	t := TextChatError(fmt.Sprint(s...))
	return &t
}

func retry(f func() error, n int) error {
	var err error
	for err = f(); err != nil && n > 0; n -= 1 {
		err = f()
	}
	return err
}

// Deserialize an object
func Deserialize(b []byte, data interface{}) error {
	return json.Unmarshal(b, data)
}

// Serialize an object
func Serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func safeClose(c chan Message) {
	for {
		select {
		case <-c:

		default:
			close(c)
			return
		}
	}
}
