package cronzilla

import (
	gerrors "github.com/go-errors/errors"

	"context"
	"errors"
	"fmt"
	"time"
)

// ErrTaskPanicError is an error returned if a Task panics during Run
type ErrTaskPanicError struct {
	message string
}

// Error returns the string message
func (e ErrTaskPanicError) Error() string {
	return e.message
}

// TaskFunc is a func that has no parameters and returns only error
type TaskFunc func() error

// ErrorlessTaskFunc wraps a func() into a TaskFunc
// TODO: Fix Name
func ErrorlessTaskFunc(f func()) TaskFunc {
	return func() error {
		f()
		return nil
	}
}

// Task is our... task. Philosophically, Todo() is run in the goro executing Run(), so in general you should
// give it it's own. This is done because I believe that if your Task.Run() panics, even though we recover and gracefully
// handle it, that should be the end of your Task unless you call Run() again. Running Todo() in a separate goro would
// allow the Task to keep on executing a panic'y Todo(), but that's just not right.
// If that's important to you, you can write a re-Run()ing wrangler that handles panic cases for you.
type Task struct {
	// Todo is a TaskFunc that gets called Every
	Todo TaskFunc
	// Every is a Duration for how often Todo()
	Every time.Duration
	// At is a Time to run Todo()
	At time.Time
	// done is tracks if Run has exited
	done bool
}

// Run takes a context and error chan.
// If Every is non-zero, calls Todo() Every until the context expires or is cancelled.
// If At is non-zero, calls Todo() At.
// If both Every and At are defined, will do both.
// The error chan WILL BE CLOSED when the function exits, which is a good way to know that the task isn't running
// anymore, otherwise will only pass non-nil errors from Todo(). If the error is an ErrTaskPanicError, then Todo() panic'd,
// and the stack trace is returned on errorChan just before it is closed.
func (t *Task) Run(ctx context.Context, errorChan chan error) {
	t.run(ctx, errorChan, false)
}

// RunOnce takes an error chan, and calls Todo() once after an Every or At.
// The error chan WILL BE CLOSED when the function exits, which is a good way to know that the task isn't running
// anymore, otherwise will only pass non-nil errors from Todo(). If the error is an ErrTaskPanicError, then Todo() panic'd,
// and the stack trace is returned on errorChan just before it is closed.
func (t *Task) RunOnce(errorChan chan error) {
	t.run(context.Background(), errorChan, true)
}

// run actually does the run
func (t *Task) run(ctx context.Context, errorChan chan error, once bool) {
	defer func() {
		// When we're done, set done and close the errorChan
		t.done = true
		close(errorChan)
	}()
	defer func() {
		if r := recover(); r != nil {
			var err error
			switch rt := r.(type) {
			case string:
				err = errors.New(rt)
			case error:
				err = rt
			default:
				err = errors.New(fmt.Sprintf("Unknown error: '%+v'", r))
			}

			// Don't block if there isn't a reader
			select {
			case errorChan <- ErrTaskPanicError{gerrors.Wrap(err, 2).ErrorStack()}:
			default:
			}
		}
	}()
	t.done = false

	if t.At.Unix() > 0 {
		atChan := make(chan struct{})
		atTimer := time.AfterFunc(time.Until(t.At), func() {
			defer func() { close(atChan) }()
			if e := t.Todo(); e != nil {
				// Don't block if there isn't a reader
				select {
				case errorChan <- e:
				default:
				}
			}
		})

		if t.Every == 0 {
			// We won't have a ticker keeping us alive
			select {
			case <-ctx.Done():
				return
			case <-atChan:
				// Win.
				return
			}
		}
		defer atTimer.Stop()
	}

	if t.Every > 0 {
		ticker := time.NewTicker(t.Every)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if e := t.Todo(); e != nil {
					// Don't block if there isn't a reader
					select {
					case errorChan <- e:
					default:
					}
				}
			}

			if once {
				return
			}
		}
	}
}

// IsDone returns an internal state bool that is set if Run() was called, but has exited because of crash, completion, or cancellation.
func (t *Task) IsDone() bool {
	return t.done
}
