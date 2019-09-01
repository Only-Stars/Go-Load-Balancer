# A Load Balanced Time Service

A simple load balancer to make a time service more reliable. The load balancer has the following interface:

```go
// LoadBalancer is used for balancing load between multiple instances of a service.
type LoadBalancer interface {
	Request(payload interface{}) chan Response
	RegisterInstance(chan Request)
}

// Request is used for making requests to services behind a load balancer.
type Request struct {
	Payload interface{}
	RspChan chan Response
}

// Response is the value returned by services behind a load balancer.
type Response interface{}
```

When a new instance of a time service appears, it will register by calling `RegisterInstance` with a request channel. When a user wants to know the time, the method `Request` will be invoked. The load balancer should forward this request to a time instance and return the response. Your task is to implement `RegisterInstance` and `Request`, such that instances may appear and go away with best possible quality of service. As an example, if two time service instances register with the load balancer, then one instance stops responding, the load balancer should still ensure that requests are serviced appropriately.
