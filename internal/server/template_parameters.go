package server

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/bot"
	"github.com/hectorgimenez/koolo/internal/config"
)

type IndexData struct {
	ErrorMessage string
	Version      string
	Status       map[string]bot.Stats
	DropCount    map[string]int
}

type DropData struct {
	NumberOfDrops int
	Character     string
	Drops         []data.Drop
}

type CharacterSettings struct {
	ErrorMessage string
	Supervisor   string
	Config       *config.CharacterCfg
	DayNames     []string
	EnabledRuns  []string
	DisabledRuns []string
	AvailableTZs map[int]string
	RecipeList   []string
	GameSettings  GameSettingsData
}

type GameSettingsData struct {
    IsLeader      bool
    FollowLeader  bool
    LeaderName    string
    JoinDelay     int
    GamePassword  string
    GameNames     []string    // List of recent game names for reference
    LastGameName  string      // Last game created/joined
}

type ConfigData struct {
	ErrorMessage string
	*config.KooloCfg
}

type AutoSettings struct {
	ErrorMessage string
}
