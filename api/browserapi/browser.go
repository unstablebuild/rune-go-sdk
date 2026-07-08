// Copyright 2026 Unstable Build, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package browserapi

import (
	"context"
	"io"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// Handler adds Close to a tui.Handler. Close must be idempotent.
type Handler interface {
	tui.Handler
	Close() error
}

// Floating is a Handler used for Floating windows.
// See handler.Floating for more details.
type Floating interface {
	Handler
	handler.Floating
}

// Window is the interface that represents
// a closeable window in a WindowManager.
type Window interface {
	WindowID() uint64
}

// Orientation represents a window orientation.
type Orientation uint8

const (
	// OrientationDefault represents the default window orientation.
	OrientationDefault Orientation = iota
	// OrientationTop sets the component at the top.
	OrientationTop
	// OrientationBottom sets the component at the bottom.
	OrientationBottom
	// OrientationLeft sets the component on the left.
	OrientationLeft
	// OrientationRight sets the component on the right.
	OrientationRight
)

// BarFrame represents the different options for creating a bar.
type BarFrame uint8

const (
	// BarFrameDefault signals that the bar should use the default frame
	// configuration of the WindowManager.
	BarFrameDefault BarFrame = iota
	// BarFrameAlways signals that the bar should always be framed,
	// regardless of the WindowManager configuration.
	BarFrameAlways
	// BarFrameNever signals that the bar should never be framed,
	// regardless of the WindowManager configuration.
	BarFrameNever
)

// BarConfig defines the configuration for a Bar
// created bia WindowManager.Bar.
type BarConfig struct {
	Orientation Orientation
	Size        int
	Frame       BarFrame
}

// WindowManager is the interface that groups tile
// window management methods.
type WindowManager interface {
	// Focus returns the current Window in focus.
	Focus() (Window, error)

	// Split splits the current window in focus in two, and installs
	// Handler in the new window.
	Split(Orientation, Window, Handler) (Window, error)

	// Floating creates a new floating window at coordinates,
	// with static width and height.
	Floating(h Floating, cfg FloatingConfig) (Window, error)

	// Bar creates a status bar with Orientation and Handler.
	// Bars differ from Split and Floating windows in that they can't
	// be in focus and can only receive mouse events.
	Bar(BarConfig, tui.Handler) error

	// Tab creates a new tab with h and returns a handle that can be
	// used with the rest of methods that take a browser.Handler.
	// URI is used to uniquely identify a tab and name is used as a label
	// to display it in the tab bar.
	Tab(uri workspaceapi.URI, icon rune, name string, h Handler) (Handler, error)

	// SetWindowContent sets the content of the given window to the given handler.
	SetWindowContent(Window, Handler) error

	// Close closes the given window.
	CloseWindow(Window) error
}

// FloatingConfig abstracts configuration for
// creating floating windows.
type FloatingConfig struct {
	// Sets the alignment of the window.
	component.Alignment
	// Offset is to be applied to the position of the window
	// after alignment has been determined.
	Offset term.Coordinates
	// NoWindowBar keeps the plain window frame instead of the window
	// bar when the window manager decorates floating windows with one.
	NoWindowBar bool
	// Title is rendered on the window bar.
	Title string
}

// NotificationLevel determines the severity of the notification.
type NotificationLevel uint8

const (
	// LevelError signals an unexpected error.
	LevelError NotificationLevel = iota
	// LevelWarn signals an expected error.
	LevelWarn
	// LevelInfo signals an informational message.
	LevelInfo
	// LevelSuccess signals a message of success.
	LevelSuccess
)

// Notifications is the interface that wraps methods to display
// messages to the user.
type Notifications interface {
	// Notify delivers the given message to the user, as a notification.
	Notify(level NotificationLevel, msg string, args ...any) (string, error)

	// NotifyOnce delivers the given notification once, and never again. A notification
	// is identified as the hash of the final message (with arguments).
	NotifyOnce(level NotificationLevel, msg string, args ...any) (string, error)

	// UpdateNotificationProgress updates the progress of a notification,
	// overriding the default timing and progress bar settings.
	//
	// After calling this once, caller is responsible for updating it
	// until progress == total.
	//
	// It's expected for progress to be smaller than or equal than total,
	// and total must never be 0. An error is returned if any of these conditions
	// are not met.
	//
	// The message argument can be empty, in which case the previous message
	// is used.
	UpdateNotificationProgress(id, message string, progress, total int64) error
}

// ResourceOpener is the interface that wraps the method Open.
type ResourceOpener interface {
	// Opens the resource with a given uri as a tab, using the underlying
	// workspaceapi.FileSystem, but doesn't switch the focus of the current
	// window in focus to it. That can be accomplished via WindowManager.SetWindowContent.
	Open(resource workspaceapi.URI) (Handler, error)
}

// EventPublisher is the interface that wraps the method Interrupt.
type EventPublisher interface {
	// Interrupt will publish an interrupt event, which will force
	// redrawing all components in the terminal.
	Interrupt(context.Context) error

	// PublishEventNone will publish an EventNone event, which will force
	// calling Handle on the component currently in focus.
	//
	// There's no guarantee that caller will be the tui.Handler that will
	// receive this event.
	PublishEventNone() error
}

// Browser is an interface that groups methods to manipulate
// the user interface of a browser.
type Browser interface {
	WindowManager
	EventPublisher
	ResourceOpener
	Notifications
	io.Closer
}
