package sets

import "encoding/json"

func initOrderedSet[T HasKey]() *OrderedSet[T] {
	return &OrderedSet[T]{set: map[string]int{}}
}

func NewOrderedSet[T HasKey](xs ...T) *OrderedSet[T] {
	set := initOrderedSet[T]()
	for _, x := range xs {
		set.Add(x)
	}
	return set
}

type HasKey interface {
	Key() string
}

type OrderedSet[T HasKey] struct {
	items []T
	set   map[string]int
}

func (s *OrderedSet[T]) Add(x T) {
	if _, found := s.set[x.Key()]; found {
		return
	}
	s.items = append(s.items, x)
	s.set[x.Key()] = len(s.items) - 1
}

func (s *OrderedSet[T]) Len() int {
	return len(s.items)
}

func (s *OrderedSet[T]) Items() []T {
	return s.items
}

func (s *OrderedSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Items())
}

func (s *OrderedSet[T]) UnmarshalJSON(b []byte) error {
	var vals []T
	if err := json.Unmarshal(b, &vals); err != nil {
		return err
	}
	xs := &OrderedSet[T]{set: map[string]int{}}
	for _, x := range vals {
		xs.Add(x)
	}
	*s = *xs
	return nil
}

func (s *OrderedSet[T]) Has(x T) bool {
	return s.HasKey(x.Key())
}

func (s *OrderedSet[T]) HasKey(key string) bool {
	_, ok := s.set[key]
	return ok
}
