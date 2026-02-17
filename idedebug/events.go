package idedebug

import "github.com/unstablebuild/rune-go-sdk/api/textapi"

// EditorEvents returns the events that Manager is
// interested in subscribing to.
func EditorEvents() []textapi.EventType {
	return []textapi.EventType{
		textapi.EventTypeOpen,
	}
}
