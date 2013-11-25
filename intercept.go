package peony

type Point int

const (
	BEFORE Point = iota
	AFTER
	FINALLY
	DEFER
)

type Interceptor struct {
	Point  Point
	Action Action
}

func (i *Interceptor) name() {

}
