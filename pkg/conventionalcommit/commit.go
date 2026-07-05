package conventionalcommit

type Footer struct {
	Token string
	Value string
}

type Commit struct {
	Raw            string
	RawHeader      string
	RawFooter      string
	Type           string
	Scope          string
	Description    string
	Body           string
	Footers        []Footer
	IsBreaking     bool
	IsConventional bool
}
