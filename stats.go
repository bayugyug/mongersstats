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
	Value   string  `json:",omitempty"`
	ByInt   int     `json:",omitempty"`
	ByFloat float64 `json:",omitempty"`
}

type Stats struct {
	QInt     map[string]int     `json:",omitempty"`
	QFloat   map[string]float64 `json:",omitempty"`
	Modified string             `json:",omitempty"`
	Queue    chan QVal          `json:"-"`
	Lock     sync.Mutex         `json:"-"`
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
			q.Queue = make(chan QVal, m)
		}
		return nil
	})
}

// NewQ creates a new Stats
// must be passed. Optional `Option` parameters may be passed
func NewQ(opts ...Option) (*Stats, error) {

	//init defaults here
	q := &Stats{
		Modified: time.Now().Format("2006-01-02 15:04:05.999"),
		Queue:    make(chan QVal, QLimit),
		QInt:     make(map[string]int),
		QFloat:   make(map[string]float64),
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
	q.Lock.Lock()
	defer q.Lock.Unlock()

	//do the reset
	q.Modified = time.Now().Format("2006-01-02 15:04:05.999")
	q.QInt = make(map[string]int)
	q.QFloat = make(map[string]float64)

	return
}

//Stringify
func (q *Stats) Stringify() string {
	//init again
	q.Lock.Lock()
	defer q.Lock.Unlock()

	//format it
	s := ""
	for k, v := range q.QInt {
		s += fmt.Sprintf("%-20s => %d\n", k, v)
	}
	for k, v := range q.QFloat {
		s += fmt.Sprintf("%-20s => %.08f\n", k, v)
	}
	return strings.TrimSpace(s)
}

//JSONify
func (q *Stats) JSONify() string {
	//init again
	q.Lock.Lock()
	defer q.Lock.Unlock()

	//format it
	var j1, j2 []byte
	if len(q.QInt) > 0 {
		j1, _ = json.MarshalIndent(q.QInt, "", "\t")
	}
	if len(q.QFloat) > 0 {
		j2, _ = json.MarshalIndent(q.QFloat, "", "\t")
	}
	return strings.TrimSpace(fmt.Sprintf("%s\n%s", string(j1), string(j2)))
}

//Incr stats by int:1
func (q *Stats) Incr(v string) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByInt: 1}
}

//IncrBy stats by int:1
func (q *Stats) IncrBy(v string, t int) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByInt: t}
}

//Decr stats by int: -1
func (q *Stats) Decr(v string) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByInt: -1}
}

//Decr stats by int:1
func (q *Stats) DecrBy(v string, t int) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByInt: -t}
}

//FloatIncr stats by decimal:1.0
func (q *Stats) FloatIncr(v string) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByFloat: 1.0}
}

//FloatIncrBy stats by decimal:n.nn
func (q *Stats) FloatIncrBy(v string, t float64) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByFloat: t * 1.0}
}

//FloatDecr stats by decimal: -1.0
func (q *Stats) FloatDecr(v string) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByFloat: -1.0}
}

//FloatDecr stats by decimal: -n.nn
func (q *Stats) FloatDecrBy(v string, t float64) {
	//dont worry about racing :-)
	q.Queue <- QVal{Value: v, ByFloat: -t * 1.0}
}

//Watch monitor the queue data
func (q *Stats) Watch(isReady chan bool) {
	isReady <- true
	for {
		select {
		case m := <-q.Queue:
			if m.ByInt != 0 {
				q.QInt[m.Value] += m.ByInt
			}
			if m.ByFloat != 0 {
				q.QFloat[m.Value] += m.ByFloat
			}
		case <-time.After(1 * time.Nanosecond):
		}
	}
}

//Dump print all available stats
func (q *Stats) Dump() {
	fmt.Println(fmt.Sprintf("%-20s => %s", "Modified", q.Modified))
	fmt.Println(q.SortIt())
	return
}

//SortIt sort the key=value pair
func (q *Stats) SortIt() string {
	//init again
	q.Lock.Lock()
	defer q.Lock.Unlock()

	var strs, fmtd []string

	//fmt here
	q1 := q.QInt
	q2 := q.QFloat

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
	q.Lock.Lock()
	defer q.Lock.Unlock()

	//fmt here
	return q.QInt, q.QFloat
}
