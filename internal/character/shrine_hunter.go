package character

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/game"
)

// ShrineHunter is a minimal character class designed specifically for scanning shrines
// It doesn't need combat capabilities, just movement to waypoints
type ShrineHunter struct {
	BaseCharacter
}

// CheckKeyBindings - Shrine hunter only needs teleport/TP
func (s ShrineHunter) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.TomeOfTownPortal}
	missingKeybindings := make([]skill.ID, 0)

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	return missingKeybindings
}

// BuffSkills - No buffs needed for shrine scanning
func (s ShrineHunter) BuffSkills() []skill.ID {
	return []skill.ID{}
}

// PreCTABuffSkills - No pre-CTA buffs needed
func (s ShrineHunter) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

// KillMonsterSequence - Not used for shrine hunting, but required by interface
func (s ShrineHunter) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	return fmt.Errorf("shrine hunter does not support combat")
}

// Boss kill methods - Not implemented for shrine hunter
func (s ShrineHunter) KillCountess() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillAndariel() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillSummoner() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillDuriel() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillMephisto() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillPindle() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillNihlathak() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillCouncil() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillDiablo() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillIzual() error {
	return fmt.Errorf("shrine hunter does not support combat")
}

func (s ShrineHunter) KillBaal() error {
	return fmt.Errorf("shrine hunter does not support combat")
}
