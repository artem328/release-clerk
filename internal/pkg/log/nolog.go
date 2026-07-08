package log

type NoLogLogger struct{}

func (NoLogLogger) Log(...any) {}

func (NoLogLogger) Logf(string, ...any) {}

func (NoLogLogger) Debug(...any) {}

func (NoLogLogger) Debugf(string, ...any) {}

func (NoLogLogger) Section(string) Logger { return NoLogLogger{} }
