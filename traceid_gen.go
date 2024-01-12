package tracing

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var (
	seededIDGen  = rand.New(rand.NewSource(time.Now().UnixNano()))
	seededIDLock sync.Mutex
)

type ID uint64

func (i ID) String() string {
	return fmt.Sprintf("%016x", uint64(i))
}

type TraceID struct {
	High uint64
	Low  uint64
}

func (t TraceID) Empty() bool {
	return t.Low == 0 && t.High == 0
}

type IDGenerator interface {
	SpanID(traceID TraceID) ID
	TraceID() TraceID
}

type randomID64 struct{}

func (r *randomID64) TraceID() (id TraceID) {
	seededIDLock.Lock()
	id = TraceID{
		Low: uint64(seededIDGen.Int63()),
	}
	seededIDLock.Unlock()
	return
}

func (t TraceID) String() string {
	if t.High == 0 {
		return fmt.Sprintf("%016x", t.Low)
	}

	return fmt.Sprintf("%016x%016x", t.High, t.Low)
}

func (r *randomID64) SpanID(traceID TraceID) (id ID) {
	if !traceID.Empty() {
		return ID(traceID.Low)
	}
	seededIDLock.Lock()
	id = ID(seededIDGen.Int63())
	seededIDLock.Unlock()
	return id
}

func NewRandom64() IDGenerator {
	return &randomID64{}
}

func TraceIDFromHex(h string) (t TraceID, err error) {
	if len(h) > 16 {
		if t.High, err = strconv.ParseUint(h[0:len(h)-16], 16, 64); err != nil {
			return
		}
		t.Low, err = strconv.ParseUint(h[len(h)-16:], 16, 64)
		return
	}
	t.Low, err = strconv.ParseUint(h, 16, 64)
	return
}

// MarshalJSON serializes an ID type (SpanID, ParentSpanID) to HEX.
func (i ID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", i.String())), nil
}

// UnmarshalJSON deserializes an ID type (SpanID, ParentSpanID) from HEX.
func (i *ID) UnmarshalJSON(b []byte) (err error) {
	var id uint64
	if len(b) < 3 {
		return nil
	}
	id, err = strconv.ParseUint(string(b[1:len(b)-1]), 16, 64)
	*i = ID(id)
	return err
}

func NewRandom128() IDGenerator {
	return &randomID128{}
}

func NewRandomTimestamped() IDGenerator {
	return &randomTimestamped{}
}

type randomID128 struct{}

func (r *randomID128) TraceID() (id TraceID) {
	seededIDLock.Lock()
	id = TraceID{
		High: uint64(seededIDGen.Int63()),
		Low:  uint64(seededIDGen.Int63()),
	}
	seededIDLock.Unlock()
	return
}

func (r *randomID128) SpanID(traceID TraceID) (id ID) {
	if !traceID.Empty() {
		return ID(traceID.Low)
	}
	seededIDLock.Lock()
	id = ID(seededIDGen.Int63())
	seededIDLock.Unlock()
	return
}

type randomTimestamped struct{}

func (t *randomTimestamped) TraceID() (id TraceID) {
	seededIDLock.Lock()
	id = TraceID{
		High: uint64(time.Now().Unix()<<32) + uint64(seededIDGen.Int31()),
		Low:  uint64(seededIDGen.Int63()),
	}
	seededIDLock.Unlock()
	return
}

func (t *randomTimestamped) SpanID(traceID TraceID) (id ID) {
	if !traceID.Empty() {
		return ID(traceID.Low)
	}
	seededIDLock.Lock()
	id = ID(seededIDGen.Int63())
	seededIDLock.Unlock()
	return
}
