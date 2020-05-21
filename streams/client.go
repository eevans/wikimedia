package streams

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/r3labs/sse"
)

const (
	// DefaultURL is the URL of the Wikimedia EventStreams service (sans any stream endpoints)
	DefaultURL string = "https://stream.wikimedia.org/v2/stream"
)

// Client is used to subscribe to the Wikimedia EventStreams service
type Client struct {
	URL        string
	Predicates map[string]interface{}
}

// NewClient returns an initialized Client
func NewClient() *Client {
	return &Client{DefaultURL, make(map[string]interface{})}
}

// Match adds a new predicate.  Predicates are used to used to establish a match based on the JSON
// attribute name.  Events match only when all predicates do.
func (client *Client) Match(attribute string, value interface{}) *Client {
	client.Predicates[attribute] = value
	return client
}

// RecentChanges subscribes to the recent changes feed. The handler is invoked with a
// RecentChangeEvent once for every matching event received.
func (client *Client) RecentChanges(handler func(evt RecentChangeEvent)) {
	sseClient := sse.NewClient(fmt.Sprintf("%s/recentchange", DefaultURL))

	sseClient.Subscribe("message", func(msg *sse.Event) {
		// This actually happens; The first event that fires is always empty
		if len(msg.Data) == 0 {
			return
		}

		evt := RecentChangeEvent{}
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Printf("Error deserializing JSON event: %s\n", err)
			return
		}

		if matching(reflect.ValueOf(evt), client.Predicates) {
			handler(evt)
		}
	})
}

// Given the reflect.Value for a JSON annotated event struct, and a predicate map, returns true if all predicates match.
func matching(v reflect.Value, p map[string]interface{}) bool {
	matches := 0

	// Iterate over the fields of a JSON annotated event struct...
	for i := 0; i < v.NumField(); i++ {
		// Obtain the value of the 'json' annotation...
		tag := v.Type().Field(i).Tag.Get("json")
		// Skip when missing or unset...
		if tag == "" || tag == "-" {
			continue
		}

		// If a predicate exists for this field, and it matches, add it to our count...
		if predicate, present := p[tag]; present {
			if predicate == v.Field(i).Interface() {
				matches++
			}
		}
	}

	// If the number of matches is equal to the number of predicates ('all matching'), then the event is a match.
	if matches == len(p) {
		return true
	}

	return false
}
