package main

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Prober interface {
	Setup()
	Initialize()
	Wait()
	AddTarget(string)
	CloseInput()
	SetOutput(chan<- ScanResult)
	CheckRateLimit()
	SetLimiter(*rate.Limiter)
}

type Probe struct {
	name    string
	options map[string]string
	waiter  sync.WaitGroup
	input   chan string
	output  chan<- ScanResult
	limiter *rate.Limiter
}

func (p *Probe) String() string {
	return p.name
}

func (p *Probe) Wait() {
	p.waiter.Wait()
}

func (p *Probe) Setup() {
	p.name = "generic"
	p.input = make(chan string)
}

func (p *Probe) Initialize() {
	p.Setup()
	p.name = "generic"
}

func (p *Probe) SetOutput(out chan<- ScanResult) {
	p.output = out
}

func (p *Probe) AddTarget(t string) {
	p.input <- t
}

func (p *Probe) CloseInput() {
	close(p.input)
}

func (p *Probe) SetLimiter(limiter *rate.Limiter) {
	p.limiter = limiter
}

func (p *Probe) CheckRateLimit() {
	for !p.limiter.Allow() {
		time.Sleep(10 * time.Millisecond)
	}
}
