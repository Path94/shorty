package shorty

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	// BaseTS is time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	BaseTS int64 = 1483228800000000000 / int64(time.Second)

	// this is b61 alphabet, a-zA-Z1-9, 0 is used a separator
	b61Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
	base        = uint64(len(b61Alphabet))
)

var (
	decoder [math.MaxUint8]byte
)

func init() {
	for i := range decoder {
		decoder[i] = 0xFF
	}
	for i, c := range b61Alphabet {
		decoder[c] = byte(i)
	}
}

// ID represents a unique ID with a theoretical limit of `math.MaxUint32` IDs per second per machine. up to year 2153.
// - 4 bytes for the timestamp
// - 4 bytes for the machine id
// - 8 bytes for the counter
type ID struct {
	ts        uint32
	machineID uint32
	counter   uint64
}

func IDFromString(id string) (*ID, error) {
	ps := strings.Split(id, "0")
	if len(ps) != 3 {
		return nil, fmt.Errorf("%x is an invalid id")
	}
	return &ID{
		counter:   b61Decode(ps[2]),
		machineID: uint32(b61Decode(ps[1])),
		ts:        uint32(b61Decode(ps[0])),
	}, nil
}

// NewID returns a new ID with the spcified machine ID, timestamp and counter.
// TS must be > BaseTS.
// Example: id := NewID(0xB00F, time.Now().Unix(), 1)
func NewID(machineID uint32, ts int64, counter uint64) *ID {
	return &ID{
		ts:        uint32(ts - BaseTS),
		machineID: machineID,
		counter:   counter,
	}
}

// String returns the base62 representation of the ID, howe
func (id *ID) String() string {
	var (
		out = make([]byte, 6+6+6+2)
		i   = len(out) - 1
	)

	i = b61Encode(i, id.counter, out)
	out[i] = '0'
	i = b61Encode(i-1, uint64(id.machineID), out[:i])
	out[i] = '0'
	i = b61Encode(i-1, uint64(id.ts), out[:i])

	return string(out[i+1:])
}

func (id *ID) Time() time.Time {
	return time.Unix(int64(id.ts)+BaseTS, 0).UTC()
}

func (id *ID) MachineID() uint32 { return id.machineID }

func (id *ID) Counter() uint64 { return id.counter }

// b61Encode converts in to base61 in-place and returns the number of bytes used
func b61Encode(i int, in uint64, out []byte) int {
	for ; in > 0; i-- {
		out[i] = b61Alphabet[in%base]
		in = in / base
	}

	return i
}

func b61Decode(s string) uint64 {
	var (
		o        uint64
		alphaLen = float64(base)
		sLen     = float64(len(s))
	)

	for i, r := range s {
		pow := sLen - (float64(i) + 1)
		o += uint64(decoder[r]) * uint64(math.Pow(alphaLen, pow))
	}
	return o
}
