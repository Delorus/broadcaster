package main

import "sync"

// inspired by github.com/nats-io/go-nats

type Broadcaster[T any] struct {
	mu   sync.Mutex
	subs []*Subscription[T]
}

func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{}
}

func (b *Broadcaster[T]) Subscribe() <-chan T {
	sub := &Subscription[T]{
		mCond: sync.NewCond(&sync.Mutex{}),
	}

	b.mu.Lock()
	b.subs = append(b.subs, sub)
	b.mu.Unlock()

	ch := make(chan T)
	go func() {
		for {
			ch <- sub.waitNext().Payload
		}
	}()
	return ch
}

func (b *Broadcaster[T]) Publish(event T) {
	for _, sub := range b.subs {
		sub.mCond.L.Lock()
		if sub.mHead == nil {
			sub.mHead = &Event[T]{Payload: event}
		} else {
			sub.mHead.Next = &Event[T]{Payload: event}
		}
		sub.mCond.Signal()
		sub.mCond.L.Unlock()
	}
}

type Subscription[T any] struct {
	mCond *sync.Cond
	mHead *Event[T]
}

func (s *Subscription[T]) waitNext() *Event[T] {
	s.mCond.L.Lock()
	defer s.mCond.L.Unlock()

	for s.mHead == nil {
		s.mCond.Wait()
	}

	msg := s.mHead
	s.mHead = s.mHead.Next

	return msg
}

type Event[T any] struct {
	Payload T
	Next    *Event[T]
}
