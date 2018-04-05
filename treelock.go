package treelock

import (
	"sort"
	"sync"
)

type WaitGroup interface {
	Add(delta int)
	Wait()
	Done()
}

type TreeLock struct {
	totalLock sync.Locker
	totalwg   WaitGroup
	locks     map[string]*TreeLock
	locksLock sync.Locker
	val       string
	parent    *TreeLock
	depth     int
}

type SimpleTreeLock struct {
	totalLock sync.Locker
	totalwg   WaitGroup
	locks     map[string]sync.Locker
	locksLock sync.Locker
}

// Global function supporting custom Mutex/WaitGroup generators.
// For example, allowing global treeLocks by handling locks through a database.
// Or alternatively, enabling callbacks when locking/unlocking occurs
var MutexGenerator func(path []string) sync.Locker = func(path []string) sync.Locker { return new(sync.Mutex) }
var WaitGroupGenerator func(path []string) WaitGroup = func(path []string) WaitGroup { return new(sync.WaitGroup) }

type Sorter [][]string

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

// Create a new tree lock
func NewTreeLock() *TreeLock {
	return &TreeLock{MutexGenerator([]string{}), WaitGroupGenerator([]string{}), make(map[string]*TreeLock), new(sync.Mutex), "", nil, 0}
}

func newTreeLock(val string, parent *TreeLock, depth int) *TreeLock {
	path := make([]string, depth)
	i := depth - 1
	pt := parent
	for pt.parent != nil && i >= 0 {
		path[i] = pt.val
		i--
		pt = pt.parent
	}
	return &TreeLock{MutexGenerator(path), WaitGroupGenerator(path), make(map[string]*TreeLock), new(sync.Mutex), val, nil, depth}
}

// Safely lock a value to prevent threads from accessing this value
func (T *TreeLock) Lock(val []string) {
	if len(val) == 0 {
		T.LockAll()
		return
	}
	T.totalLock.Lock()
	T.locksLock.Lock()
	if _, ok := T.locks[val[0]]; !ok {
		T.locks[val[0]] = newTreeLock(val[0], T, T.depth+1)
	}
	lock := T.locks[val[0]]
	T.locksLock.Unlock()
	T.totalwg.Add(1)
	T.totalLock.Unlock()
	lock.Lock(val[1:])
}

// Unlock a value to allow it to be used by another thread.
// Will panic if the value is not locked
func (T *TreeLock) Unlock(val []string) {
	if len(val) == 0 {
		T.UnlockAll()
		return
	}
	T.totalwg.Done()
	T.locksLock.Lock()
	lock := T.locks[val[0]]
	T.locksLock.Unlock()
	lock.Unlock(val[1:])
}

// Safely lock multiple values simultaneously while preventing race condition
// Use this if the same thread will need to have multiple values locked
// Attempting to lock overlapping values will deadlock
func (T *TreeLock) LockMany(vals ...[]string) {
	sort.SliceStable(vals, Sorter(vals).Less)
	for _, val := range vals {
		T.Lock(val)
	}
}

// Safely unlock multiple values simultaneously while preventing race condition
func (T *TreeLock) UnlockMany(vals ...[]string) {
	sort.SliceStable(vals, Sorter(vals).Less)
	for i := len(vals) - 1; i >= 0; i-- {
		T.Unlock(vals[i])
	}
}

// Lock the entire tree, waits for all existing locks to unlock first.
func (T *TreeLock) LockAll() {
	T.totalLock.Lock()
	T.totalwg.Wait()
}

// Unlock a lock on the entire tree
func (T *TreeLock) UnlockAll() {
	T.totalLock.Unlock()
}

// Equivalent to a TreeLock restricted to depth=1
func NewSimpleTreeLock() *SimpleTreeLock {
	return &SimpleTreeLock{MutexGenerator([]string{}), WaitGroupGenerator([]string{}), make(map[string]sync.Locker), new(sync.Mutex)}
}

func (T *SimpleTreeLock) Lock(val string) {
	T.totalLock.Lock()
	if _, ok := T.locks[val]; !ok {
		T.locks[val] = MutexGenerator([]string{val})
	}
	T.totalwg.Add(1)
	lock := T.locks[val]
	T.totalLock.Unlock()
	lock.Lock()
}

func (T *SimpleTreeLock) Unlock(val string) {
	T.totalwg.Done()
	T.totalLock.Lock()
	lock := T.locks[val]
	T.totalLock.Unlock()
	lock.Unlock()
}

func (T *SimpleTreeLock) LockMany(vals ...string) {
	sort.Strings(vals)
	T.totalLock.Lock()
	for _, val := range vals {
		if _, ok := T.locks[val]; !ok {
			T.locks[val] = MutexGenerator([]string{val})
		}
	}
	T.totalwg.Add(len(vals))
	locks := make([]sync.Locker, len(vals))
	for i, val := range vals {
		locks[i] = T.locks[val]
	}
	T.totalLock.Unlock()
	for _, lock := range locks {
		lock.Lock()
	}
}

func (T *SimpleTreeLock) UnlockMany(vals ...string) {
	sort.Strings(vals)
	T.totalLock.Lock()
	for i := len(vals) - 1; i >= 0; i-- {
		T.locks[vals[i]].Unlock()
		T.totalwg.Done()
	}
	locks := make([]sync.Locker, len(vals))
	for i, val := range vals {
		locks[i] = T.locks[val]
	}
	T.totalLock.Unlock()
	for _, lock := range locks {
		lock.Unlock()
	}
}

func (T *SimpleTreeLock) LockAll() {
	T.totalLock.Lock()
	T.totalwg.Wait()
}

func (T SimpleTreeLock) UnlockAll() {
	T.totalLock.Unlock()
}
