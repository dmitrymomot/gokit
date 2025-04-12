# State Machine

A flexible, type-safe state machine implementation for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/statemachine
```

## Features

- Simple, intuitive API for defining state machines
- Type-safety through interfaces and generics
- Fluent builder pattern for easy construction
- Support for guards and actions during transitions
- Concurrent access safety through mutex locks
- Clear error messages for troubleshooting
- String-based or custom state/event implementations

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/statemachine"
)

func main() {
	// Define states
	const (
		Draft     = statemachine.StringState("draft")
		InReview  = statemachine.StringState("in_review")
		Approved  = statemachine.StringState("approved")
		Published = statemachine.StringState("published")
		Rejected  = statemachine.StringState("rejected")
	)

	// Define events
	const (
		Submit   = statemachine.StringEvent("submit")
		Approve  = statemachine.StringEvent("approve")
		Reject   = statemachine.StringEvent("reject")
		Publish  = statemachine.StringEvent("publish")
		Withdraw = statemachine.StringEvent("withdraw")
	)

	// Create a state machine
	builder := statemachine.NewBuilder(Draft)

	// Define transitions
	builder.From(Draft).When(Submit).To(InReview).Add()
	builder.From(InReview).When(Approve).To(Approved).Add()
	builder.From(InReview).When(Reject).To(Rejected).Add()
	builder.From(Approved).When(Publish).To(Published).Add()
	builder.From(Approved).When(Withdraw).To(Draft).Add()
	builder.From(Rejected).When(Submit).To(InReview).Add()

	machine := builder.Build()

	// Use the state machine
	ctx := context.Background()

	fmt.Printf("Current state: %s\n", machine.Current().Name())

	if err := machine.Fire(ctx, Submit, nil); err != nil {
		log.Fatalf("Failed to submit: %v", err)
	}

	fmt.Printf("Current state: %s\n", machine.Current().Name())
}
```

### Advanced Usage with Guards and Actions

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dmitrymomot/gokit/statemachine"
)

func main() {
	// Define states and events
	const (
		Idle      = statemachine.StringState("idle")
		Running   = statemachine.StringState("running")
		Paused    = statemachine.StringState("paused")
		Completed = statemachine.StringState("completed")
		Failed    = statemachine.StringState("failed")
	)

	const (
		Start   = statemachine.StringEvent("start")
		Pause   = statemachine.StringEvent("pause")
		Resume  = statemachine.StringEvent("resume")
		Finish  = statemachine.StringEvent("finish")
		Fail    = statemachine.StringEvent("fail")
		Restart = statemachine.StringEvent("restart")
	)

	// Create state machine
	builder := statemachine.NewBuilder(Idle)

	// Add a guard function
	isAuthorized := func(ctx context.Context, from statemachine.State, event statemachine.Event, data any) bool {
		if userData, ok := data.(map[string]any); ok {
			return userData["is_authorized"].(bool)
		}
		return false
	}

	// Add an action function
	logTransition := func(ctx context.Context, from, to statemachine.State, event statemachine.Event, data any) error {
		fmt.Printf("Transitioning from %s to %s via %s at %s\n",
			from.Name(), to.Name(), event.Name(), time.Now().Format(time.RFC3339))
		return nil
	}

	// Define transitions with guards and actions
	builder.From(Idle).When(Start).To(Running).WithGuard(isAuthorized).WithAction(logTransition).Add()
	builder.From(Running).When(Pause).To(Paused).WithAction(logTransition).Add()
	builder.From(Paused).When(Resume).To(Running).WithGuard(isAuthorized).WithAction(logTransition).Add()
	builder.From(Running).When(Finish).To(Completed).WithAction(logTransition).Add()
	builder.From(Running).When(Fail).To(Failed).WithAction(logTransition).Add()
	builder.From(Failed).When(Restart).To(Idle).WithAction(logTransition).Add()

	machine := builder.Build()

	// Use the state machine with user data
	ctx := context.Background()
	userData := map[string]any{
		"is_authorized": true,
		"user_id": 123,
	}

	fmt.Printf("Current state: %s\n", machine.Current().Name())

	if err := machine.Fire(ctx, Start, userData); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	fmt.Printf("Current state: %s\n", machine.Current().Name())
}
```

### Custom State and Event Types

```go
package main

