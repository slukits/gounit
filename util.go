package gounit

import "sync"

// Fixture provides a simple concurrency fixture storage for gounit
// tests.  A Fixtures instance must not be copied after its first use.
type Fixtures struct {
	mutex sync.Mutex
	ff    map[*T]interface{}
}

// Set adds concurrency save a mapping from given test to given fixture.
func (ff *Fixtures) Set(t *T, fixture interface{}) {
	ff.mutex.Lock()
	defer ff.mutex.Unlock()
	if ff.ff == nil {
		ff.ff = map[*T]interface{}{}
	}
	ff.ff[t] = fixture
}

// Get maps given test to its fixture and returns it.
func (ff *Fixtures) Get(t *T) interface{} {
	ff.mutex.Lock()
	defer ff.mutex.Unlock()
	return ff.ff[t]
}

// Int returns given test's fixture interpreted as an int.
func (ff *Fixtures) Int(t *T) int {
	return ff.Get(t).(int)
}

// Del removes the mapping of given test to its fixture and returns the
// fixture.
func (ff *Fixtures) Del(t *T) interface{} {
	ff.mutex.Lock()
	defer ff.mutex.Unlock()
	fixture := ff.ff[t]
	delete(ff.ff, t)
	return fixture
}
