package shorty_test

import (
	"encoding/json"
	"log"
	"math"
	"testing"
	"time"

	"github.com/PathDNA/shorty"
)

func init() {
	log.SetFlags(log.Lshortfile)
}
func TestID(t *testing.T) {
	id := shorty.NewID(math.MaxUint32, math.MaxInt64-shorty.BaseTS, math.MaxUint32)
	t.Logf("%#+v: %s", id, id)
	nid, _ := shorty.IDFromString(id.String())
	if nid != id {
		t.Fatalf("%s (%#+v) != %s (%#+v)", id, id, nid, nid)
	}

	var jt struct {
		I1 shorty.ID
		I2 *shorty.ID
		I3 *shorty.ID
		I4 shorty.ID
	}

	jt.I1, jt.I2 = id, &id

	j, err := json.Marshal(jt)

	if err != nil {
		t.Fatal(err)
	}
	jt.I1, jt.I2 = shorty.ID{}, nil

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
		s    = shorty.New(shorty.NewMemStore(), 1)
		ids  []shorty.ID
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

	s.ForEach(func(d *shorty.Data) error {
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
	s := shorty.New(shorty.NewMemStore(), 1)
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
