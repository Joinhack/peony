package peony

type Point int

const (
	BEFORE Point = iota
	AFTER
	FINALLY
	DEFER
)

type Interceptors struct {
	Point  Point
	Action Action
}

func (i *Interceptors) name() {

}
