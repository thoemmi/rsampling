// Obtain a random fixed sized sample from a potentially infinite stream of values.
//
//     $ time seq 0 100000000 | rsampling
//     16951800
//     65338300
//     57557813
//     65041034
//     47811733
//     5457082
//     64060517
//     88134653
//     39921145
//     39455085
//     88732734
//     58189772
//     25115415
//     41692786
//     8457525
//     26550644
//
//     real 0m16.572s
//     user 0m17.506s
//     sys  0m1.048s
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

const Version = "0.2.0"

var (
	size    = flag.Int("n", 16, "number of samples to obtain")
	seed    = flag.Int64("r", int64(time.Now().Nanosecond()), "random seed")
	version = flag.Bool("version", false, "show program version")
)

// Reservoir for strings.
type Reservoir struct {
	counter int64
	size    int
	sample  []string
}

// NewReservoir creates a reservoir for 16 elements.
func NewReservoir() *Reservoir {
	return &Reservoir{size: 16}
}

// NewReservoirSize creates a reservoir a given number of elements.
func NewReservoirSize(size int) *Reservoir {
	return &Reservoir{size: size}
}

// String print out the samples, each on one line.
func (r *Reservoir) String() string {
	return strings.Join(r.sample, "\n")
}

// Sample returns the current slice.
func (r *Reservoir) Sample() []string {
	return r.sample
}

// P returns the ratio between sample size and number of elements seen. Used to
// decide whether to store an element of not.
func (r *Reservoir) P() float64 {
	if r.counter < int64(r.size) {
		return 0
	}
	return float64(r.size) / float64(r.counter)
}

// Add fills the reservoir. If the reservoir is filled, s might be discarded.
func (r *Reservoir) Add(s string) {
	if r.counter < int64(r.size) {
		r.sample = append(r.sample, s)
	} else {
		if rand.Float64() < r.P() {
			i := rand.Intn(r.size)
			r.sample[i] = s
		}
	}
	r.counter++
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}
	rand.Seed(*seed)
	rr := NewReservoirSize(*size)
	br := bufio.NewReader(os.Stdin)

	once := sync.Once{}

	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		once.Do(func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			go func() {
				for range c {
					for _, v := range rr.Sample() {
						fmt.Println(v)
					}
				}
			}()
		})

		rr.Add(strings.TrimSpace(line))
	}

	for _, v := range rr.Sample() {
		fmt.Println(v)
	}
}
