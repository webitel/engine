package utils

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	b := NewDefaultBackoff()
	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Duration(), 200*time.Millisecond)
	equals(t, b.Duration(), 400*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func TestForAttempt(t *testing.T) {

	b := NewDefaultBackoff()
	equals(t, b.ForAttempt(0), 100*time.Millisecond)
	equals(t, b.ForAttempt(1), 200*time.Millisecond)
	equals(t, b.ForAttempt(2), 400*time.Millisecond)
	b.Reset()
	equals(t, b.ForAttempt(0), 100*time.Millisecond)
}

func Test2(t *testing.T) {

	b := NewBackoff(100*time.Millisecond, 10*time.Second, 1.5, false)
	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Duration(), 150*time.Millisecond)
	equals(t, b.Duration(), 225*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func Test3(t *testing.T) {

	b := NewBackoff(100*time.Nanosecond, 10*time.Second, 1.75, false)

	equals(t, b.Duration(), 100*time.Nanosecond)
	equals(t, b.Duration(), 175*time.Nanosecond)
	equals(t, b.Duration(), 306*time.Nanosecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Nanosecond)
}

func Test4(t *testing.T) {
	b := NewBackoff(500*time.Second, 100*time.Second, 1, false)
	equals(t, b.Duration(), b.max)
}

func TestGetAttempt(t *testing.T) {
	b := NewDefaultBackoff()
	equals(t, b.Attempt(), uint64(0))
	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Attempt(), uint64(1))
	equals(t, b.Duration(), 200*time.Millisecond)
	equals(t, b.Attempt(), uint64(2))
	equals(t, b.Duration(), 400*time.Millisecond)
	equals(t, b.Attempt(), uint64(3))
	b.Reset()
	equals(t, b.Attempt(), uint64(0))
	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Attempt(), uint64(1))
}

func TestJitter(t *testing.T) {
	b := NewDefaultBackoff()

	equals(t, b.Duration(), 100*time.Millisecond)
	between(t, b.Duration(), 100*time.Millisecond, 200*time.Millisecond)
	between(t, b.Duration(), 100*time.Millisecond, 400*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func TestConcurrent(t *testing.T) {
	b := NewDefaultBackoff()

	wg := &sync.WaitGroup{}

	test := func() {
		time.Sleep(b.Duration())
		wg.Done()
	}

	wg.Add(2)
	go test()
	go test()
	wg.Wait()
}

func between(t *testing.T, actual, low, high time.Duration) {
	t.Helper()
	if actual < low {
		t.Fatalf("Got %s, Expecting >= %s", actual, low)
	}
	if actual > high {
		t.Fatalf("Got %s, Expecting <= %s", actual, high)
	}
}

func equals(t *testing.T, v1, v2 interface{}) {
	t.Helper()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("Got %v, Expecting %v", v1, v2)
	}
}
