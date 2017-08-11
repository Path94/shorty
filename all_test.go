package shorty

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestID(t *testing.T) {
	id := NewID(math.MaxUint32, math.MaxInt64-BaseTS, math.MaxUint32)
	t.Logf("%#+v: %s", id, id)
	nid, _ := IDFromString(id.String())
	if nid != id {
		t.Fatalf("%s (%#+v) != %s (%#+v)", id, id, nid, nid)
	}

	var jt struct {
		I1 ID
		I2 *ID
		I3 *ID
		I4 ID
	}

	jt.I1, jt.I2 = id, &id

	j, err := json.Marshal(jt)

	if err != nil {
		t.Fatal(err)
	}
	jt.I1, jt.I2 = ID{}, nil

	if err := json.Unmarshal(j, &jt); err != nil {
		t.Fatal(err)
	}
	if jt.I1 != id || *jt.I2 != id {
		t.Fatalf("%s (%#+v) != %s (%#+v)", id, id, nid, nid)
	}

	t.Log(id.Time())
}

func TestShorty(t *testing.T) {
	var (
		s    = New(NewMemStore(), 1)
		ids  []ID
		urls = []string{"https://google.com", "https://path94.com", "https://meteora.co"}
	)

	for _, url := range urls {
		id, err := s.GenerateID(url)
		if err != nil {
			t.Errorf("%v", err)
		}
		ids = append(ids, id)
		t.Logf("%s: %s", id, url)
	}

	s.ForEach(func(d *Data) error {
		i := d.ID.Counter()
		if int(i) > len(urls) {
			t.Fatalf("i > len(urls): %d %v", i, d)
		}

		if d.URL != urls[i] {
			t.Fatalf("expected %s, got %s", urls[i], d.URL)
		}
		if d.ID != ids[i] {
			t.Fatalf("expected %s, got %s", urls[i], d.URL)
		}
		return nil
	})

}

func TestExpiry(t *testing.T) {
	var s Shorty
	s.s = NewMemStore() // have to do it this way so the auto clean up doesn't start
	id, err := s.GenerateTimedID("http://google.com", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	n, err := s.PurgeExpired()

	if n != 1 || err != nil {
		t.Fatalf("unexpected: %d, %v", n, err)
	}

	if url := s.GetURL(id.String()); url != "" {
		t.Fatalf("expected the url to be deleted, got %s", url)
	}
}
