package main

import (
	"fmt"
	"patterns/patterns"
	"time"
)

func main() {
	cue, err := patterns.NewCueBuilder("Intro → Chorus blast").
		StartsIn(2 * time.Second).
		BPM(138).
		LightPreset("NEON-SWEEP").
		SmokeLevelPct(35).
		ScreenText("ARE YOU READY?").
		Tag("opener").
		Tag("high-energy").
		Build()

	if err != nil {
		panic(err)
	}

	fmt.Println("BUILT CUE:", cue.Name, "| bpm:", cue.BPM, "| lights:", cue.LightPreset, "| smoke:", cue.SmokeLevelPct, "%")
	fmt.Println()

	led := patterns.NewLEDWall("A1")
	proj := patterns.NewProjector("Backstage")

	big := patterns.NewBigText(led, cue.ScreenText)
	lyrics := patterns.NewLyricsCard(proj, "Chorus", []string{
		"Hands up!",
		"Feel the bass!",
		"We go again!",
	})

	_ = big.Show()
	_ = big.ShowOn(proj)
	_ = lyrics.Show()
	_ = lyrics.ShowOn(led)

	hub := patterns.NewShowHub()

	artist := patterns.NewArtist("MC")
	lights := patterns.NewLightingRig("MainRig")
	smoke := patterns.NewSmokeMachine("Fogger-01")
	sound := patterns.NewSoundRack("Rack-Sub")

	hub.Register(artist)
	hub.Register(lights)
	hub.Register(smoke)
	hub.Register(sound)

	artist.Trigger(patterns.EventTalk, "Welcome to the show!")
	time.Sleep(500 * time.Millisecond)

	artist.Trigger(patterns.EventChorus, "Go!")
	time.Sleep(500 * time.Millisecond)

	hub.Send(smoke.Name(), patterns.Event{
		Type:    patterns.EventSmokeNow,
		From:    "HUB",
		Payload: fmt.Sprintf("%d%%", cue.SmokeLevelPct),
	})

	artist.Trigger(patterns.EventDrop, "DROP NOW!!!")
	time.Sleep(500 * time.Millisecond)
}
