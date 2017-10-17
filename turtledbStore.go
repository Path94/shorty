package shorty

import (
	"encoding/json"

	"github.com/PathDNA/turtleDB"
	"github.com/itsmontoya/middleware"
)

type turtleStore struct {
	db turtleDB.DB
}

// NewturtleStore returns a new turtle-based store with the specified turtle.DB.
func NewTurtleStore(path string, key, iv []byte) (_ *turtleStore, err error) {
	fm := turtleDB.NewFuncsMap(turtleDB.MarshalJSON, turtleDB.UnmarshalJSON)
	fm.Put("ids", marshalData, unmarshalData)

	var ts turtleStore

	if key != nil {
		ts.db, err = turtleDB.New("shorty", path, fm, middleware.NewCryptyMW(key, iv))
	} else {
		ts.db, err = turtleDB.New("shorty", path, fm)
	}

	if err = ts.db.Update(func(tx turtleDB.Txn) error {
		tx.Create("counter")
		_, err = tx.Create("ids")
		return err
	}); err != nil {
		return
	}
	return &ts, nil
}

var _ Store = (*turtleStore)(nil)

func (bs *turtleStore) Get(id string) (*Data, error) {
	var d *Data
	if err := bs.db.Read(func(tx turtleDB.Txn) error {
		b, err := tx.Get("ids")
		if err != nil {
			return err
		}
		v, err := b.Get(id)
		if err == nil {
			d = v.(*Data).Dup()
		}
		return err
	}); err != nil {
		return nil, err
	}
	return d, nil
}

func (bs *turtleStore) Put(genIDFn func(counter uint32) ID, value *Data) error {
	return bs.db.Update(func(tx turtleDB.Txn) error {
		idsB, _ := tx.Get("ids")
		cntB, _ := tx.Get("counter")
		if !value.ID.Valid() {
			value.ID = genIDFn(bs.incCounter(cntB))
		}
		return idsB.Put(value.ID.String(), value)
	})
}

func (bs *turtleStore) Delete(ids ...string) error {
	return bs.db.Update(func(tx turtleDB.Txn) error {
		b, _ := tx.Get("ids")
		for _, id := range ids {
			if err := b.Delete(id); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bs *turtleStore) ForEach(fn func(id string, value *Data) error) error {
	return bs.db.Read(func(tx turtleDB.Txn) error {
		b, _ := tx.Get("ids")
		return b.ForEach(func(k string, v turtleDB.Value) error {
			return fn(string(k), v.(*Data).Dup())
		})
	})
}

func (bs *turtleStore) Close() error {
	return bs.db.Close()
}

func (bs *turtleStore) incCounter(b turtleDB.Bucket) (c uint32) {
	v, _ := b.Get("cnt")
	// float64 is the default type for json numbers
	if v, ok := v.(float64); ok {
		c = uint32(v)
	}
	b.Put("cnt", float64(c)+1)
	return
}

func marshalData(val turtleDB.Value) ([]byte, error) {
	d, ok := val.(*Data)

	if !ok {
		return nil, turtleDB.ErrInvalidType
	}

	return json.Marshal(d)
}

func unmarshalData(b []byte) (turtleDB.Value, error) {
	var d Data

	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}

	return &d, nil
}
