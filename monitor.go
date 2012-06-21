package textchat

import (
	"github.com/simonz05/godis/redis"
	"log"
	"strings"
)

// The Monitor type monitors the message updates and adds them to the database.
type Monitor struct {
	client      *redis.Client
	sub         *redis.Sub
	maxMsgCount int
}

var DefaultMonitor *Monitor

// Start up the DefaultMonitor and begin its subscription to all message update feeds.
func StartMonitor(maxMsgCount int) {
	client := redis.New(Cfg.DbAddr, Cfg.DbDb, Cfg.DbPassword)
	sub := redis.NewSub(Cfg.SubAddr, Cfg.SubDb, Cfg.SubPassword)

	err := sub.Psubscribe("textchat:*:Updates")
	if err != nil {
		log.Panicf("Error calling sub.Psubscribe: %s", err)
	}

	DefaultMonitor = &Monitor{client, sub, maxMsgCount - 1}

	go DefaultMonitor.updates()
}

func CloseMonitor() {
	DefaultMonitor.Close()
}

func (m *Monitor) Close() {
	m.sub.Close()
	m.client.Quit()
}

// Monitor all chat update channels for new messages. Upon receiving them, add them to the database (Ladd), then check to make sure the message list isn't too long. If it is, remove an element.
func (m *Monitor) updates() {
	for {
		redmsg, ok := <-m.sub.Messages
		if !ok {
			// Channel Closed
			return
		}
		area := Area{strings.Split(redmsg.Channel, ":")[1]}

		// Add it to the db
		n, err := m.client.Lpush(area.token("Messages"), redmsg.Elem.Bytes())
		if err != nil {
			log.Printf("Error on adding to area %s: %s", area.String(), err.Error())
			continue
		}
		if n > int64(m.maxMsgCount) {
			err = m.client.Ltrim(area.token("Messages"), 0, m.maxMsgCount)
			if err != nil {
				log.Printf("Error calling Ltrim for area %s: %s", area.String(), err.Error())
			}
		}
	}
}
