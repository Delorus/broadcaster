package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type Value struct {
	Value1 string
	Value2 int
}

func TestSingleSub(t *testing.T) {
	broadcaster := NewBroadcaster[Value]()

	success := make(chan struct{})
	sub := broadcaster.Subscribe()
	go func() {
		var msgCount int
		for change := range sub {
			switch msgCount {
			case 0:
				assert.Empty(t, change.Value1)
				assert.Equal(t, 42, change.Value2)
				msgCount++
			case 1:
				assert.Equal(t, "hello", change.Value1)
				assert.Empty(t, change.Value2)
				success <- struct{}{}
				return
			}
		}
	}()

	broadcaster.Publish(Value{
		Value2: 42,
	})

	broadcaster.Publish(Value{
		Value1: "hello",
	})

	select {
	case <-time.After(200 * time.Millisecond):
		assert.Fail(t, "messages not delivered")
	case <-success:
	}
}

func TestMultiSub(t *testing.T) {
	broadcaster := NewBroadcaster[Value]()

	const subCount = 100
	successes := make([]chan struct{}, 0, subCount)
	for i := 0; i < subCount; i++ {
		success := make(chan struct{})
		sub := broadcaster.Subscribe()
		go func() {
			change := <-sub
			assert.Equal(t, 42, change.Value2)
			success <- struct{}{}
		}()
		successes = append(successes, success)
	}

	broadcaster.Publish(Value{
		Value2: 42,
	})

	for i := 0; i < subCount; i++ {
		select {
		case <-time.After(200 * time.Millisecond):
			assert.Fail(t, "message not delivered, subscriber: ", i)
		case <-successes[i]:
		}
	}
}
