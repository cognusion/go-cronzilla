# go-cronzilla

[![GoDoc](https://godoc.org/github.com/cognusion/go-cronzilla?status.svg)](https://godoc.org/github.com/cognusion/go-cronzilla)


# cronzilla
`import "github.com/cognusion/go-cronzilla"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [type ErrTaskPanicError](#ErrTaskPanicError)
  * [func (e ErrTaskPanicError) Error() string](#ErrTaskPanicError.Error)
* [type Task](#Task)
  * [func (t *Task) IsDone() bool](#Task.IsDone)
  * [func (t *Task) Run(ctx context.Context, errorChan chan error)](#Task.Run)
  * [func (t *Task) RunOnce(errorChan chan error)](#Task.RunOnce)
* [type TaskFunc](#TaskFunc)
  * [func ErrorlessTaskFunc(f func()) TaskFunc](#ErrorlessTaskFunc)
* [type Wrangler](#Wrangler)
  * [func (w *Wrangler) AddAt(name string, todo TaskFunc, at time.Time) &lt;-chan error](#Wrangler.AddAt)
  * [func (w *Wrangler) AddEvery(name string, todo TaskFunc, every time.Duration) &lt;-chan error](#Wrangler.AddEvery)
  * [func (w *Wrangler) Clean() int](#Wrangler.Clean)
  * [func (w *Wrangler) Close()](#Wrangler.Close)
  * [func (w *Wrangler) Count() int](#Wrangler.Count)
  * [func (w *Wrangler) CountStale() int](#Wrangler.CountStale)
  * [func (w *Wrangler) Delete(name string)](#Wrangler.Delete)
  * [func (w *Wrangler) Exists(name string) bool](#Wrangler.Exists)
  * [func (w *Wrangler) List() []string](#Wrangler.List)
  * [func (w *Wrangler) ListStale() []string](#Wrangler.ListStale)


#### <a name="pkg-files">Package files</a>
[task.go](/src/github.com/cognusion/go-cronzilla/task.go) [wrangler.go](/src/github.com/cognusion/go-cronzilla/wrangler.go) 






## <a name="ErrTaskPanicError">type</a> [ErrTaskPanicError](/src/target/task.go?s=192:241#L14)
``` go
type ErrTaskPanicError struct {
    // contains filtered or unexported fields
}

```
ErrTaskPanicError is an error returned if a Task panics during Run










### <a name="ErrTaskPanicError.Error">func</a> (ErrTaskPanicError) [Error](/src/target/task.go?s=279:320#L19)
``` go
func (e ErrTaskPanicError) Error() string
```
Error returns the string message




## <a name="Task">type</a> [Task](/src/target/task.go?s=1138:1386#L40)
``` go
type Task struct {
    // Todo is a TaskFunc that gets called Every
    Todo TaskFunc
    // Every is a Duration for how often Todo()
    Every time.Duration
    // At is a Time to run Todo()
    At time.Time
    // contains filtered or unexported fields
}

```
Task is our... task. Philosophically, Todo() is run in the goro executing Run(), so in general you should
give it it's own. This is done because I believe that if your Task.Run() panics, even though we recover and gracefully
handle it, that should be the end of your Task unless you call Run() again. Running Todo() in a separate goro would
allow the Task to keep on executing a panic'y Todo(), but that's just not right.
If that's important to you, you can write a re-Run()ing wrangler that handles panic cases for you.










### <a name="Task.IsDone">func</a> (\*Task) [IsDone](/src/target/task.go?s=4084:4112#L149)
``` go
func (t *Task) IsDone() bool
```
IsDone returns an internal state bool that is set if Run() was called, but has exited because of crash, completion, or cancellation.




### <a name="Task.Run">func</a> (\*Task) [Run](/src/target/task.go?s=1916:1977#L58)
``` go
func (t *Task) Run(ctx context.Context, errorChan chan error)
```
Run takes a context and error chan.
If Every is non-zero, calls Todo() Every until the context expires or is cancelled.
If At is non-zero, calls Todo() At.
If both Every and At are defined, will do both.
The error chan WILL BE CLOSED when the function exits, which is a good way to know that the task isn't running
anymore, otherwise will only pass non-nil errors from Todo(). If the error is an ErrTaskPanicError, then Todo() panic'd,
and the stack trace is returned on errorChan just before it is closed.




### <a name="Task.RunOnce">func</a> (\*Task) [RunOnce](/src/target/task.go?s=2401:2445#L66)
``` go
func (t *Task) RunOnce(errorChan chan error)
```
RunOnce takes an error chan, and calls Todo() once after an Every or At.
The error chan WILL BE CLOSED when the function exits, which is a good way to know that the task isn't running
anymore, otherwise will only pass non-nil errors from Todo(). If the error is an ErrTaskPanicError, then Todo() panic'd,
and the stack trace is returned on errorChan just before it is closed.




## <a name="TaskFunc">type</a> [TaskFunc](/src/target/task.go?s=412:438#L24)
``` go
type TaskFunc func() error
```
TaskFunc is a func that has no parameters and returns only error







### <a name="ErrorlessTaskFunc">func</a> [ErrorlessTaskFunc](/src/target/task.go?s=510:551#L28)
``` go
func ErrorlessTaskFunc(f func()) TaskFunc
```
ErrorlessTaskFunc wraps a func() into a TaskFunc
TODO: Fix Name





## <a name="Wrangler">type</a> [Wrangler](/src/target/wrangler.go?s=106:146#L10)
``` go
type Wrangler struct {
    // contains filtered or unexported fields
}

```
Wrangler is a goro-safe aggregator for Tasks










### <a name="Wrangler.AddAt">func</a> (\*Wrangler) [AddAt](/src/target/wrangler.go?s=2597:2676#L112)
``` go
func (w *Wrangler) AddAt(name string, todo TaskFunc, at time.Time) <-chan error
```
AddAt will include the named task to run specifically at the specified time, once, returning an error channel to listen on




### <a name="Wrangler.AddEvery">func</a> (\*Wrangler) [AddEvery](/src/target/wrangler.go?s=2274:2363#L105)
``` go
func (w *Wrangler) AddEvery(name string, todo TaskFunc, every time.Duration) <-chan error
```
AddEvery will include the named task to run every so often, returning an error channel to listen on




### <a name="Wrangler.Clean">func</a> (\*Wrangler) [Clean](/src/target/wrangler.go?s=1139:1169#L53)
``` go
func (w *Wrangler) Clean() int
```
Clean will remove completed or crashed tasks




### <a name="Wrangler.Close">func</a> (\*Wrangler) [Close](/src/target/wrangler.go?s=930:956#L44)
``` go
func (w *Wrangler) Close()
```
Close will cancel all of the tasks being wrangled. The Wrangler may be reused after Close is called




### <a name="Wrangler.Count">func</a> (\*Wrangler) [Count](/src/target/wrangler.go?s=1409:1439#L67)
``` go
func (w *Wrangler) Count() int
```
Count returns the current number of tasks being wrangled




### <a name="Wrangler.CountStale">func</a> (\*Wrangler) [CountStale](/src/target/wrangler.go?s=1629:1664#L77)
``` go
func (w *Wrangler) CountStale() int
```
CountStale returns the current number of wrangled tasks that have completed or crashed




### <a name="Wrangler.Delete">func</a> (\*Wrangler) [Delete](/src/target/wrangler.go?s=1886:1924#L90)
``` go
func (w *Wrangler) Delete(name string)
```
Delete will cancel and remove the named task from the Wrangler




### <a name="Wrangler.Exists">func</a> (\*Wrangler) [Exists](/src/target/wrangler.go?s=2082:2125#L99)
``` go
func (w *Wrangler) Exists(name string) bool
```
Exists returns bool if the specified Task exists




### <a name="Wrangler.List">func</a> (\*Wrangler) [List](/src/target/wrangler.go?s=299:333#L21)
``` go
func (w *Wrangler) List() []string
```
List will return a string array of Task names




### <a name="Wrangler.ListStale">func</a> (\*Wrangler) [ListStale](/src/target/wrangler.go?s=575:614#L31)
``` go
func (w *Wrangler) ListStale() []string
```
ListStale will return a string array of Task names that have completed or crashed








- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
