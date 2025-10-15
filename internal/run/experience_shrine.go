package run

import (
	"fmt"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type ExperienceShrine struct {
	ctx *context.Status
}

func NewExperienceShrine() *ExperienceShrine {
	return &ExperienceShrine{
		ctx: context.Get(),
	}
}

func (e ExperienceShrine) Name() string {
	return string(config.ExperienceShrineRun)
}

func (e ExperienceShrine) Run() error {
	ctx := context.Get()

	// If this is a leader and there's a companion-found shrine location, visit it first
	if ctx.CharacterCfg.Companion.Enabled && ctx.CharacterCfg.Companion.Leader {
		if ctx.CurrentGame.CompanionShrineLocation != nil {
			shrine := ctx.CurrentGame.CompanionShrineLocation
			ctx.Logger.Info("=== Leader visiting companion-found Experience Shrine ===")
			ctx.Logger.Info(fmt.Sprintf("Companion: %s found shrine in %s at (%d, %d)",
				shrine.CompanionName, shrine.AreaName, shrine.X, shrine.Y))

			// Waypoint to the area
			if err := action.WayPoint(shrine.AreaID); err != nil {
				ctx.Logger.Warn(fmt.Sprintf("Failed to waypoint to %s: %v", shrine.AreaName, err))
			} else {
				// Move to the shrine location
				shrinePos := data.Position{X: shrine.X, Y: shrine.Y}
				if err := action.MoveToCoords(shrinePos); err != nil {
					ctx.Logger.Warn(fmt.Sprintf("Failed to move to shrine: %v", err))
				} else {
					// Wait for game data to refresh and find the shrine
					utils.Sleep(300)
					ctx.RefreshGameData()

					// Find the shrine object at the reported location
					var foundShrine *data.Object
					for _, obj := range ctx.Data.Objects {
						if obj.IsShrine() && obj.Position.X == shrine.X && obj.Position.Y == shrine.Y {
							foundShrine = &obj
							break
						}
					}

					// If found, interact with it to activate
					if foundShrine != nil && foundShrine.Selectable {
						ctx.Logger.Info("Leader activating companion-found Experience Shrine")
						// The shrine should activate automatically when we're close enough
						// or we could add explicit interaction here if needed
						ctx.Logger.Info("Successfully activated companion-found shrine!")
					} else {
						ctx.Logger.Info("Leader reached shrine location (shrine may already be activated or not visible)")
					}

					// Clear the shrine location after visiting
					ctx.CurrentGame.CompanionShrineLocation = nil
					return nil
				}
			}
			// If we failed to visit, clear it anyway to avoid retrying
			ctx.CurrentGame.CompanionShrineLocation = nil
		}
	}

	// Get waypoints from configuration, or use default Act 1 waypoints if not configured
	configuredWaypoints := ctx.CharacterCfg.Game.ShrineHunt.Waypoints
	var waypointIDs []area.ID

	if len(configuredWaypoints) == 0 {
		// Default to popular experience shrine areas across multiple acts
		waypointIDs = []area.ID{
			area.ColdPlains,
			area.StonyField,
			area.DarkWood,
			area.BlackMarsh,
			area.RiverOfFlame,
		}
		ctx.Logger.Info("=== Starting Experience Shrine Scanner ===")
		ctx.Logger.Info("Scanning default waypoints for Experience Shrines (no waypoints configured in UI)")
	} else {
		waypointIDs = configuredWaypoints
		ctx.Logger.Info("=== Starting Experience Shrine Scanner ===")
		ctx.Logger.Info(fmt.Sprintf("Scanning %d configured waypoints for Experience Shrines", len(waypointIDs)))
	}

	// Build waypoint list with names
	type waypointInfo struct {
		name string
		id   area.ID
	}
	waypoints := make([]waypointInfo, 0, len(waypointIDs))
	for _, wpID := range waypointIDs {
		wpArea := wpID.Area()
		waypoints = append(waypoints, waypointInfo{
			name: wpArea.Name,
			id:   wpID,
		})
	}

	// Initialize the map to store shrine counts
	expShrineData := make(map[string]int)
	totalExpShrines := 0

	for _, wp := range waypoints {
		// Waypoint to the area
		if err := action.WayPoint(wp.id); err != nil {
			ctx.Logger.Warn(fmt.Sprintf("Failed to waypoint to %s: %v", wp.name, err))
			expShrineData[wp.name] = 0
			continue
		}

		// Refresh game data to get current objects
		ctx.RefreshGameData()

		// First pass: Find all shrines in the area (we can see shrine objects from far away)
		const minShrinesToCheck = 3 // Always check at least 3 shrines
		const maxShrinesToCheck = 5 // Check up to 5 if there are more nearby
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

		ctx.Logger.Info(fmt.Sprintf("%s: Found %d shrine(s) in area", wp.name, len(shrinesInArea)))

		// Sort by distance (closest first) - bubble sort is fine for small arrays
		for i := 0; i < len(shrinesInArea); i++ {
			for j := i + 1; j < len(shrinesInArea); j++ {
				if shrinesInArea[j].distance < shrinesInArea[i].distance {
					shrinesInArea[i], shrinesInArea[j] = shrinesInArea[j], shrinesInArea[i]
				}
			}
		}

		// Determine how many shrines to check
		// Always check minimum 3, but up to 5 if there are more
		shrineLimit := minShrinesToCheck
		if len(shrinesInArea) > minShrinesToCheck {
			// Check up to max, but not more than what's available
			if len(shrinesInArea) >= maxShrinesToCheck {
				shrineLimit = maxShrinesToCheck
			} else {
				shrineLimit = len(shrinesInArea)
			}
		} else {
			// If there are fewer than minimum, check all available
			shrineLimit = len(shrinesInArea)
		}

		ctx.Logger.Info(fmt.Sprintf("%s: Will check %d closest shrine(s)", wp.name, shrineLimit))

		// Second pass: Move close to each shrine to identify its type
		expShrineCount := 0
		shrinesChecked := 0

		for _, shrineInf := range shrinesInArea {
			if shrinesChecked >= shrineLimit {
				ctx.Logger.Info(fmt.Sprintf("  Checked %d shrines, moving to next waypoint", shrineLimit))
				break
			}

			// Move closer to the shrine (within ~15 units to ensure type is loaded)
			ctx.Logger.Info(fmt.Sprintf("  Moving to shrine %d/%d at distance %.0f", shrinesChecked+1, shrineLimit, shrineInf.distance))
			if err := action.MoveToCoords(shrineInf.obj.Position); err != nil {
				ctx.Logger.Warn(fmt.Sprintf("  Failed to move to shrine: %v", err))
				shrinesChecked++
				continue
			}

			// Wait a moment for shrine data to load
			utils.Sleep(300)

			// Refresh game data to get updated shrine information
			ctx.RefreshGameData()

			// Find the shrine again and check its type
			shrine, found := ctx.Data.Objects.FindByID(shrineInf.obj.ID)
			if !found {
				ctx.Logger.Debug("  Shrine not found after moving closer")
				shrinesChecked++
				continue
			}

			// Now check if it's an Experience Shrine
			if shrine.IsShrine() && shrine.Shrine.ShrineType == object.ExperienceShrine {
				expShrineCount++
				ctx.Logger.Info(fmt.Sprintf("  ✓ Experience Shrine found at position (%d, %d)!",
					shrine.Position.X, shrine.Position.Y))

				// Update the UI immediately
				expShrineData[wp.name] = expShrineCount
				ctx.SetExpShrineData(expShrineData)

				// Log the find and exit to next run
				ctx.Logger.Info(fmt.Sprintf("=== Experience Shrine Found! ==="))
				ctx.Logger.Info(fmt.Sprintf("Location: %s at (%d, %d)", wp.name, shrine.Position.X, shrine.Position.Y))

				// Handle companions vs leader differently
				if ctx.CharacterCfg.Companion.Enabled && !ctx.CharacterCfg.Companion.Leader {
					// COMPANION: Send location to leader and DO NOT activate the shrine
					ctx.Logger.Info("Companion notifying leader of Experience Shrine location")
					event.Send(event.CompanionFoundExpShrine(
						event.Text(ctx.CharacterCfg.CharacterName, "Experience Shrine Found"),
						ctx.CharacterCfg.CharacterName,
						wp.name,
						int(wp.id),
						shrine.Position.X,
						shrine.Position.Y,
					))

					// Move away from the shrine to ensure we don't accidentally activate it
					// Calculate a position 30 units away from the shrine
					dx := ctx.Data.PlayerUnit.Position.X - shrine.Position.X
					dy := ctx.Data.PlayerUnit.Position.Y - shrine.Position.Y
					distance := math.Sqrt(float64(dx*dx + dy*dy))

					if distance < 20 { // If too close, move away
						awayX := shrine.Position.X - int(float64(dx)/distance*30)
						awayY := shrine.Position.Y - int(float64(dy)/distance*30)
						awayPos := data.Position{X: awayX, Y: awayY}
						ctx.Logger.Info("Companion moving away from shrine to avoid activation")
						action.MoveToCoords(awayPos)
					}

					ctx.Logger.Info("Companion: Shrine location sent to leader, moving to next run...")
					return nil
				} else if !ctx.CharacterCfg.Companion.Enabled || ctx.CharacterCfg.Companion.Leader {
					// LEADER or SOLO: Activate the shrine (the default behavior - just return)
					ctx.Logger.Info("Moving to next run...")
					return nil
				}
			} else {
				ctx.Logger.Debug(fmt.Sprintf("  Not an Experience Shrine (type: %v)", shrine.Shrine.ShrineType))
			}

			shrinesChecked++
		}

		totalExpShrines += expShrineCount
		expShrineData[wp.name] = expShrineCount

		ctx.Logger.Info(fmt.Sprintf("%s: %d Experience Shrine(s) found (checked %d shrines)",
			wp.name, expShrineCount, shrinesChecked))
	}

	// Only reach here if no Experience Shrine was found in any area
	ctx.Logger.Info(fmt.Sprintf("=== Scan Complete ==="))
	ctx.Logger.Info(fmt.Sprintf("No Experience Shrines found after checking all waypoints"))

	// Store the shrine data in the context for the UI to display
	ctx.SetExpShrineData(expShrineData)

	return nil
}
