# Companion System Fixes and Implementation Guide

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Root Cause Analysis](#root-cause-analysis)
3. [Critical Fix: Register Event Handler](#critical-fix-register-event-handler)
4. [Quick Workaround Solutions](#quick-workaround-solutions)
5. [Implementation Plan: Attack Synchronization](#implementation-plan-attack-synchronization)
6. [Implementation Plan: Movement Following](#implementation-plan-movement-following)
7. [Configuration Improvements](#configuration-improvements)
8. [Testing Strategy](#testing-strategy)
9. [Troubleshooting Guide](#troubleshooting-guide)
10. [Questions for You](#questions-for-you)

---

## Executive Summary

The companion system is **currently broken** due to one critical missing piece: the `CompanionEventHandler` is created but never registered with the event system. This means companions never receive the events that tell them which games to join.

**Severity Levels:**
- **CRITICAL** (System doesn't work at all): Event handler not registered
- **HIGH** (Planned features don't work): Attack sync and movement following not implemented
- **MEDIUM** (Works but could be better): Hard-coded constants, race conditions
- **LOW** (Nice to have): Advanced features like formation control, buff coordination

**Quick Answer:** The minimum fix to get companions joining games is a **single line of code** to register the event handler.

**This Document Provides:**
1. The critical fix to make the system work (5 minutes)
2. Workarounds if you can't modify code
3. Complete implementation plans for missing features
4. Configuration improvements
5. Testing procedures

---

## Root Cause Analysis

### Critical Issue: Event Handler Never Registered

**The Problem:**

The `CompanionEventHandler` exists and is fully implemented, but it's never registered to receive events.

**Evidence:**

**File**: `internal/bot/companion.go` (lines 1-61)
- Handler exists with full implementation
- Handles `RequestCompanionJoinGameEvent` and `ResetCompanionGameInfoEvent`
- Code is correct and functional

**File**: `internal/bot/manager.go` (line 275)
- Only `statsHandler` is registered: `mng.eventListener.Register(statsHandler.Handle)`
- **No registration for companion handler**

**File**: `cmd/koolo/main.go` (lines 80-150)
- Registers: dropWriter, discordBot, telegramBot
- **No registration for companion handler**

**What This Means:**

```
Leader sends event → Event system broadcasts → Nobody is listening →
Companion never receives event → CompanionGameName never set →
HandleCompanionMenuFlow() does nothing → Companion never joins
```

**Why This Happened:**

Looking at the code structure, it appears the companion handler was created but the registration step was forgotten during development. The handler needs to be created per-supervisor (since each character needs its own handler with its own config), but there's no code path that creates and registers it.

### Secondary Issues

**Issue 2: Attack Synchronization Not Implemented**
- Event defined but never sent
- No handler for attack events
- Config setting has no effect

**Issue 3: Movement Following Not Implemented**
- No position broadcasting
- No pathfinding override
- Only checkpoint-based waiting exists

**Issue 4: Race Conditions**
- Companion may try to join before game is fully created
- Retry logic helps but isn't perfect

**Issue 5: Hard-coded Constants**
- 20-unit wait distance can't be configured
- 30-unit clear radius can't be configured
- 5-unit clear radius can't be configured

---

## Critical Fix: Register Event Handler

This is the **minimum fix** required to make companions work.

### Solution 1: Register Handler in Manager (RECOMMENDED)

This is the cleanest solution - register the handler when each supervisor is created.

**File to modify**: `internal/bot/manager.go`
**Location**: Around line 275 (after `statsHandler` registration)

**Current code** (lines 274-276):
```go
statsHandler := NewStatsHandler(supervisorName, logger)
mng.eventListener.Register(statsHandler.Handle)
supervisor, err := NewSinglePlayerSupervisor(supervisorName, bot, statsHandler)
```

**Add after line 275**:
```go
statsHandler := NewStatsHandler(supervisorName, logger)
mng.eventListener.Register(statsHandler.Handle)

// Register companion event handler
companionHandler := NewCompanionEventHandler(supervisorName, logger, ctx.CharacterCfg)
mng.eventListener.Register(companionHandler.Handle)

supervisor, err := NewSinglePlayerSupervisor(supervisorName, bot, statsHandler)
```

**Why this location?**
- Each supervisor gets its own companion handler
- Handler has access to correct character config
- Registered before supervisor starts
- Clean, minimal change

**What this fixes:**
- Companions will now receive join game events
- Companions will receive reset game info events
- Automatic game joining will work

**What this doesn't fix:**
- Attack synchronization (not implemented)
- Movement following (not implemented)
- Other planned features

### Solution 2: Register Handler in Supervisor Constructor (ALTERNATIVE)

If Solution 1 doesn't work for your setup, you can register in the supervisor itself.

**File to modify**: `internal/bot/single_supervisor.go`
**Location**: In the `NewSinglePlayerSupervisor` function (around line 46)

**Current code** (lines 46-55):
```go
func NewSinglePlayerSupervisor(name string, bot *Bot, statsHandler *StatsHandler) (*SinglePlayerSupervisor, error) {
	bs, err := newBaseSupervisor(bot, name, statsHandler)
	if err != nil {
		return nil, err
	}

	return &SinglePlayerSupervisor{
		baseSupervisor: bs,
	}, nil
}
```

**Problem**: You don't have access to `eventListener` here. You'd need to pass it as a parameter.

**Modified function signature**:
```go
func NewSinglePlayerSupervisor(name string, bot *Bot, statsHandler *StatsHandler, eventListener *event.Listener) (*SinglePlayerSupervisor, error) {
```

**Modified function body**:
```go
func NewSinglePlayerSupervisor(name string, bot *Bot, statsHandler *StatsHandler, eventListener *event.Listener) (*SinglePlayerSupervisor, error) {
	bs, err := newBaseSupervisor(bot, name, statsHandler)
	if err != nil {
		return nil, err
	}

	// Register companion event handler
	companionHandler := NewCompanionEventHandler(name, bot.ctx.Logger, bot.ctx.CharacterCfg)
	eventListener.Register(companionHandler.Handle)

	return &SinglePlayerSupervisor{
		baseSupervisor: bs,
	}, nil
}
```

**Then update the call in manager.go** (line 276):
```go
// Old
supervisor, err := NewSinglePlayerSupervisor(supervisorName, bot, statsHandler)

// New
supervisor, err := NewSinglePlayerSupervisor(supervisorName, bot, statsHandler, mng.eventListener)
```

**Why this is more complex:**
- Requires changing function signature
- Need to pass eventListener through
- More places to modify
- But it's more encapsulated

### Solution 3: Global Registration (NOT RECOMMENDED)

You could register a single global handler, but this won't work correctly because each character needs its own handler with its own config reference.

**Don't do this** - it will cause config conflicts between characters.

### Verification Steps

After implementing the fix:

**1. Add logging to verify registration**

In `internal/bot/manager.go` after registration:
```go
companionHandler := NewCompanionEventHandler(supervisorName, logger, ctx.CharacterCfg)
mng.eventListener.Register(companionHandler.Handle)
logger.Info("Companion event handler registered", slog.String("supervisor", supervisorName))
```

**2. Check logs on startup**

You should see:
```
[INFO] Companion event handler registered supervisor=LeaderChar
[INFO] Companion event handler registered supervisor=CompanionChar
```

**3. Test event flow**

Leader creates game, check companion logs for:
```
[INFO] Companion join game event received supervisor=CompanionChar leader=LeaderChar name=game-1 password=xxx
```

**4. Verify game joining**

Companion should automatically join leader's game within a few seconds.

### Estimated Time

**Code change**: 5 minutes
**Testing**: 15 minutes
**Total**: 20 minutes

---

## Quick Workaround Solutions

If you **cannot** modify the code, here are workarounds:

### Workaround 1: Manual Game Coordination

**Setup:**
1. Configure leader with specific game name template
2. Configure companion to NOT use companion mode
3. Manually create/join games with matching names

**Leader config**:
```yaml
companion:
  enabled: false  # Don't use companion mode

game:
  createGameAtMenu: true
  gameName: 'myrun-1'  # Manually set game name
```

**Companion config**:
```yaml
companion:
  enabled: false  # Don't use companion mode

game:
  createGameAtMenu: false
  gameName: 'myrun-1'  # Same game name
```

**Process:**
1. Set both to same game name
2. Start leader first
3. Start companion 10 seconds later
4. Both join same game
5. Change game names manually for next run

**Pros:**
- No code changes needed
- Works immediately

**Cons:**
- Manual coordination required
- No automatic game name incrementing
- Must update configs between runs
- Very tedious

### Workaround 2: Script-Based Coordination

**Setup:**
1. Write external script to monitor leader's status
2. Script updates companion config when leader creates game
3. Companion picks up new config

**Script logic** (pseudocode):
```
1. Monitor leader's log file
2. When leader creates game (detect from logs):
   - Parse game name and password
   - Update companion's config.yaml
   - Trigger companion config reload
3. Wait for game to finish
4. Clear companion config
5. Repeat
```

**Pros:**
- Somewhat automated
- No Go code changes

**Cons:**
- Requires external script
- Depends on log parsing
- Fragile (log format changes break it)
- Config reload may not work properly
- Complex to set up

### Workaround 3: Single-Game Mode

**Setup:**
1. Configure both characters with same static game name
2. Run single game repeatedly
3. Never leave game

**Leader config**:
```yaml
game:
  gameName: 'permanent-game'
  createGameAtMenu: true
```

**Companion config**:
```yaml
game:
  gameName: 'permanent-game'
  createGameAtMenu: false
```

**Process:**
1. Both join "permanent-game"
2. Run indefinitely in same game
3. Never exit to menu

**Pros:**
- No coordination needed after initial join
- Simple setup

**Cons:**
- Can't change games
- No way to restart if stuck
- May hit anti-bot detection (same game for hours)
- Can't sell/repair (unless using town trips)

### Workaround 4: Use External Bot Coordinator

**Setup:**
1. Use third-party tool like AutoHotkey or similar
2. Automate clicking "Join Game" button
3. Script types game name and password

**Pros:**
- No code changes to Koolo
- Visual automation

**Cons:**
- Requires separate automation tool
- Brittle (UI changes break it)
- Window focus issues
- Not reliable

**Verdict on Workarounds:**

All workarounds are **significantly worse** than just fixing the code. The code fix is literally adding 2 lines. Unless you're completely unable to compile the project, you should implement the proper fix.

---

## Implementation Plan: Attack Synchronization

This section provides a complete implementation plan for attack target synchronization.

### Feature Overview

**Goal**: When the leader attacks an enemy, all companions attack the same enemy.

**Benefits**:
- Faster kills (focus fire)
- More efficient farming
- Better for boss fights
- Coordinated party combat

**Complexity**: Medium (3-5 hours of development)

### Design Decisions to Make

**Question 1: How often should leader broadcast target?**

**Option A**: Every time target changes
- Pros: Most accurate, real-time sync
- Cons: Many events (1-5 per second during combat)

**Option B**: Every X seconds (e.g., every 2 seconds)
- Pros: Fewer events, less overhead
- Cons: Slight delay in sync

**Option C**: Only for priority targets (bosses, elites, dangerous enemies)
- Pros: Minimal events, focuses on important targets
- Cons: Companions act independently for trash mobs

**Recommendation**: Start with Option C, make it configurable later.

**Question 2: What if companion can't reach target?**

**Options**:
- Fall back to normal targeting
- Move toward target until in range
- Wait for leader to move on

**Recommendation**: Fall back to normal targeting after 3 seconds.

**Question 3: What if target dies before companion can attack?**

**Options**:
- Wait for next target broadcast
- Resume normal targeting immediately

**Recommendation**: Resume normal targeting immediately.

### Implementation Steps

#### Step 1: Add Configuration

**File**: `internal/config/config.go` (line 296)

**Add to Companion struct**:
```go
type Companion struct {
    Enabled               bool   `yaml:"enabled"`
    Leader                bool   `yaml:"leader"`
    LeaderName            string `yaml:"leaderName"`
    GameNameTemplate      string `yaml:"gameNameTemplate"`
    GamePassword          string `yaml:"gamePassword"`
    CompanionGameName     string `yaml:"companionGameName"`
    CompanionGamePassword string `yaml:"companionGamePassword"`

    // NEW: Attack synchronization settings
    AttackSync           bool   `yaml:"attackSync"`           // Enable attack synchronization
    AttackSyncPriority   string `yaml:"attackSyncPriority"`   // "all", "elite", "boss"
} `yaml:"companion"`
```

**Update template config** (`config/template/config.yaml`, line 217):
```yaml
companion:
  enabled: false
  leader: true
  leaderName: ''
  attackSync: true           # NEW
  attackSyncPriority: 'elite' # NEW: Options: all, elite, boss
  gameNameTemplate: game-
  gamePassword: xxx
```

#### Step 2: Leader Broadcasts Attack Target

**File**: `internal/action/step/attack.go`

**Find the attack function** (you'll need to locate where attacks are executed).

**Add after target selection**:
```go
func Attack(target data.UnitID, maxAttempts int) error {
    ctx := context.Get()

    // NEW: Broadcast attack target if leader
    if ctx.CharacterCfg.Companion.Enabled &&
       ctx.CharacterCfg.Companion.Leader &&
       ctx.CharacterCfg.Companion.AttackSync {

        // Check if this is a priority target
        shouldBroadcast := false
        targetUnit, found := ctx.Data.Monsters.FindByID(target)

        if found {
            switch ctx.CharacterCfg.Companion.AttackSyncPriority {
            case "all":
                shouldBroadcast = true
            case "elite":
                shouldBroadcast = targetUnit.IsElite() || targetUnit.IsBoss()
            case "boss":
                shouldBroadcast = targetUnit.IsBoss()
            }
        }

        if shouldBroadcast {
            event.Send(event.CompanionLeaderAttack(
                event.Text(ctx.Name, "Leader attacking target"),
                target,
            ))
        }
    }

    // ... rest of attack logic
}
```

**Location notes**:
- You'll need to find where attacks actually happen
- Likely in `internal/action/step/attack.go` or similar
- May be in character-specific attack implementations
- Add broadcasting AFTER target is selected but BEFORE attack starts

#### Step 3: Companion Receives Attack Event

**File**: `internal/bot/companion.go`

**Add to Handle() function** (after line 56):
```go
case event.CompanionLeaderAttackEvent:
    // Store leader's target in context for companion to use
    if h.cfg.Companion.Enabled &&
       !h.cfg.Companion.Leader &&
       h.cfg.Companion.AttackSync {

        // Check leader name matches
        if h.cfg.Companion.LeaderName == "" ||
           evt.Leader == h.cfg.Companion.LeaderName {

            h.log.Debug("Received leader attack target",
                slog.String("targetID", evt.TargetUnitID.String()))

            // Store in context (need to add this field)
            // This is simplified - you'll need proper context storage
            // Possibly store in h.cfg or create a companion state struct
        }
    }
```

**Better approach: Create companion state struct**:

Add new file `internal/bot/companion_state.go`:
```go
package bot

import (
    "sync"
    "time"
    "github.com/hectorgimenez/d2go/pkg/data"
)

type CompanionState struct {
    mu                  sync.RWMutex
    leaderTargetID      data.UnitID
    leaderTargetSetTime time.Time
}

func NewCompanionState() *CompanionState {
    return &CompanionState{}
}

func (cs *CompanionState) SetLeaderTarget(targetID data.UnitID) {
    cs.mu.Lock()
    defer cs.mu.Unlock()
    cs.leaderTargetID = targetID
    cs.leaderTargetSetTime = time.Now()
}

func (cs *CompanionState) GetLeaderTarget() (data.UnitID, bool) {
    cs.mu.RLock()
    defer cs.mu.RUnlock()

    // Target expires after 5 seconds
    if time.Since(cs.leaderTargetSetTime) > 5*time.Second {
        return data.UnitID(0), false
    }

    return cs.leaderTargetID, cs.leaderTargetID != 0
}

func (cs *CompanionState) ClearLeaderTarget() {
    cs.mu.Lock()
    defer cs.mu.Unlock()
    cs.leaderTargetID = 0
}
```

**Update companion handler to use state**:
```go
type CompanionEventHandler struct {
    supervisor string
    log        *slog.Logger
    cfg        *config.CharacterCfg
    state      *CompanionState  // NEW
}

// Update constructor
func NewCompanionEventHandler(supervisor string, log *slog.Logger, cfg *config.CharacterCfg, state *CompanionState) *CompanionEventHandler {
    return &CompanionEventHandler{
        supervisor: supervisor,
        log:        log,
        cfg:        cfg,
        state:      state,  // NEW
    }
}

// Handle attack event
case event.CompanionLeaderAttackEvent:
    if h.cfg.Companion.Enabled &&
       !h.cfg.Companion.Leader &&
       h.cfg.Companion.AttackSync {

        if h.cfg.Companion.LeaderName == "" ||
           evt.Leader == h.cfg.Companion.LeaderName {

            h.state.SetLeaderTarget(evt.TargetUnitID)
            h.log.Debug("Set leader target for companion")
        }
    }
```

#### Step 4: Companion Uses Leader's Target

**File**: `internal/action/step/attack.go` (or wherever target selection happens)

**Modify target selection logic**:
```go
func SelectTarget() data.UnitID {
    ctx := context.Get()

    // NEW: Check if we should use leader's target
    if ctx.CharacterCfg.Companion.Enabled &&
       !ctx.CharacterCfg.Companion.Leader &&
       ctx.CharacterCfg.Companion.AttackSync {

        // Get leader's target from state
        // You'll need to pass companion state through context
        if companionState != nil {
            leaderTarget, hasTarget := companionState.GetLeaderTarget()
            if hasTarget {
                // Verify target still exists and is valid
                targetUnit, found := ctx.Data.Monsters.FindByID(leaderTarget)
                if found && !targetUnit.IsDead() && isTargetInRange(targetUnit) {
                    ctx.Logger.Debug("Using leader's target")
                    return leaderTarget
                } else {
                    // Target invalid, clear it
                    companionState.ClearLeaderTarget()
                }
            }
        }
    }

    // Fall back to normal target selection
    return normalTargetSelection()
}
```

**Challenge**: You need to pass `CompanionState` through the context.

**Add to context** (`internal/game/context.go` or `internal/context/context.go`):
```go
type Context struct {
    // ... existing fields
    CompanionState *bot.CompanionState  // NEW
}
```

**Initialize in bot creation** (`internal/bot/manager.go`):
```go
// Create companion state for this character
companionState := bot.NewCompanionState()
ctx.CompanionState = companionState

// Pass state to handler
companionHandler := NewCompanionEventHandler(supervisorName, logger, ctx.CharacterCfg, companionState)
```

#### Step 5: Handle Edge Cases

**Target out of range**:
```go
if !isTargetInRange(targetUnit) {
    // Move toward target
    // Or fall back to normal targeting after timeout
    if time.Since(ctx.CompanionState.LeaderTargetSetTime) > 3*time.Second {
        companionState.ClearLeaderTarget()
        return normalTargetSelection()
    }
}
```

**Target already dead**:
```go
if targetUnit.IsDead() {
    companionState.ClearLeaderTarget()
    return normalTargetSelection()
}
```

**Target immune to companion's damage**:
```go
if isImmuneTo(targetUnit, ctx.Char.DamageType) {
    companionState.ClearLeaderTarget()
    return normalTargetSelection()
}
```

### Testing Attack Synchronization

**Test 1: Basic Attack Sync**
1. Configure leader and companion with `attackSync: true`
2. Start both in same game
3. Leader attacks boss
4. Verify companion also attacks boss
5. Check logs for "Received leader attack target" messages

**Test 2: Priority Filtering**
1. Set `attackSyncPriority: elite`
2. Leader attacks normal monster
3. Verify companion does NOT follow (no event sent)
4. Leader attacks elite
5. Verify companion DOES follow

**Test 3: Out of Range**
1. Leader attacks target far from companion
2. Verify companion either moves toward target or uses own targeting
3. Should not get stuck waiting for unreachable target

**Test 4: Dead Target**
1. Leader attacks target
2. Target dies quickly
3. Verify companion doesn't get stuck attacking corpse
4. Companion should move to next target

### Estimated Implementation Time

- Configuration: 15 minutes
- Leader broadcasting: 30 minutes
- Companion state management: 45 minutes
- Target selection override: 1 hour
- Edge case handling: 1 hour
- Testing: 1 hour
- **Total: 4 hours**

---

## Implementation Plan: Movement Following

This section provides a complete implementation plan for active movement following.

### Feature Overview

**Goal**: Companions actively follow the leader's position, staying within a configured distance.

**Benefits**:
- Party stays together
- Companions don't get left behind
- Works with teleporting leaders (Sorceress)
- More coordinated movement

**Complexity**: Medium-High (4-6 hours of development)

### Design Decisions to Make

**Question 1: How often should leader broadcast position?**

**Options**:
- Every second
- Every 2-3 seconds
- Only when moving significantly (5+ units)

**Recommendation**: Every 2 seconds, only if position changed by 5+ units.

**Question 2: How close should companions stay?**

**Options**:
- Fixed distance (e.g., always within 15 units)
- Configurable distance
- Context-dependent (closer in combat, looser when exploring)

**Recommendation**: Configurable distance with default of 20 units.

**Question 3: Should companions interrupt actions to follow?**

**Options**:
- Always follow (interrupt everything)
- Don't interrupt combat
- Don't interrupt looting or town activities
- Follow only when idle

**Recommendation**: Follow priority: idle > exploring > looting. Never interrupt combat.

**Question 4: Multiple companions - formation?**

**Options**:
- All companions cluster around leader
- Companions maintain spread formation
- No formation control (just follow)

**Recommendation**: Start with clustering, add formation later.

### Implementation Steps

#### Step 1: Add Configuration

**File**: `internal/config/config.go`

**Add to Companion struct**:
```go
type Companion struct {
    // ... existing fields

    // Movement following settings
    FollowLeader         bool `yaml:"followLeader"`         // Enable active following
    FollowDistance       int  `yaml:"followDistance"`       // Max distance before following
    FollowPriority       int  `yaml:"followPriority"`       // 1=high, 2=medium, 3=low
} `yaml:"companion"`
```

**Update template**:
```yaml
companion:
  # ... existing settings
  followLeader: true
  followDistance: 20        # Follow if leader more than 20 units away
  followPriority: 2         # 1=always follow, 2=follow when not busy, 3=only when idle
```

#### Step 2: Leader Broadcasts Position

**Create new file**: `internal/bot/position_broadcaster.go`

```go
package bot

import (
    "time"
    "github.com/hectorgimenez/d2go/pkg/data"
    "github.com/hectorgimenez/koolo/internal/event"
    "github.com/hectorgimenez/koolo/internal/game"
)

type PositionBroadcaster struct {
    ctx            *game.Context
    lastPosition   data.Position
    lastBroadcast  time.Time
    ticker         *time.Ticker
    stopChan       chan bool
}

func NewPositionBroadcaster(ctx *game.Context) *PositionBroadcaster {
    return &PositionBroadcaster{
        ctx:      ctx,
        stopChan: make(chan bool),
    }
}

func (pb *PositionBroadcaster) Start() {
    // Only start if this is a leader
    if !pb.ctx.CharacterCfg.Companion.Enabled ||
       !pb.ctx.CharacterCfg.Companion.Leader {
        return
    }

    pb.ticker = time.NewTicker(2 * time.Second)

    go func() {
        for {
            select {
            case <-pb.ticker.C:
                pb.broadcastPosition()
            case <-pb.stopChan:
                return
            }
        }
    }()
}

func (pb *PositionBroadcaster) Stop() {
    if pb.ticker != nil {
        pb.ticker.Stop()
        close(pb.stopChan)
    }
}

func (pb *PositionBroadcaster) broadcastPosition() {
    // Don't broadcast in town
    if pb.ctx.Data.PlayerUnit.Area.IsTown() {
        return
    }

    currentPos := pb.ctx.Data.PlayerUnit.Position

    // Only broadcast if position changed significantly
    distance := pb.calculateDistance(pb.lastPosition, currentPos)
    if distance < 5 {
        return
    }

    // Broadcast position
    event.Send(event.CompanionLeaderPosition(
        event.Text(pb.ctx.Name, "Leader position update"),
        pb.ctx.CharacterCfg.CharacterName,
        currentPos,
    ))

    pb.lastPosition = currentPos
    pb.lastBroadcast = time.Now()
}

func (pb *PositionBroadcaster) calculateDistance(pos1, pos2 data.Position) float64 {
    dx := float64(pos2.X - pos1.X)
    dy := float64(pos2.Y - pos1.Y)
    return math.Sqrt(dx*dx + dy*dy)
}
```

#### Step 3: Create Position Event

**File**: `internal/event/events.go`

**Add new event** (around line 130):
```go
// CompanionLeaderPositionEvent is sent when leader changes position
type CompanionLeaderPositionEvent struct {
    BaseEvent
    Leader   string        // Leader character name
    Position data.Position // Leader's position
}

func CompanionLeaderPosition(be BaseEvent, leader string, position data.Position) CompanionLeaderPositionEvent {
    return CompanionLeaderPositionEvent{
        BaseEvent: be,
        Leader:    leader,
        Position:  position,
    }
}
```

#### Step 4: Companion Receives Position

**File**: `internal/bot/companion.go`

**Add to Handle() function**:
```go
case event.CompanionLeaderPositionEvent:
    if h.cfg.Companion.Enabled &&
       !h.cfg.Companion.Leader &&
       h.cfg.Companion.FollowLeader {

        if h.cfg.Companion.LeaderName == "" ||
           evt.Leader == h.cfg.Companion.LeaderName {

            h.state.SetLeaderPosition(evt.Position)
            h.log.Debug("Updated leader position",
                slog.Int("x", evt.Position.X),
                slog.Int("y", evt.Position.Y))
        }
    }
```

**Update companion state** (`internal/bot/companion_state.go`):
```go
type CompanionState struct {
    mu                  sync.RWMutex
    leaderTargetID      data.UnitID
    leaderTargetSetTime time.Time
    leaderPosition      data.Position  // NEW
    leaderPosSetTime    time.Time      // NEW
}

func (cs *CompanionState) SetLeaderPosition(pos data.Position) {
    cs.mu.Lock()
    defer cs.mu.Unlock()
    cs.leaderPosition = pos
    cs.leaderPosSetTime = time.Now()
}

func (cs *CompanionState) GetLeaderPosition() (data.Position, bool) {
    cs.mu.RLock()
    defer cs.mu.RUnlock()

    // Position expires after 10 seconds (leader might have left area)
    if time.Since(cs.leaderPosSetTime) > 10*time.Second {
        return data.Position{}, false
    }

    return cs.leaderPosition, cs.leaderPosition.X != 0 || cs.leaderPosition.Y != 0
}
```

#### Step 5: Companion Follows Leader

**Create new file**: `internal/action/follow_leader.go`

```go
package action

import (
    "github.com/hectorgimenez/koolo/internal/context"
    "github.com/hectorgimenez/koolo/internal/action/step"
)

func FollowLeader() error {
    ctx := context.Get()

    // Only follow if enabled
    if !ctx.CharacterCfg.Companion.Enabled ||
       ctx.CharacterCfg.Companion.Leader ||
       !ctx.CharacterCfg.Companion.FollowLeader {
        return nil
    }

    // Get leader's position
    leaderPos, hasPos := ctx.CompanionState.GetLeaderPosition()
    if !hasPos {
        return nil
    }

    // Check distance to leader
    distance := ctx.PathFinder.DistanceFromMe(leaderPos)
    followDistance := ctx.CharacterCfg.Companion.FollowDistance
    if followDistance == 0 {
        followDistance = 20 // Default
    }

    // If too far, move toward leader
    if distance > followDistance {
        ctx.Logger.Debug("Following leader",
            slog.Int("distance", distance),
            slog.Int("threshold", followDistance))

        return step.MoveTo(leaderPos)
    }

    return nil
}
```

#### Step 6: Integrate Following into Bot Loop

**File**: `internal/bot/single_supervisor.go`

**Find the main game loop** (where actions are executed).

**Add follow check** (after health checks, before combat):
```go
// In game loop, periodically check if we should follow leader
if err := action.FollowLeader(); err != nil {
    s.bot.ctx.Logger.Debug("Follow leader action failed", slog.Any("error", err))
}
```

**Better approach**: Add to the action priority system.

**Find where actions are prioritized** (might be in `internal/run/` files).

**Add following as medium-priority action**:
```go
// Priority order (example):
// 1. Emergency (chicken, health)
// 2. Combat (if engaged)
// 3. Following leader (if too far)
// 4. Looting
// 5. Normal run objectives
```

#### Step 7: Prevent Following During Critical Actions

**Modify follow logic**:
```go
func FollowLeader() error {
    ctx := context.Get()

    // Check follow priority
    followPriority := ctx.CharacterCfg.Companion.FollowPriority
    if followPriority == 0 {
        followPriority = 2 // Default: medium
    }

    // Priority 1: Always follow (even during combat)
    // Priority 2: Follow unless in combat or looting
    // Priority 3: Follow only when idle

    currentAction := ctx.LastAction

    if followPriority == 3 {
        // Only follow when idle
        if currentAction != "Idle" && currentAction != "" {
            return nil
        }
    } else if followPriority == 2 {
        // Don't interrupt combat or looting
        if strings.Contains(currentAction, "Attack") ||
           strings.Contains(currentAction, "Loot") ||
           strings.Contains(currentAction, "Combat") {
            return nil
        }
    }

    // ... rest of follow logic
}
```

#### Step 8: Start/Stop Position Broadcaster

**File**: `internal/bot/single_supervisor.go`

**In supervisor initialization**:
```go
func NewSinglePlayerSupervisor(...) (*SinglePlayerSupervisor, error) {
    // ... existing code

    // Start position broadcaster if leader
    if bot.ctx.CharacterCfg.Companion.Enabled &&
       bot.ctx.CharacterCfg.Companion.Leader {
        positionBroadcaster := bot.NewPositionBroadcaster(bot.ctx)
        positionBroadcaster.Start()

        // Store for cleanup
        // You'll need to add field to supervisor struct
        supervisor.positionBroadcaster = positionBroadcaster
    }

    return supervisor, nil
}
```

**In supervisor cleanup**:
```go
func (s *SinglePlayerSupervisor) Stop() {
    // ... existing cleanup

    // Stop position broadcaster
    if s.positionBroadcaster != nil {
        s.positionBroadcaster.Stop()
    }
}
```

### Testing Movement Following

**Test 1: Basic Following**
1. Start leader and companion in game
2. Leader moves across map
3. Verify companion follows within configured distance
4. Check logs for position updates

**Test 2: Teleport Following**
1. Use Sorceress as leader with teleport
2. Teleport rapidly across map
3. Verify companion catches up

**Test 3: Priority Interruption**
1. Set `followPriority: 2`
2. Companion in combat
3. Leader moves away
4. Verify companion finishes combat before following

**Test 4: Lost Leader**
1. Leader and companion in different areas
2. Wait 10 seconds (position expires)
3. Verify companion doesn't try to path to old position

### Estimated Implementation Time

- Configuration: 15 minutes
- Position broadcasting: 1 hour
- Position event: 15 minutes
- Companion receives position: 30 minutes
- Follow action logic: 1.5 hours
- Integration into bot loop: 1 hour
- Priority system: 45 minutes
- Testing: 1 hour
- **Total: 6 hours**

---

## Configuration Improvements

### Make Constants Configurable

**Current hard-coded values:**
- Wait distance: 20 units
- Boss area clear: 30 units
- Wait area clear: 5 units
- Position update interval: 2 seconds
- Target expiration: 5 seconds

**Proposed config additions**:

**File**: `internal/config/config.go`

```go
type Companion struct {
    // ... existing fields

    // Advanced settings
    WaitDistance         int `yaml:"waitDistance"`         // How close companions must be
    WaitClearRadius      int `yaml:"waitClearRadius"`      // Clear radius while waiting
    BossClearRadius      int `yaml:"bossClearRadius"`      // Clear radius before boss
    PositionUpdateSec    int `yaml:"positionUpdateSec"`    // How often to broadcast position
    TargetExpirationSec  int `yaml:"targetExpirationSec"`  // How long attack targets last
} `yaml:"companion"`
```

**Template with defaults**:
```yaml
companion:
  enabled: false
  leader: true
  leaderName: ''
  attackSync: true
  followLeader: true
  followDistance: 20

  # Advanced settings (use defaults if not set)
  waitDistance: 20              # Wait if companions beyond this distance
  waitClearRadius: 5            # Clear this radius while waiting
  bossClearRadius: 30           # Clear this radius before boss
  positionUpdateSec: 2          # Broadcast position every 2 seconds
  targetExpirationSec: 5        # Attack targets expire after 5 seconds
```

### Validation

Add validation to prevent invalid configs:

```go
func (c *Companion) Validate() error {
    if c.FollowDistance < 5 || c.FollowDistance > 50 {
        return fmt.Errorf("followDistance must be between 5 and 50")
    }

    if c.WaitDistance < 5 || c.WaitDistance > 50 {
        return fmt.Errorf("waitDistance must be between 5 and 50")
    }

    if c.Leader && c.LeaderName != "" {
        return fmt.Errorf("leaders cannot have a leaderName set")
    }

    if !c.Leader && c.GameNameTemplate != "" {
        // Warning, not error
        log.Warn("Companions don't use gameNameTemplate")
    }

    return nil
}
```

---

## Testing Strategy

### Unit Testing

**Test event handler registration**:
```go
func TestCompanionEventHandlerRegistered(t *testing.T) {
    // Create test supervisor
    // Verify handler is in listener's handlers list
    // Send test event
    // Verify event received
}
```

**Test event handling**:
```go
func TestCompanionJoinEvent(t *testing.T) {
    cfg := &config.CharacterCfg{
        Companion: config.Companion{
            Enabled: true,
            Leader: false,
            LeaderName: "TestLeader",
        },
    }

    handler := NewCompanionEventHandler("test", logger, cfg)

    evt := event.RequestCompanionJoinGame(
        event.BaseEvent{},
        "TestLeader",
        "game-1",
        "pass123",
    )

    handler.Handle(context.Background(), evt)

    assert.Equal(t, "game-1", cfg.Companion.CompanionGameName)
    assert.Equal(t, "pass123", cfg.Companion.CompanionGamePassword)
}
```

### Integration Testing

**Test full flow**:
1. Start leader bot
2. Start companion bot
3. Leader creates game
4. Wait 10 seconds
5. Verify companion joined
6. Verify both in same game
7. Leader finishes game
8. Wait 5 seconds
9. Verify companion left
10. Leader creates new game
11. Verify companion joins again

### Manual Testing Checklist

- [ ] Leader creates game with correct name
- [ ] Companion receives join event (check logs)
- [ ] Companion joins game within 10 seconds
- [ ] Both characters are in same game
- [ ] Leader finishes game
- [ ] Companion receives reset event (check logs)
- [ ] Companion clears stored game info
- [ ] Process repeats for multiple games
- [ ] Multiple companions can join same leader
- [ ] Companion ignores events from wrong leader
- [ ] Empty leader name accepts any leader

### Performance Testing

**Metrics to monitor**:
- Event processing time (should be < 1ms)
- Game join success rate (should be > 95%)
- Time from event to join (should be < 10 seconds)
- CPU usage per character (should be < 10%)
- Memory per character (should be < 100MB)

**Load testing**:
- Test with 1 leader + 1 companion
- Test with 1 leader + 3 companions
- Test with 1 leader + 7 companions (full party)
- Test with 2 leaders + 6 companions (two parties)

---

## Troubleshooting Guide

### Companion Not Joining Games

**Symptom**: Companion stays in menu, never joins leader's game.

**Diagnostic steps**:

**Step 1: Check logs**
Look for: `"Companion event handler registered"`
- If missing: Handler not registered (apply critical fix)
- If present: Handler is registered, continue

**Step 2: Check leader events**
Look for: `"New Game Started game-X"`
- If missing: Leader not sending events (check leader config)
- If present: Events are being sent, continue

**Step 3: Check companion events**
Look for: `"Companion join game event received"`
- If missing: Companion not receiving events
  - Check if both bots in same Koolo instance
  - Check companion's `enabled: true` and `leader: false`
  - Check `leaderName` matches exactly
- If present: Events received, continue

**Step 4: Check menu flow**
Look for: `"HandleCompanionMenuFlow"` or similar
- Check if `CompanionGameName` is set
- Check if companion is stuck on character select screen
- Check if game join is failing (game doesn't exist, wrong password, etc.)

**Step 5: Check timing**
- Leader creates game at time T
- Companion receives event at time T+1
- Companion tries to join at time T+2
- Join might fail if game not fully created yet
- Companion should retry and eventually succeed

**Common fixes**:
1. Register event handler (critical fix)
2. Fix `leaderName` spelling/capitalization
3. Ensure both bots in same Koolo process
4. Add delay before companion joins (if race condition)

### Attack Sync Not Working

**Symptom**: Companions don't attack leader's target.

**Diagnostic steps**:

**Step 1: Check if implemented**
- Attack sync is NOT implemented by default
- If you haven't added the code, it won't work
- See [Implementation Plan: Attack Synchronization](#implementation-plan-attack-synchronization)

**Step 2: Check configuration**
```yaml
companion:
  attackSync: true  # Must be true
```

**Step 3: Check logs**
Leader should log: `"Leader attacking target"`
Companion should log: `"Received leader attack target"`

**Step 4: Check event sending**
Verify leader is actually sending `CompanionLeaderAttackEvent`

**Step 5: Check target selection**
Verify companion's target selection logic uses leader's target

### Movement Following Not Working

**Symptom**: Companions don't follow leader.

**Diagnostic steps**:

**Step 1: Check if implemented**
- Movement following is NOT implemented by default
- If you haven't added the code, it won't work
- See [Implementation Plan: Movement Following](#implementation-plan-movement-following)

**Step 2: Check configuration**
```yaml
companion:
  followLeader: true
  followDistance: 20
```

**Step 3: Check position broadcasting**
Leader should log position updates every 2 seconds

**Step 4: Check companion receiving**
Companion should log: `"Updated leader position"`

**Step 5: Check follow action**
Companion should execute `FollowLeader()` action when too far

### Both Characters Acting as Leaders

**Symptom**: Both characters create their own games.

**Fix**:
Check `leader` setting in each config:
- Leader: `leader: true`
- Companion: `leader: false`

One must be `true`, others must be `false`.

### Companion Joins Wrong Leader's Game

**Symptom**: Companion joins different leader than intended.

**Fix**:
Set specific `leaderName` in companion config:
```yaml
companion:
  leaderName: 'ExactLeaderName'  # Must match exactly
```

Check spelling and capitalization - must match leader's `characterName` exactly.

### Race Condition: Game Doesn't Exist

**Symptom**: Companion gets error "game does not exist" when trying to join.

**Cause**: Companion tries to join before game is fully created.

**Fix**:
The retry logic should handle this. Companion will retry every 2 seconds until successful.

**Alternative**: Add manual delay in companion join logic (not recommended).

### Logs Show Events But Nothing Happens

**Symptom**: Logs show events sent and received, but behavior doesn't change.

**Possible causes**:
1. Config not being read correctly
2. Event handler not actually modifying config
3. Logic not checking the right config fields
4. Race condition in config access

**Debug**:
Add more detailed logging:
```go
h.log.Info("After event, game name is",
    slog.String("gameName", h.cfg.Companion.CompanionGameName))
```

Verify the config fields are actually being set.

---

## Questions for You

To provide the most relevant fixes, please answer:

### Priority Questions

**1. What's broken for you right now?**
- [ ] Companions not joining games at all
- [ ] Companions join but then just stand there
- [ ] Attack sync doesn't work
- [ ] Movement following doesn't work
- [ ] Something else: ___________

**2. What do you want to work?**
- [ ] Just basic game joining (minimum viable)
- [ ] Game joining + companions play independently
- [ ] Game joining + attack sync
- [ ] Game joining + movement following
- [ ] Everything (full synchronization)

**3. How many companions?**
- [ ] 1 leader + 1 companion (2 total)
- [ ] 1 leader + 2-3 companions
- [ ] 1 leader + 4-7 companions (full party)
- [ ] Multiple separate groups

**4. What content?**
- [ ] Boss runs (Mephisto, Diablo, Baal)
- [ ] Area farming (Pits, Chaos, etc.)
- [ ] Leveling characters together
- [ ] Uber Tristram / endgame
- [ ] Other: ___________

### Technical Questions

**5. Can you modify Go code?**
- [ ] Yes, I can edit and compile
- [ ] Yes, but need detailed instructions
- [ ] No, only config changes
- [ ] No, only workarounds

**6. Current symptoms:**
- [ ] Companion never joins leader's game
- [ ] Companion joins but is inactive
- [ ] No log messages about companion events
- [ ] Events in logs but no behavior change
- [ ] Other: ___________

**7. Have you modified the code already?**
- [ ] No modifications
- [ ] Minor config changes
- [ ] Some code changes
- [ ] Extensive modifications

### Clarification Questions

**8. When you say "doesn't work", do you mean:**
- [ ] Companions don't join games (most critical)
- [ ] Companions join but don't coordinate
- [ ] Specific features (attack/movement) don't work
- [ ] Everything is broken

**9. What have you tried?**
- [ ] Nothing yet, just discovered issue
- [ ] Changed configs
- [ ] Checked logs
- [ ] Modified code
- [ ] Other: ___________

**10. Time constraints:**
- [ ] Need quick workaround now
- [ ] Can implement proper fixes over time
- [ ] Want to understand everything first
- [ ] Just want it to work, don't care how

---

## Next Steps

Based on your answers, I can provide:

1. **Immediate fix** - Get basic game joining working (20 minutes)
2. **Short-term improvements** - Better configuration, stability fixes (2 hours)
3. **Medium-term features** - Attack sync OR movement following (4-6 hours)
4. **Long-term complete solution** - All features fully implemented (20+ hours)

**My recommendation**: Start with the critical fix (register event handler) to get basic companion joining working, then decide if you need additional features.

---

## Summary

**The #1 issue**: CompanionEventHandler not registered
- **Fix**: Add 2 lines in `internal/bot/manager.go`
- **Time**: 5 minutes
- **Result**: Companions will join leader's games

**Other issues**:
- Attack sync: Defined but not implemented (4 hours to fix)
- Movement following: Partially implemented (6 hours to complete)
- Various improvements: Configurable constants, better error handling

**Recommended path forward**:
1. Apply critical fix first (register handler)
2. Test basic game joining
3. If that works, decide which features you want next
4. Implement features one at a time
5. Test thoroughly between each feature

Let me know your answers to the questions and I can provide more specific guidance!
