package ddp

import (
	"log"
)

// Logger logs events
// should add some log level info
// at the moment works as event capture as well, should separate
type Logger struct {
	store Store
}

// NewLogger returns a new Logger
// ... WithOptions for customization?
func NewLogger(store Store) *Logger {
	return &Logger{
		store: store,
	}
}

// Append appends new event to logger store
func (l *Logger) Append(data interface{}) error {
	return l.store.C(data)
}

// AppendErr appends new event to logger store
func (l *Logger) AppendErr(err error) error {
	log.Println(err)
	return l.store.C(err)
}

// Get gets information from logger
func (l *Logger) Get(filter interface{}) (interface{}, error) {
	var data interface{}
	return data, nil
}
