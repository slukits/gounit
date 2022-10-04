// Package race fails if the race detector is turned on and passes if it
// is turned of.
package race

import "sync"

func RacingFunc() {
	j, race, wg := 0, make(chan struct{}), sync.WaitGroup{}
	wg.Add(1)
	go func() {
		<-race
		j++
		wg.Done()
	}()

	j++
	close(race)
	j++
	wg.Wait()
}
