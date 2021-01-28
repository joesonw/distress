package stat

import (
	"strconv"
	"time"
)

type Stat struct {
	Name      string
	Timestamp time.Time
	Tags      map[string]string
	Fields    map[string]float64
}

func New(name string) *Stat {
	return &Stat{
		Name:      name,
		Timestamp: time.Now(),
		Tags:      map[string]string{},
		Fields:    map[string]float64{},
	}
}

func (s *Stat) SetTime(t time.Time) *Stat {
	s.Timestamp = t
	return s
}

func (s *Stat) Tag(name, value string) *Stat {
	s.Tags[name] = strconv.Quote(value)
	return s
}

func (s *Stat) FloatField(name string, value float64) *Stat {
	s.Fields[name] = value
	return s
}

func (s *Stat) IntField(name string, value int) *Stat {
	s.Fields[name] = float64(value)
	return s
}

func (s *Stat) Int64Field(name string, value int64) *Stat {
	s.Fields[name] = float64(value)
	return s
}

type Reporter interface {
	Report(stats ...*Stat)
	Finish() error
}
