package cronzilla

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_TaskRunEvery(t *testing.T) {

	pingChan := make(chan int, 10)
	errorChan := make(chan error, 1)
	tFunc := func() error {
		pingChan <- 1
		return nil
	}
	task := Task{
		Every: time.Millisecond,
		Todo:  tFunc,
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelfunc()
	go task.Run(ctx, errorChan)

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

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter != 5 {
		t.Errorf("Expected 5, got %d\n", counter)
	}
}

func Test_TaskRunAtEvery(t *testing.T) {

	pingChan := make(chan int, 10)
	errorChan := make(chan error, 1)
	tFunc := func() error {
		pingChan <- 1
		return nil
	}
	task := Task{
		Every: time.Millisecond,
		At:    time.Now().Add(3 * time.Millisecond),
		Todo:  tFunc,
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelfunc()
	go task.Run(ctx, errorChan)

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

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter != 6 {
		t.Errorf("Expected 6, got %d\n", counter)
	}
}

func Test_TaskRunAt(t *testing.T) {

	pingChan := make(chan int, 10)
	errorChan := make(chan error, 1)
	tFunc := func() error {
		pingChan <- 1
		return nil
	}
	task := Task{
		Todo: tFunc,
		At:   time.Now().Add(3 * time.Millisecond),
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelfunc()
	go task.Run(ctx, errorChan)

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

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter != 1 {
		t.Errorf("Expected 1, got %d\n", counter)
	}
}

func Test_TaskRunAtError(t *testing.T) {

	errorChan := make(chan error, 1)
	tFunc := func() error {
		return fmt.Errorf("Error")
	}
	task := Task{
		Todo: tFunc,
		At:   time.Now().Add(3 * time.Millisecond),
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelfunc()
	go task.Run(ctx, errorChan)

	counter := 0
OUT:
	for {
		select {
		case e := <-errorChan:
			if e != nil {
				counter++
			} else {
				break OUT
			}
		}
	}

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter != 1 {
		t.Errorf("Expected 1, got %d\n", counter)
	}
}

func Test_TaskError(t *testing.T) {

	errorChan := make(chan error, 1)
	tFunc := func() error {
		return fmt.Errorf("Error!")
	}
	task := Task{
		Every: time.Millisecond,
		Todo:  tFunc,
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelfunc()
	go task.Run(ctx, errorChan)

	counter := 0
OUT:
	for {
		select {
		case e := <-errorChan:
			if e != nil {
				counter++
			} else {
				break OUT
			}
		}
	}

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter < 4 || counter > 6 {
		t.Errorf("Expected error count 4..6, got %d\n", counter)
	}
}

func Test_TaskRunPanic(t *testing.T) {

	pingChan := make(chan int, 10)
	errorChan := make(chan error, 1)
	notTime := time.Now().Add(3 * time.Millisecond)

	tFunc := func() error {
		if time.Now().Before(notTime) {
			panic("OMG")
		}
		pingChan <- 1
		return nil
	}
	task := Task{
		Every: time.Millisecond,
		Todo:  tFunc,
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelfunc()
	go task.Run(ctx, errorChan)

	counter := 0
OUT:
	for {
		select {
		case p := <-pingChan:
			counter += p
		case e := <-errorChan:
			if e != nil {
				switch et := e.(type) {
				case ErrTaskPanicError:
					// Win
					_ = et.Error()
					continue
				default:
					t.Errorf("Unexpected error %s %s\n", e, et)
				}
			} else {
				break OUT
			}
		}
	}

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter > 0 {
		t.Errorf("Expected 0, got %d\n", counter)
	}
}

func Test_TaskRunOnce(t *testing.T) {

	pingChan := make(chan int, 10)
	errorChan := make(chan error, 1)
	tFunc := func() error {
		pingChan <- 1
		return nil
	}
	task := Task{
		Every: time.Millisecond,
		Todo:  tFunc,
	}

	go task.RunOnce(errorChan)

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

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter != 1 {
		t.Errorf("Expected 1, got %d\n", counter)
	}
}

func Test_TaskRunOnceErrorless(t *testing.T) {

	pingChan := make(chan int, 10)
	errorChan := make(chan error, 1)
	tFunc := func() {
		pingChan <- 1
	}
	task := Task{
		Every: time.Millisecond,
		Todo:  ErrorlessTaskFunc(tFunc),
	}

	go task.RunOnce(errorChan)

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

	if !task.IsCrashed() {
		t.Error("Expected IsCrashed true\n")
	}
	if counter != 1 {
		t.Errorf("Expected 1, got %d\n", counter)
	}
}
