package shorty

import (
	"testing"
	"time"
)

func TestID(t *testing.T) {
	ts := time.Now()
	id := NewID(0xB00B, ts.Unix(), 0xDEAD)
	t.Logf("%#+v: %s", id, id)
	nid, _ := IDFromString(id.String())
	if *nid != *id {
		t.Fatalf("%s != %s", id, nid)
	}
}
