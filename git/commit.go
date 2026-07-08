package git

import "time"

type Commit struct {
	FullHash      string
	ShortHash     string
	AuthorDate    time.Time
	AuthorName    string
	AuthorEmail   string
	CommiterDate  time.Time
	CommiterName  string
	CommiterEmail string
	Parents       []string
	Body          string
}
