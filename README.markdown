[![Apache V2 License](https://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/eevans/wikimedia/blob/master/LICENSE.txt)

# Wikimedia SDK for Go

Golang packages for working with Wikimedia services.


## `streams` package

The `streams` package provides high-level access to the [Wikimedia
EventStreams service][1].

### Example

```go
func main() {
    client := streams.NewClient()
    
    client.RecentChanges(func(event RecentChangeEvent) {
        fmt.Printf("Change event captured!\n")
        fmt.Printf("  Title .........: %s\n", event.Title)
        fmt.Printf("  Server name ...: %s\n", event.ServerName)
        fmt.Printf("  Wiki ..........: %s\n", event.Wiki)
        fmt.Printf("  Namespace .....: %d\n", event.Namespace)
    })
}
```

Additionally, you can filter the list of events to those matching a
set of predicates.

```go
func main() {
    // Only produce events where the namespace attribute is 0, and wiki is enwiki
    client := streams.NewClient().Match("namespace": 0).Match("wiki": "enwiki")
    
    // ...
}
```

### Known issues

* Only the `recentchange` stream is currently supported
* Test coverage is very poor


[1]: https://wikitech.wikimedia.org/wiki/Event_Platform/EventStreams
