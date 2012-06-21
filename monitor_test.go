package textchat

import (
	"github.com/simonz05/godis/redis"
	"log"
	"testing"
	"time"
)

func init() {
	log.Println("Running Test for Monitor of groundfloor/textchat")
}

func pub(t *testing.T, client *redis.Client, key string, msg *Message) {
	_, err := client.Publish(key, Serialize(msg))
	if err != nil {
		t.Errorf("Error plublishing on key %s, msg %s: %s", key, msg.String(), err)
	}
}

func TestMonitor(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal("panic!", err)
		}
	}()

	StartMonitor(3)
	defer CloseMonitor()

	dbClient := redis.New(Cfg.DbAddr, Cfg.DbDb, Cfg.DbPassword)
	defer dbClient.Quit()
	subClient := redis.New(Cfg.SubAddr, Cfg.SubDb, Cfg.SubPassword)
	defer subClient.Quit()

	area1 := &Area{"monitorArea"}
	area2 := &Area{"unmonitored?"}
	area3 := &Area{"unused"}
	user1 := &User{"bigbrother"}
	user2 := &User{"lilsister"}
	msgs1 := []Message{
		*area1.joinMsg(user1),
		*area1.joinMsg(user2),
		Message{user1, "hello?"},
		Message{user2, "hello!"},
	}
	msgs2 := []Message{
		*area2.welcome(),
		*area2.joinMsg(user1),
	}

	// Cleanup
	dbClient.Del(area1.token("Messages"), area2.token("Messages"), area3.token("Messages"))

	for i := range msgs1 {
		pub(t, subClient, area1.token("Updates"), &msgs1[i])
	}
	for i := range msgs2 {
		pub(t, subClient, area2.token("Updates"), &msgs2[i])
	}

	<-time.After(500 * time.Millisecond)

	// Check in the DB
	reply, err := dbClient.Lrange(area1.token("Messages"), 0, -1)
	if err != nil {
		t.Error("Error on client.Lrange.", err)
	}
	msgBytes := reply.BytesArray()
	if len(msgBytes) != 3 {
		t.Fatal("Messages stored is not 3!", len(msgBytes))
	}
	for i := range msgBytes {
		msg := NewMessage()
		Deserialize(msgBytes[i], msg)

		if msg.String() != msgs1[3-i].String() {
			t.Errorf("Message does not match expected: Exp %s, Act %s", msgs1[2-i], msg)
		}
	}

	reply, err = dbClient.Lrange(area2.token("Messages"), 0, -1)
	if err != nil {
		t.Error("Error on client.Lrange.", err)
	}
	msgBytes = reply.BytesArray()
	if len(msgBytes) != 2 {
		t.Error("Messages stored is not 2!", len(msgBytes))
	}
	for i := range msgBytes {
		msg := NewMessage()
		Deserialize(msgBytes[i], msg)

		if msg.String() != msgs2[1-i].String() {
			t.Errorf("Message does not match expected: Exp %s, Act %s", msgs2[2-i], msg)
		}
	}

	n, err := dbClient.Llen(area3.token("Messages"))
	if err != nil {
		t.Errorf("Error on Llen: %s", err.Error())
	}
	if n != 0 {
		t.Error("More than 0 messages posted to", area3.String(), "!")
	}
}
