package run

import (
	"fmt"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type ExpPit struct {
	ctx *context.Status
}

func NewExpPit() *ExpPit {
	return &ExpPit{
		ctx: context.Get(),
	}
}

func (e ExpPit) Name() string {
	return string(config.ExpPitRun)
}

func (e ExpPit) Run() error {
	ctx := context.Get()

	// Always use elite filter for this runmak
	monsterFilter := data.MonsterEliteFilter()

	ctx.Logger.Info("Starting Experience Pit run - checking shrines in Dark Wood")

	// Waypoint to Dark Wood
	if err := action.WayPoint(area.BlackMarsh); err != nil {
		return err
	}

	// Scan for experience shrines in Dark Wood
	e.scanAndActivateExpShrines()

	// Move to Tamoe Highland
	if err := action.MoveToArea(area.TamoeHighland); err != nil {
		return err
	}

	// Check for shrines in Tamoe Highland too
	e.scanAndActivateExpShrines()

	// Move to Pit Level 1
	if err := action.MoveToArea(area.PitLevel1); err != nil {
		return err
	}

	// Open a TP if we're the leader
	action.OpenTPIfLeader()

	ctx.Logger.Info("Clearing Pit Level 1 (elites and large groups only)")
	// Clear Pit Level 1 - only elites, no chest opening
	if err := action.ClearCurrentLevel(false, monsterFilter); err != nil {
		return err
	}

	// Move to Pit Level 2
	if err := action.MoveToArea(area.PitLevel2); err != nil {
		return err
	}

	ctx.Logger.Info("Clearing Pit Level 2 (elites and large groups only)")
	// Clear Pit Level 2 - only elites, no chest opening
	return action.ClearCurrentLevel(false, monsterFilter)
}

// scanAndActivateExpShrines scans the current area for experience shrines and activates the closest one if found
func (e ExpPit) scanAndActivateExpShrines() {
	ctx := context.Get()

	// Refresh game data to get current objects
	ctx.RefreshGameData()

	// Find all shrines in the area
	const maxShrinesToCheck = 5
	type shrineInfo struct {
		obj      data.Object
		distance float64
	}
	var shrinesInArea []shrineInfo

	for _, obj := range ctx.Data.Objects {
		if obj.IsShrine() && obj.Selectable {
			// Calculate distance from player
			dx := float64(ctx.Data.PlayerUnit.Position.X - obj.Position.X)
			dy := float64(ctx.Data.PlayerUnit.Position.Y - obj.Position.Y)
			distance := math.Sqrt(dx*dx + dy*dy)

			shrinesInArea = append(shrinesInArea, shrineInfo{
				obj:      obj,
				distance: distance,
			})
		}
	}

	if len(shrinesInArea) == 0 {
		ctx.Logger.Debug("No shrines found in this area")
		return
	}

	ctx.Logger.Info(fmt.Sprintf("Found %d shrine(s) in area, checking up to %d", len(shrinesInArea), maxShrinesToCheck))

	// Sort by distance (closest first)
	for i := 0; i < len(shrinesInArea); i++ {
		for j := i + 1; j < len(shrinesInArea); j++ {
			if shrinesInArea[j].distance < shrinesInArea[i].distance {
				shrinesInArea[i], shrinesInArea[j] = shrinesInArea[j], shrinesInArea[i]
			}
		}
	}

	// Check up to maxShrinesToCheck shrines
	shrineLimit := maxShrinesToCheck
	if len(shrinesInArea) < shrineLimit {
		shrineLimit = len(shrinesInArea)
	}

	for i := 0; i < shrineLimit; i++ {
		shrineInf := shrinesInArea[i]

		// Move closer to the shrine to identify it
		ctx.Logger.Debug(fmt.Sprintf("Moving to shrine %d/%d at distance %.0f", i+1, shrineLimit, shrineInf.distance))
		if err := action.MoveToCoords(shrineInf.obj.Position); err != nil {
			ctx.Logger.Debug(fmt.Sprintf("Failed to move to shrine: %v", err))
			continue
		}

		// Wait for shrine data to load
		utils.Sleep(300)

		// Refresh game data to get updated shrine information
		ctx.RefreshGameData()

		// Find the shrine again and check its type
		shrine, found := ctx.Data.Objects.FindByID(shrineInf.obj.ID)
		if !found {
			ctx.Logger.Debug("Shrine not found after moving closer")
			continue
		}

		// Check if it's an Experience Shrine
		if shrine.IsShrine() && shrine.Shrine.ShrineType == object.ExperienceShrine {
			ctx.Logger.Info(fmt.Sprintf("Experience Shrine found! Activating at position (%d, %d)", shrine.Position.X, shrine.Position.Y))

			// Activate the shrine
			if err := step.InteractObject(shrine, func() bool {
				// Check if we have the exp shrine buff
				ctx.RefreshGameData()
				return ctx.Data.PlayerUnit.States.HasState(state.ShrineExperience)
			}); err != nil {
				ctx.Logger.Warn(fmt.Sprintf("Failed to activate Experience Shrine: %v", err))
			} else {
				ctx.Logger.Info("Experience Shrine activated successfully!")
			}

			// Only activate one exp shrine per run
			return
		}
	}

	ctx.Logger.Debug("No Experience Shrines found in area")
}
