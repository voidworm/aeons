package cli

type CallReference struct {
	Resource string
	Verb     string
	Params   map[string]string
}
