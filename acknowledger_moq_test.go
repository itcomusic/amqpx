// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package amqpx

import (
	"sync"
)

// Ensure, that AcknowledgerMock does implement Acknowledger.
// If this is not the case, regenerate this file with moq.
var _ Acknowledger = &AcknowledgerMock{}

// AcknowledgerMock is a mock implementation of Acknowledger.
//
//	func TestSomethingThatUsesAcknowledger(t *testing.T) {
//
//		// make and configure a mocked Acknowledger
//		mockedAcknowledger := &AcknowledgerMock{
//			AckFunc: func(tag uint64, multiple bool) error {
//				panic("mock out the Ack method")
//			},
//			NackFunc: func(tag uint64, multiple bool, requeue bool) error {
//				panic("mock out the Nack method")
//			},
//			RejectFunc: func(tag uint64, requeue bool) error {
//				panic("mock out the Reject method")
//			},
//		}
//
//		// use mockedAcknowledger in code that requires Acknowledger
//		// and then make assertions.
//
//	}
type AcknowledgerMock struct {
	// AckFunc mocks the Ack method.
	AckFunc func(tag uint64, multiple bool) error

	// NackFunc mocks the Nack method.
	NackFunc func(tag uint64, multiple bool, requeue bool) error

	// RejectFunc mocks the Reject method.
	RejectFunc func(tag uint64, requeue bool) error

	// calls tracks calls to the methods.
	calls struct {
		// Ack holds details about calls to the Ack method.
		Ack []struct {
			// Tag is the tag argument value.
			Tag uint64
			// Multiple is the multiple argument value.
			Multiple bool
		}
		// Nack holds details about calls to the Nack method.
		Nack []struct {
			// Tag is the tag argument value.
			Tag uint64
			// Multiple is the multiple argument value.
			Multiple bool
			// Requeue is the requeue argument value.
			Requeue bool
		}
		// Reject holds details about calls to the Reject method.
		Reject []struct {
			// Tag is the tag argument value.
			Tag uint64
			// Requeue is the requeue argument value.
			Requeue bool
		}
	}
	lockAck    sync.RWMutex
	lockNack   sync.RWMutex
	lockReject sync.RWMutex
}

// Ack calls AckFunc.
func (mock *AcknowledgerMock) Ack(tag uint64, multiple bool) error {
	if mock.AckFunc == nil {
		panic("AcknowledgerMock.AckFunc: method is nil but Acknowledger.Ack was just called")
	}
	callInfo := struct {
		Tag      uint64
		Multiple bool
	}{
		Tag:      tag,
		Multiple: multiple,
	}
	mock.lockAck.Lock()
	mock.calls.Ack = append(mock.calls.Ack, callInfo)
	mock.lockAck.Unlock()
	return mock.AckFunc(tag, multiple)
}

// AckCalls gets all the calls that were made to Ack.
// Check the length with:
//
//	len(mockedAcknowledger.AckCalls())
func (mock *AcknowledgerMock) AckCalls() []struct {
	Tag      uint64
	Multiple bool
} {
	var calls []struct {
		Tag      uint64
		Multiple bool
	}
	mock.lockAck.RLock()
	calls = mock.calls.Ack
	mock.lockAck.RUnlock()
	return calls
}

// Nack calls NackFunc.
func (mock *AcknowledgerMock) Nack(tag uint64, multiple bool, requeue bool) error {
	if mock.NackFunc == nil {
		panic("AcknowledgerMock.NackFunc: method is nil but Acknowledger.Nack was just called")
	}
	callInfo := struct {
		Tag      uint64
		Multiple bool
		Requeue  bool
	}{
		Tag:      tag,
		Multiple: multiple,
		Requeue:  requeue,
	}
	mock.lockNack.Lock()
	mock.calls.Nack = append(mock.calls.Nack, callInfo)
	mock.lockNack.Unlock()
	return mock.NackFunc(tag, multiple, requeue)
}

// NackCalls gets all the calls that were made to Nack.
// Check the length with:
//
//	len(mockedAcknowledger.NackCalls())
func (mock *AcknowledgerMock) NackCalls() []struct {
	Tag      uint64
	Multiple bool
	Requeue  bool
} {
	var calls []struct {
		Tag      uint64
		Multiple bool
		Requeue  bool
	}
	mock.lockNack.RLock()
	calls = mock.calls.Nack
	mock.lockNack.RUnlock()
	return calls
}

// Reject calls RejectFunc.
func (mock *AcknowledgerMock) Reject(tag uint64, requeue bool) error {
	if mock.RejectFunc == nil {
		panic("AcknowledgerMock.RejectFunc: method is nil but Acknowledger.Reject was just called")
	}
	callInfo := struct {
		Tag     uint64
		Requeue bool
	}{
		Tag:     tag,
		Requeue: requeue,
	}
	mock.lockReject.Lock()
	mock.calls.Reject = append(mock.calls.Reject, callInfo)
	mock.lockReject.Unlock()
	return mock.RejectFunc(tag, requeue)
}

// RejectCalls gets all the calls that were made to Reject.
// Check the length with:
//
//	len(mockedAcknowledger.RejectCalls())
func (mock *AcknowledgerMock) RejectCalls() []struct {
	Tag     uint64
	Requeue bool
} {
	var calls []struct {
		Tag     uint64
		Requeue bool
	}
	mock.lockReject.RLock()
	calls = mock.calls.Reject
	mock.lockReject.RUnlock()
	return calls
}
