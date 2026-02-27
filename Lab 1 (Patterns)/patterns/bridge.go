package patterns

import (
	"fmt"
	"strings"
)

type Display interface {
	Draw(frame string) error
	Name() string
}

type LEDWall struct{ id string }

func NewLEDWall(id string) *LEDWall { return &LEDWall{id: id} }
func (d *LEDWall) Name() string     { return "LED-WALL#" + d.id }
func (d *LEDWall) Draw(frame string) error {
	fmt.Printf("[%s]\n%s\n\n", d.Name(), frame)
	return nil
}

type Projector struct{ room string }

func NewProjector(room string) *Projector { return &Projector{room: room} }
func (d *Projector) Name() string         { return "PROJECTOR(" + d.room + ")" }
func (d *Projector) Draw(frame string) error {
	fmt.Printf("[%s]\n%s\n\n", d.Name(), frame)
	return nil
}

type Visual interface {
	Render() string
	ShowOn(Display) error
}

type BigText struct {
	display Display
	Text    string
}

func NewBigText(display Display, text string) *BigText {
	return &BigText{display: display, Text: text}
}

func (v *BigText) Render() string {
	line := strings.Repeat("=", len(v.Text)+8)
	return fmt.Sprintf("%s\n==  %s  ==\n%s", line, v.Text, line)
}

func (v *BigText) ShowOn(d Display) error {
	return d.Draw(v.Render())
}

func (v *BigText) Show() error {
	return v.display.Draw(v.Render())
}

type LyricsCard struct {
	display Display
	Title   string
	Lines   []string
}

func NewLyricsCard(display Display, title string, lines []string) *LyricsCard {
	return &LyricsCard{display: display, Title: title, Lines: lines}
}

func (v *LyricsCard) Render() string {
	var b strings.Builder
	b.WriteString("♪ " + v.Title + "\n")
	b.WriteString(strings.Repeat("-", len(v.Title)+2) + "\n")
	for _, ln := range v.Lines {
		b.WriteString(ln + "\n")
	}
	return b.String()
}

func (v *LyricsCard) ShowOn(d Display) error {
	return d.Draw(v.Render())
}

func (v *LyricsCard) Show() error {
	return v.display.Draw(v.Render())
}
