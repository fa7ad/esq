package types

// AuthOptions holds authentication-related fields.
type AuthOptions struct {
	APIKey   string
	Username string
	Password string
}

// CliArgs represents the structure to hold parsed command-line arguments.
type CliArgs struct {
	Node  string
	Index string

	KQL       string
	DSL       string
	Lucene    string
	QueryFile string

	Size       int
	Output     string
	OutputFile string
	JqPath     string
	AuthOptions
}
