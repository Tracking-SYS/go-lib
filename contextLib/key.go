package contextLib

type Key interface {
	String() string
}

func NewKey(name string) Key {
	return &key{name}
}

type key struct {
	name string
}

func (k *key) String() string {
	return "go-tracking_" + k.name
}
