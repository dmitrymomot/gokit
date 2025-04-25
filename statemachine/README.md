# State Machine

A flexible, type-safe state machine implementation for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/statemachine
```

## Overview

The `statemachine` package provides a clean, flexible implementation of the state machine pattern for Go applications. It offers a fluent builder API and type-safety through interfaces, making it ideal for modeling complex workflows, business processes, and application states.

## Features

- Fluent builder pattern for intuitive state machine construction
- Type-safe implementation using Go interfaces and generics
- Built-in support for guards (transition conditions) and actions (side effects)
- Thread-safe operation with mutex locks for concurrent access
- String-based or custom state/event implementations
- Comprehensive error handling with specific error types
- Simple API with minimal boilerplate

## Usage

### Basic State Machine

```go
import (
	"context"
	"fmt"
	"github.com/dmitrymomot/gokit/statemachine"
)

func main() {
	// Define states using string constants
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

	// Create and configure the state machine using the fluent builder API
	builder := statemachine.NewBuilder(Draft)
	builder.From(Draft).When(Submit).To(InReview).Add()
	builder.From(InReview).When(Approve).To(Approved).Add()
	builder.From(InReview).When(Reject).To(Rejected).Add()
	builder.From(Approved).When(Publish).To(Published).Add()
	builder.From(Approved).When(Withdraw).To(Draft).Add()
	builder.From(Rejected).When(Submit).To(InReview).Add()

	machine := builder.Build()

	// Use the state machine
	ctx := context.Background()
	fmt.Printf("Current state: %s\n", machine.Current().Name()) // "draft"

	// Trigger transitions
	machine.Fire(ctx, Submit, nil)
	fmt.Printf("Current state: %s\n", machine.Current().Name()) // "in_review"
	
	machine.Fire(ctx, Approve, nil)
	fmt.Printf("Current state: %s\n", machine.Current().Name()) // "approved"
}
```

### Guards and Actions

```go
// Add conditional transition with a guard
isAuthorized := func(ctx context.Context, from statemachine.State, event statemachine.Event, data any) bool {
	userData, ok := data.(map[string]any)
	return ok && userData["is_authorized"].(bool)
}

// Add a side effect with an action
logTransition := func(ctx context.Context, from, to statemachine.State, event statemachine.Event, data any) error {
	fmt.Printf("Transition: %s -> %s via %s\n", from.Name(), to.Name(), event.Name())
	return nil
}

// Add a transition with guard and action
builder.From(Idle).When(Start).To(Running)
	.WithGuard(isAuthorized)
	.WithAction(logTransition)
	.Add()

// Fire event with context data
userData := map[string]any{"is_authorized": true, "user_id": 123}
machine.Fire(ctx, Start, userData)
```

### Custom State and Event Types

```go
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

// Create state machine with custom types
states := struct {
	New        OrderState
	Processing OrderState
	Shipped    OrderState
}{
	New:        OrderState{code: "new", description: "Order created"},
	Processing: OrderState{code: "processing", description: "In process"},
	Shipped:    OrderState{code: "shipped", description: "Shipped"},
}

events := struct {
	Process OrderEvent
	Ship    OrderEvent
}{
	Process: OrderEvent{code: "process"},
	Ship:    OrderEvent{code: "ship"},
}

// Configure state machine
machine := statemachine.NewSimpleStateMachine(states.New)
machine.AddTransition(states.New, states.Processing, events.Process, nil, nil)
machine.AddTransition(states.Processing, states.Shipped, events.Ship, nil, nil)
```

### Error Handling

```go
// Try an invalid transition
err := machine.Fire(ctx, InvalidEvent, nil)
if statemachine.IsNoTransitionAvailableError(err) {
	fmt.Printf("Error: %v\n", err) // No transition defined for event
}

// Add a guard that rejects the transition
alwaysFalse := func(ctx context.Context, from statemachine.State, event statemachine.Event, data any) bool {
	return false
}

machine.AddTransition(Initial, Final, Event, []statemachine.Guard{alwaysFalse}, nil)

// Try a transition that will be rejected by the guard
err = machine.Fire(ctx, Event, nil)
if statemachine.IsTransitionRejectedError(err) {
	fmt.Printf("Error: %v\n", err) // Transition rejected by guard
}
```

## API Reference

### Core Types

```go
// State interface - implement this for custom states
type State interface {
	Name() string
}

// Event interface - implement this for custom events
type Event interface {
	Name() string
}

// Action function type - executed during transitions
type Action func(ctx context.Context, from, to State, event Event, data any) error

// Guard function type - controls whether transitions can occur
type Guard func(ctx context.Context, from State, event Event, data any) bool

// Transition structure - represents a possible state change
type Transition struct {
	From    State
	To      State
	Event   Event
	Guards  []Guard
	Actions []Action
}
```

### State Machine Interface

```go
// StateMachine interface - core API
type StateMachine interface {
	// Returns the current state
	Current() State
	
	// Adds a new transition
	AddTransition(from, to State, event Event, guards []Guard, actions []Action) error
	
	// Triggers an event, potentially causing a state transition
	Fire(ctx context.Context, event Event, data any) error
	
	// Checks if an event can be fired in the current state
	CanFire(ctx context.Context, event Event, data any) bool
	
	// Resets to initial state
	Reset() error
}
```

### Builder Pattern

```go
// Create a new builder with initial state
builder := statemachine.NewBuilder(initialState)

// Fluent API for defining transitions
builder.From(stateA).When(eventX).To(stateB).WithGuard(guard).WithAction(action).Add()

// Build the state machine
machine := builder.Build()
```

### Helper Functions

```go
// Check for specific error types
statemachine.IsNoTransitionAvailableError(err)
statemachine.IsTransitionRejectedError(err)
statemachine.IsInvalidTransitionError(err)
statemachine.IsActionExecutionError(err)
```

## Best Practices

1. **Separate State Logic**: Use the state machine to handle state transitions, but keep business logic in your application code.

2. **Use the Builder Pattern**: The fluent builder API is more readable than direct transition creation, especially for complex state machines.

3. **Context Propagation**: Always pass the context through the state machine to ensure proper cancellation and timeout handling.

4. **Immutable State Objects**: Define state and event objects as immutable to prevent bugs from unexpected mutations.

5. **Error Handling**: Check for specific error types using the provided helper functions rather than string comparisons.

6. **Thread Safety**: The state machine is safe for concurrent use, but consider whether your application needs additional synchronization.

7. **Testing**: Create test cases that validate state transitions, especially those with guards and actions.

8. **Documentation**: Document your states, events, and transitions to make the state machine's behavior clear to others.

## Thread Safety

The state machine implementation is thread-safe for concurrent access with internal mutex locks. All methods that modify state are protected, making the state machine safe to use in concurrent environments without additional synchronization.
