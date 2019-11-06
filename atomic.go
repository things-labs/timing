// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// atomic provides simple wrappers around numerics to enforce atomic access.
// @see https://github.com/uber-go/atomic

package timing

import (
	"sync/atomic"
	"time"
)

// Int32 is an atomic wrapper around an int32.
type Int32 struct {
	v int32
}

// NewInt32 creates an Int32.
func NewInt32(i int32) *Int32 {
	return &Int32{i}
}

// Load atomically loads the wrapped value.
func (i *Int32) Load() int32 {
	return atomic.LoadInt32(&i.v)
}

// Add atomically adds to the wrapped int32 and returns the new value.
func (i *Int32) Add(n int32) int32 {
	return atomic.AddInt32(&i.v, n)
}

// Sub atomically subtracts from the wrapped int32 and returns the new value.
func (i *Int32) Sub(n int32) int32 {
	return atomic.AddInt32(&i.v, -n)
}

// CAS is an atomic compare-and-swap.
func (i *Int32) CompareAndSwapInt64(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&i.v, old, new)
}

// Store atomically stores the passed value.
func (i *Int32) Store(n int32) {
	atomic.StoreInt32(&i.v, n)
}

// Swap atomically swaps the wrapped int32 and returns the old value.
func (i *Int32) Swap(n int32) int32 {
	return atomic.SwapInt32(&i.v, n)
}

type Duration struct {
	v int64
}

// NewDuration new duration
func NewDuration(d time.Duration) *Duration {
	return &Duration{int64(d)}
}

// Load atomically loads the wrapped value.
func (sf *Duration) Load() time.Duration {
	return time.Duration(atomic.LoadInt64(&sf.v))
}

// Store atomically stores the passed value.
func (sf *Duration) Store(n time.Duration) {
	atomic.StoreInt64(&sf.v, int64(n))
}

// Add atomically adds to the wrapped time.Duration and returns the new value.
func (sf *Duration) Add(n time.Duration) time.Duration {
	return time.Duration(atomic.AddInt64(&sf.v, int64(n)))
}

// Sub atomically subtracts from the wrapped time.Duration and returns the new value.
func (sf *Duration) Sub(n time.Duration) time.Duration {
	return time.Duration(atomic.AddInt64(&sf.v, int64(n)))
}

// Swap atomically swaps the wrapped time.Duration and returns the old value.
func (sf *Duration) Swap(n time.Duration) time.Duration {
	return time.Duration(atomic.SwapInt64(&sf.v, int64(n)))
}

// CompareAndSwapInt64 is an atomic compare-and-swap.
func (sf *Duration) CompareAndSwapInt64(old, new time.Duration) bool {
	return atomic.CompareAndSwapInt64(&sf.v, int64(old), int64(new))
}
