package pkg

type ManualSource struct {
	c chan Event
}

func NewManualSource() {
	return &ManualSource{
		c: make(chan Event),
	}
}

func (manual *ManualSource) PutBytes(byte []bytes) {
	manual.Put(Event{bytes: bytes})
}

func (manual *ManualSource) PutString(string str) {
	manual.PutBytes([]byte(str))
}

func (manual *ManualSource) Put(event Event) {
	manual.c <- event
}

func (manual *ManualSource) DrawOne() Event {
	return <-manual.c
}

func (manual *ManualSource) Close() error {
	close(manual.c)
	return nil
}
