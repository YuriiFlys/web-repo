package patterns

import (
	"fmt"
	"time"
)

type ShowCue struct {
	Name          string
	StartsIn      time.Duration
	BPM           int
	LightPreset   string
	SmokeLevelPct int
	ScreenText    string
	Tags          []string
}

type CueBuilder struct {
	cue ShowCue
	err error
}

func NewCueBuilder(name string) *CueBuilder {
	return &CueBuilder{
		cue: ShowCue{
			Name:          name,
			BPM:           120,
			SmokeLevelPct: 0,
			Tags:          make([]string, 0),
		},
	}
}

func (b *CueBuilder) StartsIn(d time.Duration) *CueBuilder {
	b.cue.StartsIn = d
	return b
}

func (b *CueBuilder) BPM(v int) *CueBuilder {
	if b.err != nil {
		return b
	}
	if v <= 0 || v > 400 {
		b.err = fmt.Errorf("BPM must be in range 1..400, got %d", v)
		return b
	}
	b.cue.BPM = v
	return b
}

func (b *CueBuilder) LightPreset(preset string) *CueBuilder {
	b.cue.LightPreset = preset
	return b
}

func (b *CueBuilder) SmokeLevelPct(pct int) *CueBuilder {
	if b.err != nil {
		return b
	}
	if pct < 0 || pct > 100 {
		b.err = fmt.Errorf("SmokeLevelPct must be in range 0..100, got %d", pct)
		return b
	}
	b.cue.SmokeLevelPct = pct
	return b
}

func (b *CueBuilder) ScreenText(text string) *CueBuilder {
	b.cue.ScreenText = text
	return b
}

func (b *CueBuilder) Tag(tag string) *CueBuilder {
	if tag != "" {
		b.cue.Tags = append(b.cue.Tags, tag)
	}
	return b
}

func (b *CueBuilder) Build() (ShowCue, error) {
	if b.err != nil {
		return ShowCue{}, b.err
	}
	if b.cue.Name == "" {
		return ShowCue{}, fmt.Errorf("cue name is required")
	}
	return b.cue, nil
}
