// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2017-2026 Unstable Build, All Rights Reserved.
//
// NOTICE: All information contained herein is, and remains the property of COMPANY.
// The intellectual and technical concepts contained herein are proprietary to
// COMPANY and may be covered by U.S. and Foreign Patents, patents in process,
// and are protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written permission
// is obtained from COMPANY. Access to the source code contained herein is hereby
// forbidden to anyone except current COMPANY employees, managers or contractors who
// have executed Confidentiality and Non-disclosure agreements explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or
// disclosure of this source code, which includes information that is confidential and/or
// proprietary, and is a trade secret, of COMPANY. ANY REPRODUCTION, MODIFICATION,
// DISTRIBUTION, PUBLIC  PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE CODE
// WITHOUT  THE EXPRESS WRITTEN CONSENT OF COMPANY IS STRICTLY PROHIBITED, AND IN
// VIOLATION OF APPLICABLE LAWS AND INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF
// THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY OR IMPLY ANY RIGHTS TO
// REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL
// ANYTHING THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.


// Package jsonrpc2 is a minimal implementation of the JSON RPC 2 spec.
// https://www.jsonrpc.org/specification
// It is intended to be compatible with other implementations at the wire level.
package jsonrpc2

import (
	"context"
	"errors"
)

var (
	// ErrIdleTimeout is returned when serving timed out waiting for new connections.
	ErrIdleTimeout = errors.New("timed out waiting for new connections")

	// ErrNotHandled is returned from a Handler or Preempter to indicate it did
	// not handle the request.
	//
	// If a Handler returns ErrNotHandled, the server replies with
	// ErrMethodNotFound.
	ErrNotHandled = errors.New("JSON RPC not handled")

	// ErrAsyncResponse is returned from a handler to indicate it will generate a
	// response asynchronously.
	//
	// ErrAsyncResponse must not be returned for notifications,
	// which do not receive responses.
	ErrAsyncResponse = errors.New("JSON RPC asynchronous response")
)

// Preempter handles messages on a connection before they are queued to the main
// handler.
// Primarily this is used for cancel handlers or notifications for which out of
// order processing is not an issue.
type Preempter interface {
	// Preempt is invoked for each incoming request before it is queued for handling.
	//
	// If Preempt returns ErrNotHandled, the request will be queued,
	// and eventually passed to a Handle call.
	//
	// Otherwise, the result and error are processed as if returned by Handle.
	//
	// Preempt must not block. (The Context passed to it is for Values only.)
	Preempt(ctx context.Context, req *Request) (result any, err error)
}

// A PreempterFunc implements the Preempter interface for a standalone Preempt function.
type PreempterFunc func(ctx context.Context, req *Request) (any, error)

func (f PreempterFunc) Preempt(ctx context.Context, req *Request) (any, error) {
	return f(ctx, req)
}

var _ Preempter = PreempterFunc(nil)

// Handler handles messages on a connection.
type Handler interface {
	// Handle is invoked sequentially for each incoming request that has not
	// already been handled by a Preempter.
	//
	// If the Request has a nil ID, Handle must return a nil result,
	// and any error may be logged but will not be reported to the caller.
	//
	// If the Request has a non-nil ID, Handle must return either a
	// non-nil, JSON-marshalable result, or a non-nil error.
	//
	// The Context passed to Handle will be canceled if the
	// connection is broken or the request is canceled or completed.
	// (If Handle returns ErrAsyncResponse, ctx will remain uncanceled
	// until either Cancel or Respond is called for the request's ID.)
	Handle(ctx context.Context, req *Request) (result any, err error)
}

type defaultHandler struct{}

func (defaultHandler) Preempt(context.Context, *Request) (any, error) {
	return nil, ErrNotHandled
}

func (defaultHandler) Handle(context.Context, *Request) (any, error) {
	return nil, ErrNotHandled
}

// A HandlerFunc implements the Handler interface for a standalone Handle function.
type HandlerFunc func(ctx context.Context, req *Request) (any, error)

func (f HandlerFunc) Handle(ctx context.Context, req *Request) (any, error) {
	return f(ctx, req)
}

var _ Handler = HandlerFunc(nil)

// async is a small helper for operations with an asynchronous result that you
// can wait for.
type async struct {
	ready    chan struct{} // closed when done
	firstErr chan error    // 1-buffered; contains either nil or the first non-nil error
}

func newAsync() *async {
	var a async
	a.ready = make(chan struct{})
	a.firstErr = make(chan error, 1)
	a.firstErr <- nil
	return &a
}

func (a *async) done() {
	close(a.ready)
}

func (a *async) wait() error {
	<-a.ready
	err := <-a.firstErr
	a.firstErr <- err
	return err
}

func (a *async) setError(err error) {
	storedErr := <-a.firstErr
	if storedErr == nil {
		storedErr = err
	}
	a.firstErr <- storedErr
}
