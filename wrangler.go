package cronzilla

import (
	"context"
	"sync"
	"time"
)

// Wrangler is a goro-safe aggregator for Tasks
type Wrangler struct {
	tasks sync.Map
}

type wrangledTask struct {
	task       *Task
	errorChan  chan error
	cancelfunc context.CancelFunc
}

// List will return a string array of Task names
func (w *Wrangler) List() []string {
	tasks := make([]string, 0)
	w.tasks.Range(func(name, wtask interface{}) bool {
		tasks = append(tasks, name.(string))
		return true
	})
	return tasks
}

// ListStale will return a string array of Task names that have completed or crashed
func (w *Wrangler) ListStale() []string {
	tasks := make([]string, 0)
	w.tasks.Range(func(name, wtask interface{}) bool {
		t := wtask.(wrangledTask).task
		if t.IsCrashed() {
			tasks = append(tasks, name.(string))
		}
		return true
	})
	return tasks
}

// Close will cancel all of the tasks being wrangled. The Wrangler may be reused after Close is called
func (w *Wrangler) Close() {
	w.tasks.Range(func(name, wtask interface{}) bool {
		w.tasks.Delete(name)
		wtask.(wrangledTask).cancelfunc()
		return true
	})
}

// Clean will remove completed or crashed tasks
func (w *Wrangler) Clean() int {
	c := 0
	w.tasks.Range(func(name, wtask interface{}) bool {
		t := wtask.(wrangledTask).task
		if t.IsCrashed() {
			c++
			w.tasks.Delete(name)
		}
		return true
	})
	return c
}

// Count returns the current number of tasks being wrangled
func (w *Wrangler) Count() int {
	c := 0
	w.tasks.Range(func(name, wtask interface{}) bool {
		c++
		return true
	})
	return c
}

// CountStale returns the current number of wrangled tasks that have completed or crashed
func (w *Wrangler) CountStale() int {
	c := 0
	w.tasks.Range(func(name, wtask interface{}) bool {
		t := wtask.(wrangledTask).task
		if t.IsCrashed() {
			c++
		}
		return true
	})
	return c
}

// Delete will cancel and remove the named task from the Wrangler
func (w *Wrangler) Delete(name string) {
	old, ok := w.tasks.Load(name)
	if ok {
		old.(wrangledTask).cancelfunc()
		w.tasks.Delete(name)
	}
}

// AddEvery will include the named task to run every so often, returning an error channel to listen on
func (w *Wrangler) AddEvery(name string, todo TaskFunc, every time.Duration) <-chan error {
	errorChan := make(chan error, 1)
	w.add(name, todo, every, time.Time{}, errorChan)
	return errorChan
}

// AddAt will include the named task to run specifically at the specified time, once, returning an error channel to listen on
func (w *Wrangler) AddAt(name string, todo TaskFunc, at time.Time) <-chan error {
	errorChan := make(chan error, 1)
	w.add(name, todo, 0, at, errorChan)
	return errorChan
}

// add does the heavy lifting of creating a Task, wrangledTask, dealing with dupes, and running
func (w *Wrangler) add(name string, todo TaskFunc, every time.Duration, at time.Time, errorChan chan error) {
	task := Task{
		Todo:  todo,
		Every: every,
		At:    at,
	}
	ctx, cancelfunc := context.WithCancel(context.Background())

	wt := wrangledTask{
		task:       &task,
		errorChan:  errorChan,
		cancelfunc: cancelfunc,
	}

	old, loaded := w.tasks.LoadOrStore(name, wt)

	if loaded {
		old.(wrangledTask).cancelfunc()
		w.tasks.Delete(name)
		w.tasks.Store(name, wt)
	}

	go task.Run(ctx, errorChan)

}
