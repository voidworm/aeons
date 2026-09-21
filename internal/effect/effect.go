package effect

type Effect interface {
	Apply()
}

type Ticker interface {
	Tick()
}
