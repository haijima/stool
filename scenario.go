package stool

import (
	"io"
)

type ScenarioProfiler struct {
}

type ScenarioOption struct {
	MatchingGroups []string
	IgnorePatterns []string
	TimeFormat     string
}

func NewScenarioProfiler() *ScenarioProfiler {
	return &ScenarioProfiler{}
}

type Pattern struct{}

func (p *ScenarioProfiler) Profile(in io.Reader, opt ScenarioOption) (*Pattern, error) {
	return nil, nil
}
