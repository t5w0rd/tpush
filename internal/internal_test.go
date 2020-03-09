package internal

import (
	"testing"
)

func TestBIndex_RemoveUser(t *testing.T) {
	bi := NewBIndex()
	bi.AddUserTag("LJ", "NB")
	bi.AddUserTag("LJ", "SWITCH")
	bi.AddUserTag("CJL", "NB")
	bi.AddUserTag("CJL", "KOF")

	bi.RemoveTag("NB")

	var s []interface{}
	bi.Tags("LJ", &s)
	t.Log(s)

	bi.Users("KOF", &s)
	t.Log(s)

	bi.AllTags(&s)
	t.Log(s)

	bi.AllUsers(&s)
	t.Log(s)
}
