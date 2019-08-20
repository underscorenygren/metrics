package pkg

type ChannelPipeline struct {
	MapFn   MapFn
	source  Source
	drainTo Sink
}

func defaulMapFn(evt Event) Event {
	return evt
}

func NewChannelPipeline(source Source, sink Sink) Pipeline {

	return &ChannelPipeline{
		source:  source,
		MapFn:   defaultMapFn,
		drainTo: Sink,
	}
}

func (pipe *ChannelPipeline) Flow() {

	for e := range pipe.source.DrawOne() {
		pipe.drainTo.Drain(
			pipe.MapFn(e))
	}
}
