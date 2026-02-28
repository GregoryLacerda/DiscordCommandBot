package events

import (
	"enque-learning/internal/errors"
	"sync"
)

type EventDispatcher struct {
	handlers map[string][]EventHandlerInterface
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandlerInterface),
	}
}

func (e *EventDispatcher) Dispatch(event EventInterface) error {
	if handlers, ok := e.handlers[event.GetName()]; ok {
		wg := sync.WaitGroup{}
		wg.Add(len(handlers))
		errorChannel := make(chan error, len(handlers))

		for _, handler := range handlers {
			wg.Go(func() {
				defer wg.Done()

				err := handler.HandleEvent(event)
				if err != nil {
					errorChannel <- err
					return
				}
			})
		}

		wg.Wait()
		close(errorChannel)

		if len(errorChannel) == len(handlers) {
			return <-errorChannel
		}

		for err := range errorChannel {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *EventDispatcher) RegisterHandler(eventName string, handler EventHandlerInterface) error {

	if !e.HasHandler(eventName, handler) {
		e.handlers[eventName] = append(e.handlers[eventName], handler)
		return nil
	}

	return errors.NewAlreadyRegisteredError(eventName)
}

func (e *EventDispatcher) RemoveHandler(eventName string, handler EventHandlerInterface) error {
	if handlers, ok := e.handlers[eventName]; ok {
		for i, h := range handlers {
			if h == handler {
				e.handlers[eventName] = append(handlers[:i], handlers[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (e *EventDispatcher) HasHandler(eventName string, handler EventHandlerInterface) bool {
	if handlers, ok := e.handlers[eventName]; ok {
		for _, h := range handlers {
			if h == handler {
				return true
			}
		}
	}
	return false
}

func (e *EventDispatcher) HasAnyHandler(eventName string) bool {
	if handlers, ok := e.handlers[eventName]; ok {
		return len(handlers) > 0
	}
	return false
}

func (e *EventDispatcher) ClearHandlers() error {
	e.handlers = make(map[string][]EventHandlerInterface)
	return nil
}
