# TreeLock #

TreeLock is a tree-based locking mechanism in the Go Programming Language. It acts similar to sync.Mutex, but instead locks on a name-by-name basis, basically a map[string]sync.Mutex, with some additional features:

* Hierarchical locking. Using the separator rune '.', Locking "a" will lock any string that starts with "a.", and so on. So, if you want to edit a single element of use "argusdusty", you could lock "user.argusdusty.element", thus another thread could lock "user.argusdusty.field" while this is going on, but if another wanted to gain control over the whole object (i.e. for saving into the database) it could lock "user.argusdusty", which would wait for existing "user.argusdusty." threads to finish, then cause any new ones to wait.

* Locking multiple strings simultaneously. If a thread wanted to lock objects named "a" and "b", you could try `treelock.Lock("a"); treelock.Lock("b")` but another thread could run `treelock.Lock("b"); treelock.Lock("a")` leading to a potential deadlock. treelock has LockMany, which will avoid a deadlock by sorting the incoming strings beforehand, so now you just run `treelock.LockMany("a", "b")` or `treelock.LockMany("b", "a")` and, of course `treelock.UnlockMany("a", "b")` and you don't have to worry about potential deadlock conditions.


### SimpleTreeLock ###

The SimpleTreeLock structure is the initial iteration of this code. It has every feature as TreeLock except for hierarchical locking, so locking "a" will not lock "a.b". Thus, no separator rune is required to initialize a SimpleTreeLock


### Attribution ###

This code was developed for Tamber, Inc. (www.tamber.com), and is used in the backend for the Tamber Concerts app to optimize data manipulation while ensuring thread safety over arbitrarily defined collections of data.