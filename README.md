# TreeLock #

TreeLock is a tree-based locking mechanism in the Go Programming Language. It acts similar to sync.Mutex, but instead locks on a name-by-name basis, basically a map[string]sync.Mutex, with some additional features:

* Hierarchical locking: Using arrays of strings to lock sub-values. Locking []string{"a", "b"} will lock any string array that starts with []string{"a", "b"}, such as []string{"a", "b", "c"}. So, if you want to edit a single element of use "argusdusty", you could run `treelock.Lock([]string{"users", "argusdusty", "element"})`,  and another thread could still lock []string{"users", "argusdusty", "element2"} without waiting for the first to finish, but if another wanted to gain control over the whole object it could lock []string{"users, "argusdusty"}, which would wait for any existing threads using the "argusdusty" user to finish, then would block any future threads from using that user until unlocked.

* Locking multiple strings simultaneously: If a thread wanted to lock objects named "a" and "b", you could try locking "a", then "b", but another thread could lock "b", then "a" leading to a potential deadlock. treelock has LockMany, which will avoid a deadlock by sorting the incoming values beforehand, so now you just run `treelock.LockMany([]string{"a"}, []string{"b"})` or `treelock.LockMany([]string{"b"}, []string{"a"})` and, of course `treelock.UnlockMany([]string{"a"}, []string{"b"})` and you don't have to worry about potential deadlock conditions.

This allows you to minimize overhead in asynchronous access to a large number of distinct objects, without fear of two threads undoing each other.


## Examples ##
* Lock all users: `treelock.Lock([]string{"users"})`
* Lock a single user named "argusdusty": `treelock.Lock([]string{"users", "argusdusty"})`
* Lock two users named "1234" and "asdf": `treelock.LockMany([]string{"users", "1234"}, []string{"users", "asdf"})`
* Lock all users: `treelock.LockMany([]string{"users"})`
* Lock everything: `treelock.LockAll()`


### SimpleTreeLock ###

The SimpleTreeLock structure is the same as TreeLock, except limited to depth=1, so it's more of a MapLock:
* Lock a single item: `simpletreelock.Lock("asdf")`
* Lock multiple items: `simpletreelock.LockMany("asdf", "1234")`
* Lock everything: `simpletreelock.LockAll()`


### Attribution ###

This code was developed for Tamber, Inc. (www.tamber.com), and is used in the backend for the Tamber Concerts app to optimize data manipulation while ensuring thread safety over arbitrarily defined collections of data.