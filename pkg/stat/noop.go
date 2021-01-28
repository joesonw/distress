package stat

func Noop() Reporter {
	return noop{}
}

type noop struct{}

func (noop) Report(stats ...*Stat) {
}

func (noop) Finish() error { return nil }
