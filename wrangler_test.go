package cronzilla

import (
	"testing"
	"time"
)

func Test_Wrangler(t *testing.T) {

	pingChan := make(chan int, 10)

	w := Wrangler{}
	defer w.Close()

	tFunc := func() error {
		pingChan <- 1
		return nil
	}

	errorChan := w.AddEvery("test", tFunc, time.Millisecond)
	go func() {
		<-time.After(5 * time.Millisecond)
		w.Delete("test")
	}()

	counter := 0
OUT:
	for {
		select {
		case p := <-pingChan:
			counter += p
		case e := <-errorChan:
			if e != nil {
				t.Errorf("Unexpected error %s\n", e)
			} else {
				break OUT
			}
		}
	}

	if counter < 4 || counter > 6 {
		t.Errorf("Expected 4..6, got %d\n", counter)
	}
}

func Test_WranglerAddTwice(t *testing.T) {

	pingChan := make(chan int, 10)

	w := Wrangler{}
	defer w.Close()

	tFunc := func() error {
		pingChan <- 10
		return nil
	}

	tFunc2 := func() error {
		pingChan <- 1
		return nil
	}

	errorChan := w.AddEvery("test", tFunc, time.Millisecond)
	errorChan2 := w.AddEvery("test", tFunc2, time.Millisecond)

	go func() {
		<-time.After(5 * time.Millisecond)
		w.Delete("test")
	}()

	counter := 0
OUT:
	for {
		select {
		case p := <-pingChan:
			counter += p
		case e := <-errorChan:
			if e != nil {
				t.Errorf("Unexpected error %s\n", e)
			}
		case e := <-errorChan2:
			if e != nil {
				t.Errorf("Unexpected error2 %s\n", e)
			} else {
				break OUT
			}
		}
	}

	if counter < 4 || counter > 6 {
		t.Errorf("Expected 4..6, got %d\n", counter)
	}
}

func Test_WranglerCountCleanClose(t *testing.T) {

	exitTime := time.Now().Add(3 * time.Millisecond)

	w := Wrangler{}
	defer w.Close()

	tFunc := func() error {
		if time.Now().After(exitTime) {
			panic("oh no!")
		}
		return nil
	}

	errorChan := w.AddEvery("test", tFunc, time.Millisecond)

	go func() {
		for {
			select {
			case e := <-errorChan:
				if e == nil {
					return
				}
			}
		}
	}()

	// Should have 1 item
	if c := w.Count(); c != 1 {
		t.Errorf("Count is %d, expected 1!\n", c)
	}

	// That 1 item should not be stale
	if c := w.CountStale(); c != 0 {
		t.Errorf("CountStale is %d, expected 0!\n", c)
	}

	// Wait until after our panic
	time.Sleep(5 * time.Millisecond)

	// Should still have 1 item
	if c := w.Count(); c != 1 {
		t.Errorf("Count is %d, expected 1!\n", c)
	}

	// That 1 item should _now_ be stale
	if c := w.CountStale(); c != 1 {
		t.Errorf("CountStale is %d, expected 1!\n", c)
	}

	// When we call Clean, it should clean 1 item
	if c := w.Clean(); c != 1 {
		t.Errorf("Clean should have been 1, was %d!\n", c)
	}

	// Should now have 0 items
	if c := w.Count(); c != 0 {
		t.Errorf("Count is %d, expected 0!\n", c)
	}

	// And 0 should be stale
	if c := w.CountStale(); c != 0 {
		t.Errorf("CountStale is %d, expected 0!\n", c)
	}

}
