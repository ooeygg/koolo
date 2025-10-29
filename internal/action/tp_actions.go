// internal/action/tp_actions.go
package action

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func checkPlayerDeathForTP(ctx *context.Status) error {
	if ctx.Data.PlayerUnit.HPPercent() <= 0 {
		return health.ErrDied
	}
	// Player chicken check
	if ctx.Data.PlayerUnit.HPPercent() <= ctx.Data.CharacterCfg.Health.ChickenAt {
		return health.ErrChicken
	}
	// Mercenary chicken check
	if ctx.Data.MercHPPercent() > 0 && ctx.Data.MercHPPercent() <= ctx.Data.CharacterCfg.Health.MercChickenAt {
		return health.ErrMercChicken
	}
	return nil
}

func ReturnTown() error {
	ctx := context.Get()
	ctx.SetLastAction("ReturnTown")
	ctx.PauseIfNotPriority()

	// Proactive death check at the start of the action
	if err := checkPlayerDeathForTP(ctx); err != nil {
		return err
	}

	if ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	err := step.OpenPortal()
	if err != nil {
		// If opening portal fails, check if we died
		if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
			return errCheck
		}
		return err
	}
	portal, found := ctx.Data.Objects.FindOne(object.TownPortal)
	if !found {
		// If portal not found, check if we died
		if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
			return errCheck
		}
		return errors.New("portal not found")
	}

	initialInteractionErr := InteractObject(portal, func() bool {
		// Check for death during interaction callback
		if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
			return false // Returning false will stop the interaction loop, and the error will be caught outside
		}
		return ctx.Data.PlayerUnit.Area.IsTown()
	})

	if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
		return errCheck // Returning false will stop the interaction loop, and the error will be caught outside
	}

	if initialInteractionErr != nil {
		ctx.Logger.Debug("Initial portal interaction failed, attempting to clear area.", "error", initialInteractionErr)
		// If initial interaction fails, THEN clear the area
		if err = ClearAreaAroundPosition(portal.Position, 8, data.MonsterAnyFilter()); err != nil {
			ctx.Logger.Warn("Error clearing area around portal", "error", err)
			// Even if clearing area fails, check if we died during the process
			if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
				return errCheck
			}
		}

		// After (attempting to) clear, try to interact with the portal again
		err = InteractObject(portal, func() bool {
			// Check for death during interaction callback
			if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
				return false // Returning false will stop the interaction loop, and the error will be caught outside
			}
			return ctx.Data.PlayerUnit.Area.IsTown()
		})
		if err != nil {
			// If even after clearing, interaction fails, check for death and return error
			if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
				return errCheck
			}
			return err
		}
	}

	// Wait for area transition and data sync
	utils.Sleep(1000)
	ctx.RefreshGameData()

	// Wait for town area data to be fully loaded
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		// Check for death during the wait for town data
		if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
			return errCheck
		}

		if ctx.Data.PlayerUnit.Area.IsTown() {
			// Verify area data exists and is loaded
			if townData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
				if townData.IsInside(ctx.Data.PlayerUnit.Position) {
					return nil
				}
			}
		}
		utils.Sleep(100)
		ctx.RefreshGameData()
	}

	return fmt.Errorf("failed to verify town area data after portal transition")
}

func UsePortalInTown() error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalInTown")

	// Proactive death check at the start of the action
	if err := checkPlayerDeathForTP(ctx); err != nil {
		return err
	}

	tpArea := town.GetTownByArea(ctx.Data.PlayerUnit.Area).TPWaitingArea(*ctx.Data)
	_ = MoveToCoords(tpArea) // MoveToCoords already has death checks

	err := UsePortalFrom(ctx.Data.PlayerUnit.Name)
	if err != nil {
		// If using portal fails, check if we died
		if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
			return errCheck
		}
		return err
	}

	// Wait for area sync before attempting any movement
	utils.Sleep(500)
	ctx.RefreshGameData()
	// Check for death after refreshing game data
	if err := checkPlayerDeathForTP(ctx); err != nil {
		return err
	}

	if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		return err
	}

	// Ensure we're not in town
	if ctx.Data.PlayerUnit.Area.IsTown() {
		return fmt.Errorf("failed to leave town area")
	}

	// Move slightly away from portal
	moveSlightlyFromPortal()

	// Buff immediately after leaving town via portal
	ctx.Logger.Debug("Buffing immediately after leaving town via portal...")
	// Wait a moment for game state to fully sync after area transition
	utils.Sleep(500)
	ctx.RefreshGameData()
	// Force full buff including CTA Battle Orders
	ctx.LastBuffAt = time.Time{} // Reset to ensure full buff happens
	Buff()
	// Verify Battle Orders was applied if CTA is equipped
	utils.Sleep(300)
	ctx.RefreshGameData()
	ensureBattleOrdersApplied()

	// Perform item pickup after re-entering the portal
	err = ItemPickup(40)
	if err != nil {
		ctx.Logger.Warn("Error during item pickup after portal use", "error", err)
		// If item pickup fails, check if we died
		if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
			return errCheck
		}
	}

	return nil
}

// moveSlightlyFromPortal moves the character slightly away from the portal
func moveSlightlyFromPortal() {
	ctx := context.Get()
	ctx.SetLastAction("moveSlightlyFromPortal")

	// Get current position
	currentPos := ctx.Data.PlayerUnit.Position

	// Move 5-10 units in a random direction to avoid standing on portal
	offsetX := 8
	offsetY := 8

	targetPos := data.Position{
		X: currentPos.X + offsetX,
		Y: currentPos.Y + offsetY,
	}

	// Use MoveToCoords to move slightly
	_ = MoveToCoords(targetPos)
	utils.Sleep(300)
}

func UsePortalFrom(owner string) error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalFrom")

	// Proactive death check at the start of the action
	if err := checkPlayerDeathForTP(ctx); err != nil {
		return err
	}

	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	for _, obj := range ctx.Data.Objects {
		if obj.IsPortal() && obj.Owner == owner {
			return InteractObjectByID(obj.ID, func() bool {
				// Check for death during interaction callback
				if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
					return false // Returning false will stop the interaction loop, and the error will be caught outside
				}

				if !ctx.Data.PlayerUnit.Area.IsTown() {
					// Ensure area data is synced after portal transition
					utils.Sleep(500)
					ctx.RefreshGameData()
					// Check for death after refreshing game data
					if errCheck := checkPlayerDeathForTP(ctx); errCheck != nil {
						return false
					}

					if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
						return false
					}
					return true
				}
				return false
			})
		}
	}

	return errors.New("portal not found")
}
