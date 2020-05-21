package streams

type meta struct {
	URI       string `json:"uri"`
	RequestID string `json:"request_id"`
	ID        string `json:"id"`
	Dt        string `json:"dt"`
	Domain    string `json:"domain"`
	Stream    string `json:"stream"`
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int    `json:"offset"`
}

type event struct {
	Schema string `json:"$schema"`
}

// RecentChangeEvent corresponds to the JSON event objects returned by the recent changes stream (see:
// https://github.com/wikimedia/mediawiki-event-schemas/blob/master/jsonschema/mediawiki/recentchange/1.0.0.yaml).
type RecentChangeEvent struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Namespace int    `json:"namespace"`
	Title     string `json:"title"`
	Comment   string `json:"comment"`
	Timestamp int    `json:"timestamp"`
	User      string `json:"user"`
	Bot       bool   `json:"bot"`
	Minor     bool   `json:"minor"`
	Patrolled bool   `json:"patrolled"`
	Length    struct {
		Old int
		New int
	}
	Revision struct {
		Old int
		New int
	}
	ServerURL        string `json:"server_url"`
	ServerName       string `json:"server_name"`
	ServerScriptPath string `json:"server_script_path"`
	Wiki             string `json:"wiki"`
	ParsedComment    string `json:"parsedcomment"`
	event
	meta
}
