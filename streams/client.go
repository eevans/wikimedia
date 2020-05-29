package streams

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/r3labs/sse"
)

const (
	// DefaultURL is the URL of the Wikimedia EventStreams service (sans any stream endpoints)
	DefaultURL string = "https://stream.wikimedia.org/v2/stream"
)

// Client is used to subscribe to the Wikimedia EventStreams service
type Client struct {
	BaseURL       string
	Predicates    map[string]interface{}
	Since         string
	lastTimestamp string
}

// NewClient returns an initialized Client
func NewClient() *Client {
	return &Client{BaseURL: DefaultURL, Predicates: make(map[string]interface{})}
}

// Match adds a new predicate.  Predicates are used to used to establish a match based on the JSON
// attribute name.  Events match only when all predicates do.
func (client *Client) Match(attribute string, value interface{}) *Client {
	client.Predicates[attribute] = value
	return client
}

// LastTimestamp returns the ISO8601 formatted timestamp of the last event received.
func (client *Client) LastTimestamp() string {
	return client.lastTimestamp
}

// RecentChanges subscribes to the recent changes feed. The handler is invoked with a
// RecentChangeEvent once for every matching event received.
func (client *Client) RecentChanges(handler func(evt RecentChangeEvent)) error {
	sseClient := sse.NewClient(client.url("recentchange"))

	return sseClient.Subscribe("", func(msg *sse.Event) {
		// This actually happens; The first event that fires is always empty
		if len(msg.Data) == 0 {
			return
		}

		evt := RecentChangeEvent{}
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Printf("Error deserializing JSON event: %s\n", err)
			return
		}

		client.lastTimestamp = evt.Meta.Dt

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

func (client *Client) url(stream string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s/%s", client.BaseURL, stream)
	if client.Since != "" {
		fmt.Fprintf(&b, "?since=%s", client.Since)
	}
	return b.String()
}
