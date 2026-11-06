package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/shr-go/bili_live_tui/internal/tui"
	"github.com/shr-go/bili_live_tui/pkg/logging"
	"os"
)

func main() {
	logging.InitLogConfig()
	defer logging.Cleanup()
	if err := tui.LoadConfig("config.toml"); err != nil {
		logging.Fatalf("%v", err)
	}
	logging.Infof("tui start")
	client := tui.GetCustomHttpClient()
	room, err := tui.PrepareEnterRoom(client)
	if err != nil || room == nil {
		logging.Fatalf("Connect server error, err=%v", err)
	}
	m := tui.InitialModel(room)
	defer m.Close()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		// Fatal exits without running defers; stop audio and room workers first.
		m.Close()
		logging.Fatalf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
