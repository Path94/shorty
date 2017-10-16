package shorty_test

import (
	"path/filepath"
	"testing"

	"github.com/PathDNA/shorty"
	"github.com/PathDNA/testutils"
	"github.com/boltdb/bolt"
)

func TestStores(t *testing.T) {
	path, cleanup, ok := testutils.TmpDir(t, "shorty")
	if !ok {
		return
	}
	defer cleanup()

	t.Run("MemStore", func(t *testing.T) {
		testStore(t, shorty.NewMemStore())
	})

	t.Run("BoltStore", func(t *testing.T) {
		db, err := bolt.Open(filepath.Join(path, "shorty.db"), 0644, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Close()

		st, err := shorty.NewBoltStore(db)
		if err != nil {
			t.Error(err)
			return
		}

		testStore(t, st)
	})

	t.Run("TurtleStore", func(t *testing.T) {
		st, err := shorty.NewTurtleStore(path, nil, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer st.Close()
		testStore(t, st)
	})
}

func testStore(t *testing.T, st shorty.Store) {
	t.Helper()

	var (
		s    = shorty.New(st, 0x1337)
		ids  []shorty.ID
		urls = []string{"https://google.com", "https://pathdna.com", "https://meteora.co"}
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
