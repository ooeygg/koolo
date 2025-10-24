package action

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func BuffIfRequired() {
	ctx := context.Get()

	// Special handling for Fade in town (Assassin only)
	if ctx.Data.PlayerUnit.Area.IsTown() {
		// Check if Fade is needed and we're an assassin with Fade skill
		if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Fade); found {
			// Only cast Fade if it's missing and at least 60 seconds have passed
			if !ctx.Data.PlayerUnit.States.HasState(state.Fade) && time.Since(ctx.LastBuffAt) >= time.Second*60 {
				ctx.Logger.Debug("Casting Fade in town...")
				castFadeInTown()
				return
			}
		}
		return
	}

	if !IsRebuffRequired() {
		return
	}

	// Don't buff if we have 2 or more monsters close to the character.
	// Don't merge with the previous if, because we want to avoid this expensive check if we don't need to buff
	closeMonsters := 0
	for _, m := range ctx.Data.Monsters {
		if ctx.PathFinder.DistanceFromMe(m.Position) < 15 {
			closeMonsters++
		}
		// cheaper to check here and end function if say first 2 already < 15
		// so no need to compute the rest
		if closeMonsters >= 2 {
			return
		}
	}

	Buff()
}

// castFadeInTown casts Fade specifically in town for Assassins
func castFadeInTown() {
	ctx := context.Get()
	ctx.SetLastAction("castFadeInTown")

	kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Fade)
	if !found {
		return
	}

	utils.Sleep(100)
	ctx.HID.PressKeyBinding(kb)
	utils.Sleep(180)
	ctx.HID.Click(game.RightButton, 640, 340)
	utils.Sleep(100)
	ctx.LastBuffAt = time.Now()
}

func Buff() {
	ctx := context.Get()
	ctx.SetLastAction("Buff")

	// Reduce cooldown to 30 seconds to prevent double buffing from network lag
	// Actual rebuffing should happen every 2-3 minutes based on IsRebuffRequired
	if ctx.Data.PlayerUnit.Area.IsTown() || time.Since(ctx.LastBuffAt) < time.Second*30 {
		return
	}

	// Check if we're in loading screen
	if ctx.Data.OpenMenus.LoadingScreen {
		ctx.Logger.Debug("Loading screen detected. Waiting for game to load before buffing...")
		ctx.WaitForGameToLoad()

		// Give it half a second more
		utils.Sleep(500)
	}

	ctx.Logger.Debug("Starting full buff sequence...")

	preKeys := make([]data.KeyBinding, 0)
	for _, buff := range ctx.Char.PreCTABuffSkills() {
		kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff)
		if !found {
			ctx.Logger.Info("Key binding not found, skipping buff", slog.String("skill", buff.Desc().Name))
		} else {
			preKeys = append(preKeys, kb)
		}
	}

	if len(preKeys) > 0 {
		ctx.Logger.Debug("PRE CTA Buffing...")
		for _, kb := range preKeys {
			utils.Sleep(100)
			ctx.HID.PressKeyBinding(kb)
			utils.Sleep(180)
			ctx.HID.Click(game.RightButton, 640, 340)
			utils.Sleep(100)
		}
	}

	// Always call buffCTA to ensure Battle Orders is cast
	buffCTA()

	postKeys := make([]data.KeyBinding, 0)
	for _, buff := range ctx.Char.BuffSkills() {
		kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff)
		if !found {
			ctx.Logger.Info("Key binding not found, skipping buff", slog.String("skill", buff.Desc().Name))
		} else {
			postKeys = append(postKeys, kb)
		}
	}

	if len(postKeys) > 0 {
		ctx.Logger.Debug("Post CTA Buffing...")

		for _, kb := range postKeys {
			utils.Sleep(100)
			ctx.HID.PressKeyBinding(kb)
			utils.Sleep(180)
			ctx.HID.Click(game.RightButton, 640, 340)
			utils.Sleep(100)
		}
	}

	// IMPORTANT: Always update LastBuffAt after buffing, regardless of whether there were post-CTA buffs
	ctx.LastBuffAt = time.Now()
	ctx.Logger.Debug("Buff sequence completed")
}

func IsRebuffRequired() bool {
	ctx := context.Get()
	ctx.SetLastAction("IsRebuffRequired")

	// Don't buff if we are in town, or we did it recently (it prevents double buffing because of network lag)
	if ctx.Data.PlayerUnit.Area.IsTown() || time.Since(ctx.LastBuffAt) < time.Second*30 {
		return false
	}

	// Rebuff every 2-3 minutes (150 seconds = 2.5 minutes) for periodic maintenance
	if time.Since(ctx.LastBuffAt) >= time.Second*150 {
		ctx.Logger.Debug("Periodic rebuff needed (2.5 minutes elapsed)")
		return true
	}

	if ctaFound(*ctx.Data) && (!ctx.Data.PlayerUnit.States.HasState(state.Battleorders) || !ctx.Data.PlayerUnit.States.HasState(state.Battlecommand)) {
		return true
	}

	// TODO: Find a better way to convert skill to state
	buffs := ctx.Char.BuffSkills()
	for _, buff := range buffs {
		if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff); found {
			if buff == skill.HolyShield && !ctx.Data.PlayerUnit.States.HasState(state.Holyshield) {
				return true
			}
			if buff == skill.FrozenArmor && (!ctx.Data.PlayerUnit.States.HasState(state.Frozenarmor) && !ctx.Data.PlayerUnit.States.HasState(state.Shiverarmor) && !ctx.Data.PlayerUnit.States.HasState(state.Chillingarmor)) {
				return true
			}
			if buff == skill.EnergyShield && !ctx.Data.PlayerUnit.States.HasState(state.Energyshield) {
				return true
			}
			if buff == skill.CycloneArmor && !ctx.Data.PlayerUnit.States.HasState(state.Cyclonearmor) {
				return true
			}
			if buff == skill.BurstOfSpeed && !ctx.Data.PlayerUnit.States.HasState(state.Quickness) {
				return true
			}
			if buff == skill.Fade && !ctx.Data.PlayerUnit.States.HasState(state.Fade) {
				return true
			}
			if buff == skill.BoneArmor && !ctx.Data.PlayerUnit.States.HasState(state.Bonearmor) {
				return true
			}
		}
	}

	return false
}

