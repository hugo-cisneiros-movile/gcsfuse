// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ratelimit

import (
	"fmt"
	"math"
	"time"

	"github.com/jacobsa/gcloud/syncutil"
)

// A measurement of the amount of real time since some fixed epoch.
//
// TokenBucket doesn't care about calendar time, time of day, etc.
// Unfortunately time.Time takes these things into account, and in particular
// time.Now() is not monotonic -- it may jump arbitrarily far into the future
// or past when the system's wall time is changed.
//
// Instead we reckon in terms of a monotonic measurement of time elapsed since
// the bucket was initialized, and leave it up to the user to provide this. See
// SystemTimeTokenBucket for a convenience in doing so.
type MonotonicTime time.Duration

// A bucket of tokens that refills at a specific rate up to a particular
// capacity. Users can remove tokens in sizes up to that capacity, can are told
// how long they should wait before proceeding.
//
// If users cooperate by waiting to take whatever action they are rate limiting
// as told by the token bucket, the overall action rate will be limited to the
// token bucket's fill rate.
//
// Safe for concurrent access.
//
// Cf. http://en.wikipedia.org/wiki/Token_bucket
type TokenBucket interface {
	// Return the maximum number of tokens that the bucket can hold.
	Capacity() (c uint64)

	// Remove the specified number of tokens from the token bucket at the given
	// time. The user should wait until sleepUntil before proceeding in order to
	// obey the rate limit.
	//
	// REQUIRES: tokens <= Capacity()
	Remove(
		now MonotonicTime,
		tokens uint64) (sleepUntil MonotonicTime)
}

// Choose a token bucket capacity that ensures that the action gated by the
// token bucket will be limited to within a few percent of `rateHz * window`
// for any window of the given size.
//
// This is not be possible for all rates and windows. In that case, an error
// will be returned.
func ChooseTokenBucketCapacity(
	rateHz float64,
	window time.Duration) (capacity uint64, err error) {
	// Check that the input is reasonable.
	if rateHz <= 0 || math.IsInf(rateHz, 0) {
		err = fmt.Errorf("Illegal rate: %f", rateHz)
		return
	}

	if window <= 0 {
		err = fmt.Errorf("Illegal window: %v", window)
		return
	}

	// We cannot help but allow the rate to exceed the configured maximum by some
	// factor in an arbitrary window, no matter how small we scale the max
	// accumulated credit -- the bucket may be full at the start of the window,
	// be immediately exhausted, then be repeatedly exhausted just before filling
	// throughout the window.
	//
	// For example: let the window W = 10 seconds, and the bandwidth B = 20 MiB/s.
	// Set the max accumulated credit C = W*B/2 = 100 MiB. Then this
	// sequence of events is allowed:
	//
	//  *  T=0:        Allow through 100 MiB.
	//  *  T=4.999999: Allow through nearly 100 MiB.
	//  *  T=9.999999: Allow through nearly 100 MiB.
	//
	// Therefore we exceed the allowed bytes for the window by nearly 50%. Note
	// however that this trend cannot continue into the next window, so this must
	// be a transient spike.
	//
	// In general if we set C <= W*B/N, then we're off by no more than a factor
	// of (N+1)/N within any window of size W.
	//
	// Choose a reasonable N.
	const N = 50

	capacityFloat := math.Floor(rateHz * (float64(window) / float64(time.Second)))
	if !(capacityFloat > 0 && capacityFloat < float64(math.MaxUint64)) {
		err = fmt.Errorf(
			"Can't use a token bucket to limit to %f Hz over a window of %v "+
				"(result is a capacity of %f)",
			rateHz,
			window,
			capacityFloat)
	}

	capacity = uint64(capacityFloat)
	if capacity == 0 {
		panic(fmt.Sprintf(
			"Calculated a zero capacity for inputs %f, %v",
			rateHz,
			window))
	}

	return
}

// Create a token bucket that fills at the given rate in tokens per second, up
// to the given capacity. ChooseTokenBucketCapacity may help you decide on a
// capacity.
//
// REQUIRES: rateHz > 0
// REQUIRES: capacity > 0
func NewTokenBucket(
	rateHz float64,
	capacity uint64) (tb TokenBucket) {
	typed := &tokenBucket{
		rateHz:   rateHz,
		capacity: capacity,
	}

	typed.mu = syncutil.NewInvariantMutex(typed.checkInvariants)

	tb = typed
	return
}

////////////////////////////////////////////////////////////////////////
// Implementation
////////////////////////////////////////////////////////////////////////

type tokenBucket struct {
	/////////////////////////
	// Constant data
	/////////////////////////

	rateHz   float64
	capacity uint64

	/////////////////////////
	// Mutable state
	/////////////////////////

	mu syncutil.InvariantMutex
}

func (tb *tokenBucket) checkInvariants() {
	panic("TODO")
}

func (tb *tokenBucket) Capacity() (c uint64) {
	c = tb.capacity
	return
}

func (tb *tokenBucket) Remove(
	now MonotonicTime,
	tokens uint64) (sleepUntil MonotonicTime) {
	panic("TODO")
}
