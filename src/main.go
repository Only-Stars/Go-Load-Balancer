package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Request is used for making requests to services behind a load balancer.
type Request struct {
	Payload interface{}
	RspChan chan Response
}

// Response is the value returned by services behind a load balancer.
type Response interface{}

// LoadBalancer is used for balancing load between multiple instances of a service.
type LoadBalancer interface {
	Request(payload interface{}) chan Response
	RegisterInstance(chan Request)
}

type MyLoadBalancer struct {
	selectedTimerIdx int
	timerInstances   []chan Request
}

// Implementation of MyLoadBalancer logic
func (lb *MyLoadBalancer) getNextTimer(payload interface{}) (chan Response, error) {
	var (
		selectedRequestChannel  chan Request
		selectedResponseChannel = make(chan Response)
	)
	if len(lb.timerInstances) == 0 {
		return nil, errors.New("no timers")
	}
	defer func() {
		if recover() != nil {
			selectedResponseChannel = nil
			lb.timerInstances = append(lb.timerInstances[:lb.selectedTimerIdx], lb.timerInstances[lb.selectedTimerIdx+1:]...)
		}
	}()

	if lb.selectedTimerIdx >= len(lb.timerInstances)-1 {
		lb.selectedTimerIdx = 0
	} else {
		lb.selectedTimerIdx++
	}

	selectedRequestChannel = lb.timerInstances[lb.selectedTimerIdx]
	selectedRequestChannel <- Request{
		Payload: payload,
		RspChan: selectedResponseChannel,
	}

	return selectedResponseChannel, nil
}

func (lb *MyLoadBalancer) Request(payload interface{}) (respChan chan Response) {
	var (
		err error
	)
	if len(lb.timerInstances) > 0 {
		for {
			respChan, err = lb.getNextTimer(payload)
			if err != nil {
				return
			}
			if respChan != nil {
				break
			}
		}
	}

	return
}

// RegisterInstance is currently a dummy implementation. Please implement it!
func (lb *MyLoadBalancer) RegisterInstance(ch chan Request) {
	lb.timerInstances = append(lb.timerInstances, ch)
}

/******************************************************************************
 *  STANDARD TIME SERVICE IMPLEMENTATION -- MODIFY IF YOU LIKE                *
 ******************************************************************************/

// TimeService is a single instance of a time service.
type TimeService struct {
	Dead            chan struct{}
	ReqChan         chan Request
	AvgResponseTime float64
}

// Run will make the TimeService start listening to the two channels Dead and ReqChan.
func (ts *TimeService) Run() {
	for {
		select {
		case <-ts.Dead:
			close(ts.ReqChan)
			return
		case req := <-ts.ReqChan:
			processingTime := time.Duration(ts.AvgResponseTime+1.0-rand.Float64()) * time.Second
			time.Sleep(processingTime)
			req.RspChan <- time.Now()
		}
	}
}

/******************************************************************************
 *  CLI -- YOU SHOULD NOT NEED TO MODIFY ANYTHING BELOW                       *
 ******************************************************************************/

// main runs an interactive console for spawning, killing and asking for the
// time.
func main() {
	rand.Seed(int64(time.Now().Nanosecond()))

	bio := bufio.NewReader(os.Stdin)
	var lb LoadBalancer = &MyLoadBalancer{}

	manager := &TimeServiceManager{}

	for {
		fmt.Printf("> ")
		cmd, err := bio.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command: ", err)
			continue
		}
		switch strings.TrimSpace(cmd) {
		case "kill":
			manager.Kill()
		case "spawn":
			ts := manager.Spawn()
			lb.RegisterInstance(ts.ReqChan)
			go ts.Run()
		case "time":
			fmt.Println("Request for time...")
			select {
				case rsp := <-lb.Request(nil):
					fmt.Println(rsp)
				case <-time.After(5 * time.Second):
					fmt.Println("Timeout")
			}
		default:
			fmt.Printf("Unknown command: %s Available commands: time, spawn, kill\n", cmd)
		}
	}
}

// TimeServiceManager is responsible for spawning and killing.
type TimeServiceManager struct {
	Instances []TimeService
}

// Kill makes a random TimeService instance unresponsive.
func (m *TimeServiceManager) Kill() {
	if len(m.Instances) > 0 {
		n := rand.Intn(len(m.Instances))
		close(m.Instances[n].Dead)
		m.Instances = append(m.Instances[:n], m.Instances[n+1:]...)
	}
}

// Spawn creates a new TimeService instance.
func (m *TimeServiceManager) Spawn() TimeService {
	ts := TimeService{
		Dead:            make(chan struct{}, 0),
		ReqChan:         make(chan Request, 10),
		AvgResponseTime: rand.Float64() * 3,
	}
	m.Instances = append(m.Instances, ts)

	return ts
}
