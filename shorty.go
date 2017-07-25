package shorty

import (
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
	return &Shorty{
		s:         store,
		machineID: machineID,
	}
}

// GenerateID generates a unique id for the provided url.
func (s *Shorty) GenerateID(url string) (ID, error) {
	_, err := U.Parse(url)
	if err != nil {
		return ID{}, err
	}
	d := &Data{URL: url}
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
	ID   ID                     `json:"id,omitempty"`
	URL  string                 `json:"url,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"` // just a placeholder until we define
}