func buffCTA() {
	ctx := context.Get()
	ctx.SetLastAction("buffCTA")

	if !ctaFound(*ctx.Data) {
		ctx.Logger.Debug("No CTA found in equipment, skipping Battle Command/Orders")
		return
	}

	ctx.Logger.Info("CTA found: swapping weapon and casting Battle Command / Battle Orders")

	// Swap weapon only in case we don't have the CTA, sometimes CTA is already equipped
	if _, found := ctx.Data.PlayerUnit.Skills[skill.BattleCommand]; !found {
		ctx.Logger.Debug("Swapping to CTA weapon...")
		step.SwapToCTA()
		utils.Sleep(300) // Extra wait after weapon swap
		ctx.RefreshGameData()
	} else {
		ctx.Logger.Debug("CTA already equipped")
	}

	// Verify we have the skills before casting
	if _, found := ctx.Data.PlayerUnit.Skills[skill.BattleCommand]; !found {
		ctx.Logger.Error("Battle Command skill not available after CTA swap!")
		return
	}
	if _, found := ctx.Data.PlayerUnit.Skills[skill.BattleOrders]; !found {
		ctx.Logger.Error("Battle Orders skill not available after CTA swap!")
		return
	}

	ctx.Logger.Debug("Casting Battle Command 3 times...")
	// Cast Battle Command 3 times for maximum duration/effectiveness
	for i := 0; i < 3; i++ {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.BattleCommand))
		utils.Sleep(180)
		ctx.HID.Click(game.RightButton, 300, 300)
		utils.Sleep(100)
	}

	ctx.Logger.Debug("Casting Battle Orders 3 times...")
	// Cast Battle Orders 3 times for maximum effectiveness
	for i := 0; i < 3; i++ {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.BattleOrders))
		utils.Sleep(180)
		ctx.HID.Click(game.RightButton, 300, 300)
		utils.Sleep(100)
	}

	utils.Sleep(500)
	ctx.Logger.Debug("Swapping back to main weapon...")
	step.SwapToMainWeapon()
	utils.Sleep(300) // Extra wait after weapon swap
	ctx.RefreshGameData()

	ctx.Logger.Info("CTA buffing completed")
}

// ensureBattleOrdersApplied checks if Battle Orders is missing when CTA is equipped and recasts if needed
func ensureBattleOrdersApplied() {
	ctx := context.Get()
	ctx.SetLastAction("ensureBattleOrdersApplied")

	if !ctaFound(*ctx.Data) {
		ctx.Logger.Debug("No CTA equipped, skipping Battle Orders verification")
		return
	}

	hasBattleOrders := ctx.Data.PlayerUnit.States.HasState(state.Battleorders)
	hasBattleCommand := ctx.Data.PlayerUnit.States.HasState(state.Battlecommand)

	ctx.Logger.Info("Battle Orders verification",
		slog.Bool("hasBattleOrders", hasBattleOrders),
		slog.Bool("hasBattleCommand", hasBattleCommand))

	// Check if Battle Orders or Battle Command is missing
	if !hasBattleOrders || !hasBattleCommand {
		ctx.Logger.Warn("Battle Orders/Command missing after buff, recasting CTA...")

		// Reset LastBuffAt to allow immediate rebuff
		ctx.LastBuffAt = time.Time{}

		buffCTA()
		utils.Sleep(500)
		ctx.RefreshGameData()

		// Final check and log
		hasBattleOrders = ctx.Data.PlayerUnit.States.HasState(state.Battleorders)
		hasBattleCommand = ctx.Data.PlayerUnit.States.HasState(state.Battlecommand)

		if !hasBattleOrders || !hasBattleCommand {
			ctx.Logger.Error("Battle Orders/Command still missing after retry!",
				slog.Bool("hasBattleOrders", hasBattleOrders),
				slog.Bool("hasBattleCommand", hasBattleCommand))
		} else {
			ctx.Logger.Info("Battle Orders/Command successfully applied on retry")
		}
	} else {
		ctx.Logger.Debug("Battle Orders/Command verification passed")
	}
}

func ctaFound(d game.Data) bool {
	for _, itm := range d.Inventory.ByLocation(item.LocationEquipped) {
		_, boFound := itm.FindStat(stat.NonClassSkill, int(skill.BattleOrders))
		_, bcFound := itm.FindStat(stat.NonClassSkill, int(skill.BattleCommand))

		if boFound && bcFound {
			return true
		}
	}

	return false
}
