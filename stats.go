package mongersstats

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const QLimit = 20000

type QVal struct {
	value   string  `json:",omitempty"`
	byInt   int     `json:",omitempty"`
	byFloat float64 `json:",omitempty"`
}

type Stats struct {
	qInts    map[string]int     `json:",omitempty"`
	qFloats  map[string]float64 `json:",omitempty"`
	modified string             `json:",omitempty"`
	qRows    chan QVal          `json:"-"`
	lock     sync.Mutex         `json:"-"`
}

// Option is used to pass optional arguments to
// the Stats constructor
type Option interface {
	Configure(*Stats) error
}

// OptionCallback is a type of Option that is represented
// by a single function that gets called for Configure()
type OptionCallback func(*Stats) error

// Configure
func (opts OptionCallback) Configure(q *Stats) error {
	return opts(q)
}

// WithQLimit for chan qsize
func WithQLimit(m int) Option {
	return OptionCallback(func(q *Stats) error {
		//just in case
		if m > 0 {
			q.qRows = make(chan QVal, m)
		}
		return nil
	})
}

// NewQ creates a new Stats
// must be passed. Optional `Option` parameters may be passed
func NewQ(opts ...Option) (*Stats, error) {

	//init defaults here
	q := &Stats{
		modified: time.Now().Format("2006-01-02 15:04:05.999"),
		qRows:    make(chan QVal, QLimit),
		qInts:    make(map[string]int),
		qFloats:  make(map[string]float64),
	}

	//add options if any
	for _, opt := range opts {
		opt.Configure(q)
	}

	//monitor
	ready := make(chan bool, 1)
	go q.Watch(ready)
	<-ready

	//give it a try ;-)
	return q, nil
}

//Reload set to initial state
func (q *Stats) Reload() {
	//init again
	q.lock.Lock()
	defer q.lock.Unlock()

	//do the reset
	q.modified = time.Now().Format("2006-01-02 15:04:05.999")
	q.qInts = make(map[string]int)
	q.qFloats = make(map[string]float64)

	return
}

//Stringify
func (q *Stats) Stringify() string {
	//init again
	q.lock.Lock()
	defer q.lock.Unlock()

	//format it
	s := ""
	for k, v := range q.qInts {
		s += fmt.Sprintf("%-20s => %d\n", k, v)
	}
	for k, v := range q.qFloats {
		s += fmt.Sprintf("%-20s => %.08f\n", k, v)
	}
	return strings.TrimSpace(s)
}

//JSONify
func (q *Stats) JSONify() string {
	//init again
	q.lock.Lock()
	defer q.lock.Unlock()

	//format it
	var j1, j2 []byte
	if len(q.qInts) > 0 {
		j1, _ = json.MarshalIndent(q.qInts, "", "\t")
	}
	if len(q.qFloats) > 0 {
		j2, _ = json.MarshalIndent(q.qFloats, "", "\t")
	}
	return strings.TrimSpace(fmt.Sprintf("%s\n%s", string(j1), string(j2)))
}

//Incr stats by int:1
func (q *Stats) Incr(v string) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byInt: 1}
}

//IncrBy stats by int:1
func (q *Stats) IncrBy(v string, t int) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byInt: t}
}

//Decr stats by int: -1
func (q *Stats) Decr(v string) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byInt: -1}
}

//Decr stats by int:1
func (q *Stats) DecrBy(v string, t int) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byInt: -t}
}

//FloatIncr stats by decimal:1.0
func (q *Stats) FloatIncr(v string) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byFloat: 1.0}
}

//FloatIncrBy stats by decimal:n.nn
func (q *Stats) FloatIncrBy(v string, t float64) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byFloat: t * 1.0}
}

//FloatDecr stats by decimal: -1.0
func (q *Stats) FloatDecr(v string) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byFloat: -1.0}
}

//FloatDecr stats by decimal: -n.nn
func (q *Stats) FloatDecrBy(v string, t float64) {
	//dont worry about racing :-)
	q.qRows <- QVal{value: v, byFloat: -t * 1.0}
}

//Watch monitor the queue data
func (q *Stats) Watch(isReady chan bool) {
	isReady <- true
	for {
		select {
		case m := <-q.qRows:
			if m.byInt != 0 {
				q.qInts[m.value] += m.byInt
			}
			if m.byFloat != 0 {
				q.qFloats[m.value] += m.byFloat
			}
		case <-time.After(1 * time.Nanosecond):
		}
	}
}

//Dump print all available stats
func (q *Stats) Dump() {
	fmt.Println(fmt.Sprintf("%-20s => %s\n%s", "Modified", q.modified, q.SortIt()))
}

//SortIt sort the key=value pair
func (q *Stats) SortIt() string {
	//init again
	q.lock.Lock()
	defer q.lock.Unlock()

	var strs, fmtd []string

	//fmt here
	q1 := q.qInts
	q2 := q.qFloats

	//dump stats::ints
	for k, _ := range q1 {
		strs = append(strs, k)
	}
	sort.Strings(strs)
	for _, sv := range strs {
		fmtd = append(fmtd, fmt.Sprintf("%-20s => %d", sv, q1[sv]))
	}

	//dump stats::float
	strs = []string{}
	for k, _ := range q2 {
		strs = append(strs, k)
	}
	sort.Strings(strs)
	for _, sv := range strs {
		fmtd = append(fmtd, fmt.Sprintf("%-20s => %.08f", sv, q2[sv]))
	}

	//give the formatted 1
	return strings.Join(fmtd, "\n")
}

//Raw share the actual raw data
func (q *Stats) Raw() (map[string]int, map[string]float64) {
	//init again
	q.lock.Lock()
	defer q.lock.Unlock()

	//fmt here
	return q.qInts, q.qFloats
}
