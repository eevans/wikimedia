package streams

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// const sampleEvent string = `
// {
// 	"$schema": "/mediawiki/recentchange/1.0.0",
// 	"meta": {
// 	  "uri": "https://commons.wikimedia.org/wiki/File:Abydos-Bold-hieroglyph-O10A.png",
// 	  "request_id": "eb2f4e7b-0aaa-4df0-8dc6-d49cfbb62178",
// 	  "id": "013846b7-9e2a-430a-8959-7a423bf38385",
// 	  "dt": "2020-05-21T00:26:42Z",
// 	  "domain": "commons.wikimedia.org",
// 	  "stream": "mediawiki.recentchange",
// 	  "topic": "eqiad.mediawiki.recentchange",
// 	  "partition": 0,
// 	  "offset": 2420955111
// 	},
// 	"id": 1390275467,
// 	"type": "edit",
// 	"namespace": 6,
// 	"title": "File:Abydos-Bold-hieroglyph-O10A.png",
// 	"comment": "/* wbeditentity-update:0| */ automatically adding claims based on file information: date",
// 	"timestamp": 1590020802,
// 	"user": "SchlurcherBot",
// 	"bot": true,
// 	"minor": false,
// 	"patrolled": true,
// 	"length": {
// 	  "old": 366,
// 	  "new": 1010
// 	},
// 	"revision": {
// 	  "old": 326564921,
// 	  "new": 420613296
// 	},
// 	"server_url": "https://commons.wikimedia.org",
// 	"server_name": "commons.wikimedia.org",
// 	"server_script_path": "/w",
// 	"wiki": "commonswiki",
// 	"parsedcomment": "‎<span dir=\"auto\"><span class=\"autocomment\">Ein Objekt geändert: </span></span> automatically adding claims based on file information: date"
//   }
// `

// var url string
// var httpSrv *httptest.Server
// var sseSrv *sse.Server

// func setup() {
// 	sseSrv = sse.New()

// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/events", sseSrv.HTTPHandler)
// 	httpSrv = httptest.NewServer(mux)
// 	url = httpSrv.URL + "/events"

// 	sseSrv.CreateStream("message")

// 	go func() {
// 		for a := 0; a < 100000000; a++ {
// 			sseSrv.Publish("message", &sse.Event{Data: []byte(sampleEvent)})
// 			time.Sleep(time.Millisecond * 50)
// 		}
// 	}()
// }

// func cleanup() {
// 	httpSrv.CloseClientConnections()
// 	httpSrv.Close()
// 	sseSrv.Close()
// }

func TestRecentChanges(t *testing.T) {
	// setup()
	// defer cleanup()

	client := &Client{BaseURL: DefaultURL, Predicates: map[string]interface{}{"namespace": 0}}
	events := make(chan RecentChangeEvent)

	go func() {
		err := client.RecentChanges(func(evt RecentChangeEvent) {
			events <- evt
		})
		if err != nil {
			require.FailNow(t, "Client returned error: %v", err)
		}
	}()

	select {
	case evt := <-events:
		require.Equal(t, 0, evt.Namespace, "Received event not matching namespace predicate")
	case <-time.After(time.Second * 5):
		require.FailNow(t, "No events received.")
	}
}