import (
	"context"
	"fmt"

	"github.com/dmitrymomot/gokit/statemachine"
)

// Define custom state type
type OrderState struct {
	code        string
	description string
}

func (s OrderState) Name() string {
	return s.code
}

// Define custom event type
type OrderEvent struct {
	code string
	data map[string]any
}

func (e OrderEvent) Name() string {
	return e.code
}

func main() {
	// Define states with custom type
	states := struct {
		New       OrderState
		Processing OrderState
		Shipped    OrderState
		Delivered  OrderState
		Cancelled  OrderState
	}{
		New:       OrderState{code: "new", description: "Order has been created"},
		Processing: OrderState{code: "processing", description: "Order is being processed"},
		Shipped:    OrderState{code: "shipped", description: "Order has been shipped"},
		Delivered:  OrderState{code: "delivered", description: "Order has been delivered"},
		Cancelled:  OrderState{code: "cancelled", description: "Order has been cancelled"},
	}

	// Define events with custom type
	events := struct {
		Process  OrderEvent
		Ship     OrderEvent
		Deliver  OrderEvent
		Cancel   OrderEvent
	}{
		Process:  OrderEvent{code: "process", data: map[string]any{}},
		Ship:     OrderEvent{code: "ship", data: map[string]any{}},
		Deliver:  OrderEvent{code: "deliver", data: map[string]any{}},
		Cancel:   OrderEvent{code: "cancel", data: map[string]any{}},
	}

	// Create and configure state machine
	machine := statemachine.NewSimpleStateMachine(states.New)

	// Add transitions
	machine.AddTransition(states.New, states.Processing, events.Process, nil, nil)
	machine.AddTransition(states.Processing, states.Shipped, events.Ship, nil, nil)
	machine.AddTransition(states.Shipped, states.Delivered, events.Deliver, nil, nil)
	machine.AddTransition(states.New, states.Cancelled, events.Cancel, nil, nil)
	machine.AddTransition(states.Processing, states.Cancelled, events.Cancel, nil, nil)

	// Use the state machine
	ctx := context.Background()

	fmt.Printf("Current state: %s\n", machine.Current().Name())

	if err := machine.Fire(ctx, events.Process, nil); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Current state: %s\n", machine.Current().Name())
}
```

## Error Handling

```go
package main

import (
	"context"
	"fmt"

	"github.com/dmitrymomot/gokit/statemachine"
)

func main() {
	// Define states and events
	const (
		Initial = statemachine.StringState("initial")
		Final   = statemachine.StringState("final")
	)

	const (
		Valid   = statemachine.StringEvent("valid")
		Invalid = statemachine.StringEvent("invalid")
	)

	// Create state machine with only one valid transition
	machine := statemachine.NewSimpleStateMachine(Initial)
	machine.AddTransition(Initial, Final, Valid, nil, nil)

	ctx := context.Background()

	// Try an invalid transition
	err := machine.Fire(ctx, Invalid, nil)
	if statemachine.IsNoTransitionAvailableError(err) {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Add a guard that always fails
	alwaysFalse := func(ctx context.Context, from statemachine.State, event statemachine.Event, data any) bool {
		return false
	}

	machine.AddTransition(Initial, Final, Invalid, []statemachine.Guard{alwaysFalse}, nil)

	// Try a transition that will be rejected by the guard
	err = machine.Fire(ctx, Invalid, nil)
	if statemachine.IsTransitionRejectedError(err) {
		fmt.Printf("Expected error: %v\n", err)
	}
}
```

## Implementation Notes

- The state machine is safe for concurrent use with its internal mutex
- All methods that modify state are thread-safe
- Custom state and event implementations can be created by implementing the `State` and `Event` interfaces
- Guards can be used to implement complex conditional transitions
- Actions allow side effects during transitions while maintaining the state machine's integrity

## License

MIT
