package treelock

import (
	"sort"
	"strings"
	"sync"
)

type TreeLock struct {
	totalLock *sync.Mutex
	totalwg   *sync.WaitGroup
	sep       rune
	locks     map[string]*TreeLock
}

type SimpleTreeLock struct {
	totalLock *sync.Mutex
	totalwg   *sync.WaitGroup
	locks     map[string]*sync.Mutex
}

// Create a new tree lock with the separator 'sep'
func NewTreeLock(sep rune) *TreeLock {
	return &TreeLock{new(sync.Mutex), new(sync.WaitGroup), sep, make(map[string]*TreeLock)}
}

// Safely lock a value to prevent threads from accessing this value
func (T TreeLock) Lock(val string) {
	b := strings.IndexRune(val, T.sep)
	T.totalLock.Lock()
	if b == -1 {
		if _, ok := T.locks[val]; !ok {
			T.locks[val] = NewTreeLock(T.sep)
		}
	} else {
		if _, ok := T.locks[val[:b]]; !ok {
			T.locks[val[:b]] = NewTreeLock(T.sep)
		}
	}
	T.totalwg.Add(1)
	T.totalLock.Unlock()
	if b == -1 {
		T.locks[val].LockAll()
	} else {
		T.locks[val[:b]].Lock(val[b+1:])
	}
}

// Unlock a value to allow it to be used by another thread.
// Will panic if the value is not locked
func (T TreeLock) Unlock(val string) {
	T.totalwg.Done()
	b := strings.IndexRune(val, T.sep)
	if b == -1 {
		T.locks[val].UnlockAll()
	} else {
		T.locks[val[:b]].Unlock(val[b+1:])
	}
}

// Safely lock multiple values simultaneously while preventing race condition
func (T TreeLock) LockMany(vals ...string) {
	sort.Strings(vals)
	bmap := map[string]int{}
	for _, val := range vals {
		bmap[val] = strings.IndexRune(val, T.sep)
	}
	T.totalLock.Lock()
	for _, val := range vals {
		b := bmap[val]
		if b == -1 {
			if _, ok := T.locks[val]; !ok {
				T.locks[val] = NewTreeLock(T.sep)
			}
		} else {
			if _, ok := T.locks[val[:b]]; !ok {
				T.locks[val[:b]] = NewTreeLock(T.sep)
			}
		}
	}
	T.totalwg.Add(len(vals))
	T.totalLock.Unlock()
	for _, val := range vals {
		b := bmap[val]
		if b == -1 {
			T.locks[val].LockAll()
		} else {
			T.locks[val[:b]].Lock(val[b+1:])
		}
	}
}

// Safely unlock multiple values simultaneously while preventing race condition
func (T TreeLock) UnlockMany(vals ...string) {
	sort.Strings(vals)
	for i := len(vals) - 1; i >= 0; i-- {
		val := vals[i]
		b := strings.IndexRune(val, T.sep)
		if b == -1 {
			T.locks[val].UnlockAll()
		} else {
			T.locks[val[:b]].Unlock(val[b+1:])
		}
		T.totalwg.Done()
	}
}

// Lock the entire tree, waits for all existing locks to unlock first.
func (T TreeLock) LockAll() {
	T.totalLock.Lock()
	T.totalwg.Wait()
}

// Unlock a lock on the entire tree
func (T TreeLock) UnlockAll() {
	T.totalLock.Unlock()
}

// Equivalent to a TreeLock restricted to depth=1
func NewSimpleTreeLock() *SimpleTreeLock {
	return &SimpleTreeLock{new(sync.Mutex), new(sync.WaitGroup), make(map[string]*sync.Mutex)}
}

func (T SimpleTreeLock) Lock(val string) {
	T.totalLock.Lock()
	if _, ok := T.locks[val]; !ok {
		T.locks[val] = new(sync.Mutex)
	}
	T.totalwg.Add(1)
	T.totalLock.Unlock()
	T.locks[val].Lock()
}

func (T SimpleTreeLock) Unlock(val string) {
	T.totalwg.Done()
	T.locks[val].Unlock()
}

func (T SimpleTreeLock) LockMany(vals ...string) {
	sort.Strings(vals)
	T.totalLock.Lock()
	for _, val := range vals {
		if _, ok := T.locks[val]; !ok {
			T.locks[val] = new(sync.Mutex)
		}
	}
	T.totalwg.Add(len(vals))
	T.totalLock.Unlock()
	for _, val := range vals {
		T.locks[val].Lock()
	}
}

func (T SimpleTreeLock) UnlockMany(vals ...string) {
	sort.Strings(vals)
	for i := len(vals) - 1; i >= 0; i-- {
		T.locks[vals[i]].Unlock()
		T.totalwg.Done()
	}
}

func (T SimpleTreeLock) LockAll() {
	T.totalLock.Lock()
	T.totalwg.Wait()
}

func (T SimpleTreeLock) UnlockAll() {
	T.totalLock.Unlock()
}
