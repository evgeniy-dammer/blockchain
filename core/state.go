package core

import "fmt"

// State
type State struct {
	data map[string][]byte
}

// NewState is a constructor for the State
func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

// Put puts the given kay and value into state.data
func (s *State) Put(key, value []byte) error {
	s.data[string(key)] = value

	return nil
}

// Delete deletes the value from state.data with given key
func (s *State) Delete(key []byte) error {
	delete(s.data, string(key))

	return nil
}

// Get returns value with the given key
func (s *State) Get(k []byte) ([]byte, error) {
	key := string(k)

	value, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("given key %s not found", key)
	}

	return value, nil
}
