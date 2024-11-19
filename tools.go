package worker_pool

import (
	"fmt"
	"strings"
	"sync"
)

const (
	Ok    = 0x2713
	Error = 0x2715
)

func chanSpeaker(ch chan<- string, str string, times int, stop <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < times; i++ {
		select {
		case ch <- str:
		case <-stop:
			return
		}
	}
}

type Set map[interface{}]struct{}

func NewSet() Set {
	return make(Set)
}

func (s Set) Add(element interface{}) {
	s[element] = struct{}{}
}

func (s Set) Contains(element interface{}) bool {
	_, exists := s[element]
	return exists
}

func (s1 Set) Equals(s2 Set) bool {
	if len(s1) != len(s2) {
		return false
	}

	for elem := range s1 {
		if !s2.Contains(elem) {
			return false
		}
	}
	return true
}

func (s Set) String() string {
	if len(s) == 0 {
		return "{}"
	}

	var elements []string
	for elem := range s {
		elements = append(elements, fmt.Sprintf("%v", elem))
	}

	return "{" + strings.Join(elements, ", ") + "}"
}

// Implement Writer
type buffer []string

func (b *buffer) Write(p []byte) (n int, err error) {
	*b = append(*b, string(p))
	n = len(p)
	return
}
