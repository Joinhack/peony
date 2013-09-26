package peony

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
