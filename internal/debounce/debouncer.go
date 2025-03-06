/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package debounce

import (
	"sync"
)

type AtMostOnceEvent struct {
	Id  string
	Err error
}

// Debouncer manages a WaitGroup with at-most-once semantics for a given ID.
type Debouncer interface {
	// Register records the ID for later debouncing.
	Register(wg *sync.WaitGroup, id string)

	// Notify decrements the WaitGroup if the event's ID is registered,
	// otherwise it is dropped.
	// If the ID was registered, the ID is deregistered to avoid future
	// decrements for the same ID.
	Notify(wg *sync.WaitGroup, event AtMostOnceEvent)
}

type registrations struct {
	mutex *sync.Mutex
	ids   map[string]bool
}

// Build creates a new Debouncer
func Build() Debouncer {
	return &registrations{
		mutex: &sync.Mutex{},
		ids:   make(map[string]bool),
	}
}

func (d *registrations) Register(wg *sync.WaitGroup, id string) {
	d.mutex.Lock()
	d.ids[id] = true
	d.mutex.Unlock()
	wg.Add(1)
}

func (d *registrations) Notify(wg *sync.WaitGroup, event AtMostOnceEvent) {
	var removed = false
	d.mutex.Lock()
	if d.ids[event.Id] {
		delete(d.ids, event.Id)
		removed = true
	}
	d.mutex.Unlock()
	if removed {
		wg.Done()
	}
}
