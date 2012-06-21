package textchat

import (
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

func Deserialize(b []byte, data Serializer) {
	if data == nil {
		panic("Can't deserialize to nil Serializer!")
	}
	data.Parse(string(b))
}
func Serialize(data Serializer) []byte {
	return []byte(data.Store())
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
