package semaphore

import (
	"errors"
	"time"
)

var (
	ErrNoTickets      = errors.New("don't acquire semaphore")
	ErrIllegalRelease = errors.New("don't release semaphore")
)

// Interface содержит поведение семафора,
// который может быть захвачен (Acquire)
// и/или освобожден (Release).
type Interface interface {
	Acquire() error
	Release() error
}

type Semaphore struct {
	sem     chan struct{}
	timeout time.Duration
}

type SemMap struct {
	mp map[string]Semaphore
}

var SemMapInstance = make(map[string]Semaphore)

func GetSemMap() *SemMap {
	return &SemMap{
		mp: SemMapInstance,
	}
}
func Acquire(sn string) error {
	if semaphore, ok := SemMapInstance[sn]; !ok {
		SemMapInstance[sn] = Semaphore{
			sem:     make(chan struct{}, 1),
			timeout: 10 * time.Second,
		}
		semaphore = SemMapInstance[sn]
		return semaphore.Acquire()
	} else {
		return semaphore.Acquire()
	}
}
func Release(sn string) error {
	if semaphore, ok := SemMapInstance[sn]; ok {
		return semaphore.Release()
	}
	return ErrIllegalRelease
}

func (s *Semaphore) Acquire() error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-time.After(s.timeout):
		return ErrNoTickets
	}
}

func (s *Semaphore) Release() error {
	select {
	case <-s.sem:
		return nil
	case <-time.After(s.timeout):
		return ErrIllegalRelease
	}
}

func New(tickets int, timeout time.Duration, sn string) {
	SemMapInstance[sn] = Semaphore{
		sem:     make(chan struct{}, tickets),
		timeout: timeout,
	}
}
