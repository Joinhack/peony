package peony

import (
	"io/ioutil"
	"strings"
)

type Set map[interface{}]bool

func (s *Set) Add(a interface{}) {
	(*s)[a] = true
}

func (s *Set) Has(a interface{}) bool {
	if (*s)[a] == true {
		return true
	} else {
		return false
	}
}

func NewSet() Set {
	return make(Set, 4)
}

// Reads the lines of the given file.  Panics in the case of error.
func MustReadLines(filename string) []string {
	r, err := ReadLines(filename)
	if err != nil {
		panic(err)
	}
	return r
}

// Reads the lines of the given file.  Panics in the case of error.
func ReadLines(filename string) ([]string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), "\n"), nil
}
