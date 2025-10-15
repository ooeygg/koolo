package action

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

// FindExperienceShrineAct1 quickly searches for an Experience Shrine near waypoints
func FindExperienceShrineAct1() error {
	ctx := context.Get()
	ctx.SetLastAction("FindExperienceShrineAct1")

	// Get waypoints from configuration, or use default Act 1 waypoints if not configured
	areasToSearch := ctx.CharacterCfg.Game.ShrineHunt.Waypoints
	if len(areasToSearch) == 0 {
		// Default to Act 1 waypoints if not configured
		areasToSearch = []area.ID{area.ColdPlains, area.StonyField, area.DarkWood, area.BlackMarsh}
		ctx.Logger.Info("Starting quick Experience Shrine hunt near Act 1 waypoints (default)")
	} else {
		ctx.Logger.Info(fmt.Sprintf("Starting quick Experience Shrine hunt near %d configured waypoints", len(areasToSearch)))
	}

	for _, searchArea := range areasToSearch {
		ctx.Logger.Debug(fmt.Sprintf("Checking for Experience Shrine near %s waypoint", searchArea.Area().Name))

		// Move to the area
		if err := WayPoint(searchArea); err != nil {
			ctx.Logger.Warn(fmt.Sprintf("Failed to waypoint to %s: %v", searchArea.Area().Name, err), slog.Any("error", err))
			continue
		}

		// Quick search around waypoint
		shrineFound, err := quickSearchAroundWaypoint(searchArea)
		if err != nil {
			ctx.Logger.Warn(fmt.Sprintf("Error while searching %s: %v", searchArea.Area().Name, err), slog.Any("error", err))
			continue
		}

		if shrineFound {
			ctx.Logger.Info(fmt.Sprintf("Experience Shrine found and activated near %s waypoint!", searchArea.Area().Name))
			// Return to town after finding the shrine
			if err := ReturnTown(); err != nil {
				ctx.Logger.Warn("Failed to return to town after shrine activation", slog.Any("error", err))
			}
			return nil
		}

		ctx.Logger.Debug(fmt.Sprintf("No Experience Shrine found near %s waypoint", searchArea.Area().Name))
	}

	ctx.Logger.Info("No Experience Shrine found near any waypoints, returning to town")

	// Return to town even if no shrine was found
	if err := ReturnTown(); err != nil {
		ctx.Logger.Warn("Failed to return to town after shrine search", slog.Any("error", err))
	}

	return nil
}

// quickSearchAroundWaypoint does a fast search around the waypoint for shrines
func quickSearchAroundWaypoint(searchArea area.ID) (bool, error) {
	ctx := context.Get()
	ctx.SetLastAction("quickSearchAroundWaypoint")

	// Verify we're in the correct area
	if ctx.Data.PlayerUnit.Area != searchArea {
		return false, fmt.Errorf("not in expected area %s, currently in %s", searchArea.Area().Name, ctx.Data.PlayerUnit.Area.Area().Name)
	}

	const searchRadius = 50 // Search within 50 units of waypoint
	const maxShrinesToCheck = 3 // Move to next waypoint after checking 3 shrines
	startPosition := ctx.Data.PlayerUnit.Position
	shrinesChecked := 0

	ctx.Logger.Debug(fmt.Sprintf("Quick searching %d units around waypoint in %s", searchRadius, searchArea.Area().Name))

	// Refresh game data to get current object list
	ctx.RefreshGameData()

	// Find all shrines near the waypoint
	type shrineDistance struct {
		shrine   data.Object
		distance float64
	}
	var nearbyShrines []shrineDistance

	for _, obj := range ctx.Data.Objects {
		if obj.IsShrine() && obj.Selectable {
			distance := calculateDistance(startPosition, obj.Position)
			if distance <= searchRadius {
				nearbyShrines = append(nearbyShrines, shrineDistance{shrine: obj, distance: distance})
				ctx.Logger.Debug(fmt.Sprintf("Found shrine type %v at distance %.0f", obj.Shrine.ShrineType, distance))
			}
		}
	}

	ctx.Logger.Debug(fmt.Sprintf("Found %d shrines within %d units of waypoint", len(nearbyShrines), searchRadius))

	// Sort shrines by distance (closest first) - simple bubble sort
	for i := 0; i < len(nearbyShrines); i++ {
		for j := i + 1; j < len(nearbyShrines); j++ {
			if nearbyShrines[j].distance < nearbyShrines[i].distance {
				nearbyShrines[i], nearbyShrines[j] = nearbyShrines[j], nearbyShrines[i]
			}
		}
	}

	// Check each shrine, stopping after maxShrinesToCheck
	for _, sd := range nearbyShrines {
		ctx.PauseIfNotPriority()

		// Stop after checking 3 shrines
		if shrinesChecked >= maxShrinesToCheck {
			ctx.Logger.Debug(fmt.Sprintf("Checked %d shrines, moving to next waypoint", maxShrinesToCheck))
			break
		}

		shrinesChecked++

		// If it's an Experience Shrine, activate it
		if sd.shrine.Shrine.ShrineType == object.ExperienceShrine {
			ctx.Logger.Info(fmt.Sprintf("Experience Shrine found at position %v (distance: %.0f)!", sd.shrine.Position, sd.distance))

			// Move to the shrine
			if err := step.MoveTo(sd.shrine.Position); err != nil {
				ctx.Logger.Warn("Failed to move to Experience Shrine", slog.Any("error", err))
				continue
			}

			// Interact with the shrine
			if err := interactWithExperienceShrine(&sd.shrine); err != nil {
				ctx.Logger.Warn("Failed to interact with Experience Shrine", slog.Any("error", err))
				continue
			}

			return true, nil
		}

		ctx.Logger.Debug(fmt.Sprintf("Shrine %d/%d is type %v, not Experience Shrine", shrinesChecked, maxShrinesToCheck, sd.shrine.Shrine.ShrineType))
	}

	return false, nil
}

// calculateDistance returns the Euclidean distance between two positions
func calculateDistance(p1, p2 data.Position) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// interactWithExperienceShrine interacts with the Experience Shrine
func interactWithExperienceShrine(shrine *data.Object) error {
	ctx := context.Get()
	ctx.Logger.Info(fmt.Sprintf("Attempting to activate Experience Shrine at %v", shrine.Position))

	attempts := 0
	maxAttempts := 5

	for attempts < maxAttempts {
		ctx.RefreshGameData()

		s, found := ctx.Data.Objects.FindByID(shrine.ID)
		if !found || !s.Selectable {
			ctx.Logger.Info("Experience Shrine successfully activated!")
			return nil
		}

		// Calculate distance to shrine
		dx := float64(ctx.Data.PlayerUnit.Position.X - s.Position.X)
		dy := float64(ctx.Data.PlayerUnit.Position.Y - s.Position.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		// If too far, move closer
		if distance > 5 {
			ctx.Logger.Debug(fmt.Sprintf("Too far from shrine (distance: %.2f), moving closer", distance))
			if err := step.MoveTo(s.Position); err != nil {
				ctx.Logger.Warn("Failed to move closer to shrine", slog.Any("error", err))
			}
			attempts++
			continue
		}

		// Click on the shrine
		x, y := ui.GameCoordsToScreenCords(s.Position.X, s.Position.Y)
		ctx.HID.Click(game.LeftButton, x, y)

		attempts++
		utils.Sleep(200)
	}

	return fmt.Errorf("failed to activate Experience Shrine after %d attempts", maxAttempts)
}
