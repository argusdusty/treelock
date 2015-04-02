package treelock

import (
	"sort"
	"sync"
)

type TreeLock struct {
	totalLock *sync.Mutex
	totalwg   *sync.WaitGroup
	locks     map[string]*TreeLock
}

type SimpleTreeLock struct {
	totalLock *sync.Mutex
	totalwg   *sync.WaitGroup
	locks     map[string]*sync.Mutex
}

type Sorter [][]string

func (S Sorter) Len() int { return len(S) }
func (S Sorter) Less(i, j int) bool {
	s := len(S[i]) < len(S[j])
	m := len(S[j])
	if s {
		m = len(S[i])
	}
	for k := 0; k < m; k++ {
		if S[i][k] < S[j][k] {
			return true
		} else if S[i][k] > S[j][k] {
			return false
		}
	}
	return s
}
func (S Sorter) Swap(i, j int) { S[i], S[j] = S[j], S[i] }

// Create a new tree lock
func NewTreeLock() *TreeLock {
	return &TreeLock{new(sync.Mutex), new(sync.WaitGroup), make(map[string]*TreeLock)}
}

// Safely lock a value to prevent threads from accessing this value
func (T TreeLock) Lock(val []string) {
	if len(val) == 0 {
		T.LockAll()
		return
	}
	T.totalLock.Lock()
	if _, ok := T.locks[val[0]]; !ok {
		T.locks[val[0]] = NewTreeLock()
	}
	T.totalwg.Add(1)
	T.totalLock.Unlock()
	T.locks[val[0]].Lock(val[1:])
}

// Unlock a value to allow it to be used by another thread.
// Will panic if the value is not locked
func (T TreeLock) Unlock(val []string) {
	if len(val) == 0 {
		T.UnlockAll()
		return
	}
	T.totalwg.Done()
	T.locks[val[0]].Unlock(val[1:])
}

// Safely lock multiple values simultaneously while preventing race condition
// Use this if the same thread will need to have multiple values locked
// Attempting to lock overlapping values will deadlock
func (T TreeLock) LockMany(vals ...[]string) {
	sort.Sort(Sorter(vals))
	for _, val := range vals {
		T.Lock(val)
	}
}

// Safely unlock multiple values simultaneously while preventing race condition
func (T TreeLock) UnlockMany(vals ...[]string) {
	sort.Sort(Sorter(vals))
	for i := len(vals) - 1; i >= 0; i-- {
		T.Unlock(vals[i])
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
