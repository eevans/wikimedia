package streams

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jpillora/backoff"
	"github.com/r3labs/sse"
)

const (
	// DefaultURL is the URL of the Wikimedia EventStreams service (sans any stream endpoints)
	DefaultURL string = "https://stream.wikimedia.org/v2/stream"

	// Wikimedia's traffic layer will disconnect clients after 15 minutes (see: https://phabricator.wikimedia.org/T242767).
	// This manifests as an http.http2StreamError (https://golang.org/pkg/net/http/?m=all#http2StreamError) returned by
	// sse.Client#Subscribe (code http2ErrCodeNo), (and this error is NOT handled by sse's ReconnectStrategy).  To work
	// around this, errors that are spaced at least `resetInterval` apart will be tried `retries` times with an
	// exponential back-off.
	resetInterval = time.Minute * 10
	retries       = 3
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
	var bOff = &backoff.Backoff{}
	var err error
	var events *sse.Client
	var lastSub time.Time

	for {
		// Reconnect on each iteration; Client#url will include a `since` param with the
		// timestamp of the last observed event from any previous iterations.
		events = sse.NewClient(client.url("recentchange"))
		lastSub = time.Now()

		err = events.Subscribe("message", func(msg *sse.Event) {
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

		if err == nil {
			return err
		}

		// If we've been running for resetInterval or longer, we'll treat this as a new set of retries
		if time.Now().Sub(lastSub) >= resetInterval {
			bOff.Reset()
		}

		// Backoff
		time.Sleep(bOff.Duration())

		// Bail-out if reaching the limit on retries
		if bOff.Attempt() >= retries {
			return err
		}

		// Start the next iteration where we last left off.
		client.Since = client.lastTimestamp
	}
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
