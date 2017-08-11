package shorty

import (
	"log"
	U "net/url"
	"time"
)

// Shorty is a simple URL shortner, using unique base61 ids.
type Shorty struct {
	s Store
	// the counter will overflow to 0 eventually, however since each id has a timestamp, that's not an issue.
	machineID uint32
}

// New returns a new shorty, machineID can be any unique id to embed in the generated ids.
func New(store Store, machineID uint32) *Shorty {
	s := &Shorty{
		s:         store,
		machineID: machineID,
	}
	go s.purgeLoop()
	return s
}

func (s *Shorty) purgeLoop() {
	for {
		if _, err := s.PurgeExpired(); err != nil {
			log.Println(err)
			break
		}
		time.Sleep(time.Hour)
	}
}

// PurgeExpired purges all expired IDs and returns the number or an error.
func (s *Shorty) PurgeExpired() (int, error) {
	var (
		now     = time.Now()
		expired []string
	)
	if err := s.s.ForEach(func(id string, v *Data) error {
		if v.Expired(now) {
			expired = append(expired, id)
		}
		return nil
	}); err != nil {
		return 0, err
	}
	if len(expired) > 0 {
		if err := s.s.Delete(expired...); err != nil {
			return 0, err
		}
	}
	return len(expired), nil
}

// GenerateID generates a unique id for the provided url.
func (s *Shorty) GenerateID(url string) (id ID, err error) {
	return s.GenerateTimedID(url, -1)
}

// GenerateTimedID generates a unique id for the provided url, the id gets deleted after maxAge.
// if maxAge <= 0, it's ignored.
func (s *Shorty) GenerateTimedID(url string, maxAge time.Duration) (id ID, err error) {
	if _, err = U.Parse(url); err != nil {
		return
	}
	d := &Data{
		URL: url,
	}
	if maxAge > 0 {
		d.MaxAge = maxAge
	}
	err = s.s.Put(s.genID, d)
	return d.ID, err
}

// Get returns the Data of the given ID or nil if it doesn't exist.
func (s *Shorty) Get(id string) *Data {
	d, err := s.s.Get(id)
	if err != nil {
		return nil
	}
	return d
}

// GetURL returns the url associcated with the id or empty.
func (s *Shorty) GetURL(id string) string {
	if d := s.Get(id); d != nil {
		return d.URL
	}
	return ""
}

// ForEach loops over all the stored urls.
func (s *Shorty) ForEach(fn func(d *Data) error) error {
	return s.s.ForEach(func(id string, v *Data) error {
		return fn(v)
	})
}

func (s *Shorty) genID(c uint32) ID {
	return NewID(s.machineID, time.Now().Unix(), c)
}

// Data is the internal representation of what Shorty stores.
type Data struct {
	ID     ID                     `json:"id,omitempty"`
	URL    string                 `json:"url,omitempty"`
	MaxAge time.Duration          `json:"maxAge,omitempty"` // the url will get deleted if time.Now() > id.Time() + maxAge
	Meta   map[string]interface{} `json:"meta,omitempty"`   // just a placeholder until we define
}

func (d *Data) Expired(t time.Time) bool {
	if d.MaxAge < 1 {
		return false
	}
	return t.After(d.ID.Time().Add(d.MaxAge))
}
