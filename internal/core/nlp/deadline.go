package nlp

import (
	"errors"
	"strings"
	"time"

	"github.com/olebedev/when"
	en "github.com/olebedev/when/rules/en"
)

var parser = func() *when.Parser {
	p := when.New(nil)
	p.Add(en.All...)
	return p
}()

func ParseDeadline(input string, now time.Time) (*time.Time, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return nil, nil
	}
	res, err := parser.Parse(s, now)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("could not parse deadline")
	}
	t := res.Time
	return &t, nil
}
