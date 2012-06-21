package gochat

import (
	"log"
	"runtime"
	"testing"
)

var tc Server

func init() {
	runtime.GOMAXPROCS(2)
	tc = NewServer()

	log.Println("Running Test for Server of groundfloor/textchat")
}

func createArea(t *testing.T, a *Area) {
	err := tc.CreateArea(a)
	if err != nil {
		t.Error("Error on CreateArea", err)
	}
}

func TestCreateDelete(t *testing.T) {
	area := &Area{"createdelete"}
	createArea(t, area)

	tc.DeleteArea(area)
}

func TestJoinLeave(t *testing.T) {
	area := &Area{"joinleave"}
	createArea(t, area)

	user := &User{"joiner"}
	err := tc.JoinArea(area, user)
	if err != nil {
		t.Error("Error on JoinArea.", err)
	}

	err = tc.LeaveArea(area, user)
	if err != nil {
		t.Error("Error on LeaveArea.", err)
	}

	tc.DeleteArea(area)
}

func TestTwoUsers(t *testing.T) {
	area := &Area{"twousers"}
	u1, u2 := &User{"user1"}, &User{"user2"}
	createArea(t, area)

	err := tc.JoinArea(area, u1)
	if err != nil {
		t.Error("Error on JoinArea", err)
	}

	err = tc.JoinArea(area, u2)
	if err != nil {
		t.Error("Error on JoinArea", err)
	}

	tc.LeaveArea(area, u2)
	tc.LeaveArea(area, u1)

	tc.DeleteArea(area)
}

func TestListAreas(t *testing.T) {
	a1, a2 := &Area{"area1"}, &Area{"area2"}
	createArea(t, a1)
	createArea(t, a2)

	areas, err := tc.ListAreas()
	if err != nil {
		t.Error("Error on ListAreas.", err)
	}

	var b1, b2 bool
	for i := range areas {
		if *a1 == areas[i] {
			b1 = true
		}
		if *a2 == areas[i] {
			b2 = true
		}
	}
	if !b1 {
		t.Error("ListAreas did not contain ", a1, ".", areas)
	}
	if !b2 {
		t.Error("ListAreas did not contain ", a2, ".", areas)
	}

	tc.DeleteArea(a1)
	tc.DeleteArea(a2)
}

func TestListUsers(t *testing.T) {
	area := &Area{"listusers"}
	createArea(t, area)

	u1, u2 := &User{"u1"}, &User{"u2"}
	err := tc.JoinArea(area, u1)
	if err != nil {
		t.Error("Error on JoinArea for user", u1, err)
	}

	err = tc.JoinArea(area, u2)
	if err != nil {
		t.Error("Error on JoinArea for user", u2, err)
	}

	users, err := tc.ListUsers(area)
	if err != nil {
		t.Error("Error on ListUsers:", err)
	}
	if len(users) != 2 || !(*u1 == users[0] || *u1 == users[1]) ||
		!(*u2 == users[0] || *u2 == users[1]) {
		t.Error("Bad return from ListUsers, gave back bad list:", users, ". Expected:", []User{*u1, *u2})
	}

	tc.LeaveArea(area, u1)
	tc.LeaveArea(area, u2)

	tc.DeleteArea(area)
}
