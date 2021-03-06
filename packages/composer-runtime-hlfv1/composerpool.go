/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import "sync"

// ComposerPool holds a pool of Composer objects.
type ComposerPool struct {
	Pool      chan *Composer
	PoolMutex *sync.RWMutex
	PoolCount int
	PoolMax   int
}

// NewComposerPool creates a new pool of Composer objects.
func NewComposerPool(max int) (result *ComposerPool) {
	logger.Debug("Entering NewComposerPool", max)
	defer func() { logger.Debug("Exiting NewComposerPool", result) }()

	result = &ComposerPool{
		Pool:      make(chan *Composer, max),
		PoolCount: 0,
		PoolMax:   max,
		PoolMutex: &sync.RWMutex{},
	}
	return result
}

// Get returns an existing Composer object from the pool, or creates a new one
// if no existing Composer objects are available.
func (cp *ComposerPool) Get() (result *Composer) {
	logger.Debug("Entering ComposerPool.Get")
	defer func() { logger.Debug("Exiting ComposerPool.Get", result) }()

	// lock the pool and check to see how many Composer objects
	// have been created - create a new one if we haven't hit the max yet
	cp.PoolMutex.RLock()
	if cp.PoolCount < cp.PoolMax {
		cp.PoolMutex.RUnlock()
		cp.PoolMutex.Lock()
		defer cp.PoolMutex.Unlock()
		if cp.PoolCount < cp.PoolMax {
			result := NewComposer()
			result.Index = cp.PoolCount
			logger.Debug("Creating a new Composer object for pool", result.Index)
			cp.PoolCount++
			return result
		}
	} else {
		cp.PoolMutex.RUnlock()
	}

	// we will get the newly created one, or wait for one. Potentially
	// we could have had one put back during this time which we could
	// have used if a new one was created, but never mind.
	result = <-cp.Pool
	logger.Debug("Got Composer object from pool", result.Index)
	return result
}

// Put stores an existing Composer object in the pool, or discards it if the pool
// is currently at capacity.
func (cp *ComposerPool) Put(composer *Composer) (result bool) {
	logger.Debug("Entering ComposerPool.Put", composer)
	defer func() { logger.Debug("Exiting ComposerPool.Put", result) }()

	logger.Debug("Putting Composer object into pool", composer.Index)
	cp.Pool <- composer
	return true
}
