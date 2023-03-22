package closer

import (
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
)

type Effector func() error

var globalCloser = New()

func Add(f ...Effector) {
	globalCloser.Add(f...)
}

func CloseAll() {
	globalCloser.CloseAll()
}

type Closer struct {
	lg          logger.Logger
	closerFuncs []Effector
}

func New() *Closer {
	return &Closer{
		closerFuncs: make([]Effector, 0),
	}
}

func (c *Closer) SetErrorLogger(lg logger.Logger) {
	c.lg = lg
}

func (c *Closer) Add(f ...Effector) {
	c.closerFuncs = append(c.closerFuncs, f...)
}

func (c *Closer) CloseAll() {
	for _, f := range c.closerFuncs {
		if err := f(); err != nil {
			if c.lg != nil {
				c.lg.Error(fmt.Errorf("error on close: %w", err))
			}
		}
	}

	c.closerFuncs = make([]Effector, 0)
}
