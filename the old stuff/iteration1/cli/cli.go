package cli

type CallReference struct {
	Name      string
	Resource string
	Verb     string
	Params   map[string]string
}
