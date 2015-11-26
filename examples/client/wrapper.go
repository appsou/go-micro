/*
	An example of how to use client.Wrapper as middleware
*/
package main

import (
	"fmt"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	c "github.com/micro/go-micro/context"
	example "github.com/micro/go-micro/examples/server/proto/example"
	"golang.org/x/net/context"
)

// log wrapper logs every time a request is made
type logWrapper struct {
	client.Client
}

func (l *logWrapper) Call(ctx context.Context, req client.Request, rsp interface{}) error {
	md, _ := c.GetMetadata(ctx)
	fmt.Printf("[Log Wrapper] ctx: %v service: %s method: %s\n", md, req.Service(), req.Method())
	return l.Client.Call(ctx, req, rsp)
}

// trace wrapper attaches a unique trace ID - timestamp
type traceWrapper struct {
	client.Client
}

func (t *traceWrapper) Call(ctx context.Context, req client.Request, rsp interface{}) error {
	ctx = c.WithMetadata(ctx, map[string]string{
		"X-Trace-Id": fmt.Sprintf("%d", time.Now().Unix()),
	})
	return t.Client.Call(ctx, req, rsp)
}

// Implements client.Wrapper as logWrapper
func logWrap(c client.Client) client.Client {
	return &logWrapper{c}
}

// Implements client.Wrapper as traceWrapper
func traceWrap(c client.Client) client.Client {
	return &traceWrapper{c}
}

// Calls the example service
func call(i int) {
	// Create new request to service go.micro.srv.example, method Example.Call
	req := client.NewRequest("go.micro.srv.example", "Example.Call", &example.Request{
		Name: "John",
	})

	// create context with metadata
	ctx := c.WithMetadata(context.Background(), map[string]string{
		"X-User-Id": "john",
		"X-From-Id": "script",
	})

	rsp := &example.Response{}

	// Call service
	if err := client.Call(ctx, req, rsp); err != nil {
		fmt.Println("call err: ", err, rsp)
		return
	}
}

func main() {
	cmd.Init()

	// Wrap the default client
	client.DefaultClient = logWrap(client.DefaultClient)

	call(0)

	// Wrap using client.Wrap option
	client.DefaultClient = client.NewClient(
		client.Wrap(traceWrap),
		client.Wrap(logWrap),
	)

	call(1)
}
