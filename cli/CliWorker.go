package cli

import (
	"errors"
	"time"
)

type CliWorker struct {
	run bool

	sleepPeriod int // time between ui refreshes in milliseconds
}

func NewCliWorker(refreshPerSecond int) (*CliWorker, error) {
	if refreshPerSecond < 1 {
		return nil, errors.New("ui must be refreshed at least once per second")
	}
	if refreshPerSecond > 100 {
		return nil, errors.New("ui must NOT be refreshed more than 100 times per second")
	}

	return &CliWorker{
		sleepPeriod: 1000 / refreshPerSecond,
	}, nil
}

func (w *CliWorker) Start() {
	w.run = true
	for w.run {
		// TODO: update ui
		time.Sleep(time.Duration(w.sleepPeriod) * time.Millisecond)
	}
}

func (w *CliWorker) Stop() {
	w.run = false
}
