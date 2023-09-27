package rabbitmq

import (
	"fmt"
	"time"
	"sync"
	"sync/atomic"
	"strings"
)

const DEFAULT_TTL = time.Second * 10
const DEFAULT_INTERVAL = time.Second * 4

func NewConflictResolver() *resolver {
	r := &resolver{
		locks: make(map[string][]item),
	}
	r.CheckLocks(DEFAULT_INTERVAL, DEFAULT_TTL, nil, nil)
	return r
}

type ConflictResolver interface {
	CheckLocks(interval time.Duration, ttl time.Duration, errfn func(string), infofn func(string))
	Enter(key string, priority uint64) (id uint64, conflict bool)
	Leave(key string, id uint64) (e error)
}

type resolver struct {
	idseq idseq
	mutex sync.Mutex
	locks map[string][]item

	ticker *time.Ticker
}

type item struct {
	id         uint64
	priority   uint64
	processing chan uint64
	conflict   bool
	created    time.Time
}

type idseq struct {
        mutex     sync.Mutex
        currval   uint64
}

func (seq *idseq) nextval() uint64 {
	atomic.AddUint64(&seq.currval, 1)
        return seq.currval
}

func (this *resolver) CheckLocks(
	interval time.Duration,
	ttl time.Duration,
	errfn func(string),
	infofn func(string),
) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if t := this.ticker; t != nil {
		t.Stop()
	}

	this.ticker = time.NewTicker(interval)
	go (func(t *time.Ticker){
		for _ = range t.C {
			this.resolveExpires(ttl, errfn, infofn)
		}
	})(this.ticker)
}

func (this *resolver) Enter(key string, priority uint64) (uint64, bool) {
	l := this.lockEntity(
		key,
		priority,
		this.idseq.nextval(),
	)

	// waiting for resolve
	this.wait(l)

	return l.id, l.conflict
}

func (this *resolver) Leave(key string, id uint64) error {
	return this.unlockEntity(key, id)
}

func (this *resolver) wait(i item) {
	<-i.processing
}

func (this *resolver) resolve(i item) {
	i.processing <- i.id
}

func (this *resolver) lockEntity(key string, priority uint64, id uint64) item {
	// key = key[:7]

	this.mutex.Lock()
	defer this.mutex.Unlock()

	lock := item{
		id: id,
		priority: priority,
		processing: make(chan uint64),
		created: time.Now(),
	}
	if l := len(this.locks[key]); l > 0 {
		if (this.locks[key][l-1].priority >= lock.priority) {
			lock.conflict = true
		}
	} else {
		// no locks for key
		// can start processing
		go this.resolve(lock)
	}
	this.locks[key] = append(this.locks[key], lock)
	return lock
}

func (this *resolver) unlockEntity(key string, id uint64) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	locks := this.locks[key]
	if locks == nil {
		return ErrorLockForKeyNotFoundError
	}
	// TODO: investigate for multiple positions
	// no need to investigate multiple positions case
	// currently all generated ids are unique
	pos := -1
	for i, l := range locks {
		if l.id == id {
			pos = i
		}
	}
	if pos < 0 {
		return ErrorLockForIdNotFoundError
	}
	if len(locks) > 1 {
		// for multiple pos: if len(locks) > len(positions)
		// for multiple pos: newlocks := make([]item, len(locks)-len(positions))
		newlocks := make([]item, len(locks)-1)
		copy(newlocks[:pos], locks[:pos])
		copy(newlocks[pos:], locks[pos+1:])
		this.locks[key] = newlocks
		if pos == 0 {
			go this.resolve(newlocks[0])
		}
		return nil
	}
	delete(this.locks, key)
	return nil
}

func (this *resolver) resolveExpires(ttl time.Duration, errfn func(string), infofn func(string)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	infos := []string{}

	now := time.Now()
	maxcreated := now.Add(-1 * ttl)
	for key, locks := range this.locks {
		if infofn != nil {
			infos = append(infos, fmt.Sprintf("%s in queue %d", key, len(locks)))
		}
		if l := locks[0]; l.created.Before(maxcreated) {
			// resolve only first(hanged) item
			// others locks left in queue may be resolved correctly
			if errfn != nil {
				errmsg := fmt.Sprintf(
					"%s ttl(%s) expired, created %s ago",
                                        key,
                                        ttl.String(),
                                        l.created.Sub(now).String(),
                                )
				go errfn(errmsg)
			}
			if len(locks) > 1 {
				newlocks := make([]item, len(locks)-1)
				copy(newlocks, locks[1:])
				this.locks[key] = newlocks
				go this.resolve(newlocks[0])
				continue
			}
			delete(this.locks, key)
		}
	}

	if infofn != nil {
		infomsg := fmt.Sprintf(
			"processing: [%s]",
			strings.Join(infos, ","),
		)
		go infofn(infomsg)
	}
}
