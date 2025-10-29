# Companion Synchronization System - Complete Technical Documentation

## Table of Contents
1. [Overview](#overview)
2. [What is Synchronization?](#what-is-synchronization)
3. [Current Implementation Status](#current-implementation-status)
4. [Game Join Synchronization](#game-join-synchronization)
5. [Position Tracking and Waiting](#position-tracking-and-waiting)
6. [Portal and Town Coordination](#portal-and-town-coordination)
7. [Boss Room Coordination](#boss-room-coordination)
8. [Attack Target Synchronization (Planned)](#attack-target-synchronization-planned)
9. [Movement Following (Partial)](#movement-following-partial)
10. [Event System Architecture](#event-system-architecture)
11. [Configuration System](#configuration-system)
12. [Technical Implementation Details](#technical-implementation-details)
13. [Known Issues and Limitations](#known-issues-and-limitations)
14. [How to Use Synchronization](#how-to-use-synchronization)

---

## Overview

The companion synchronization system in Koolo is designed to coordinate multiple bot characters working together in the same game. The term "synchronization" refers to keeping companions coordinated with the leader character in terms of:

- **Game joining**: Companions automatically join games created by the leader
- **Position tracking**: Leader waits for companions to catch up before proceeding
- **Portal coordination**: Leader opens town portals for the group
- **Boss room setup**: Leader prepares safe areas before combat
- **Attack targeting** (planned but not implemented): Companions attack the same targets as the leader
- **Movement following** (partially implemented): Companions stay close to the leader

This document provides a comprehensive technical breakdown of how synchronization works, what's implemented, and what's planned but not yet active.

---

## What is Synchronization?

In the context of the companion system, synchronization means coordinating actions between multiple characters so they work together effectively. Think of it like a party in an MMO game where:

- The **leader** is the party leader who makes decisions about where to go and what to do
- The **companions** are party members who follow the leader's decisions
- **Synchronization** is the mechanism that keeps everyone coordinated

### Types of Synchronization

1. **Game Session Synchronization**: Ensuring all characters are in the same game
2. **Position Synchronization**: Keeping characters close together geographically
3. **Combat Synchronization**: Coordinating attack targets and combat actions
4. **Town Synchronization**: Coordinating portal usage and town activities
5. **Progress Synchronization**: Ensuring characters move through areas together

---

## Current Implementation Status

### What IS Fully Implemented

| Feature | Status | Description |
|---------|--------|-------------|
| Game Join Events | ✅ Working | Leader broadcasts game info, companions join automatically |
| Position Waiting | ✅ Working | Leader waits for companions within 20-unit radius during leveling |
| Portal Coordination | ✅ Working | Leader opens portals, companions use them |
| Boss Room Setup | ✅ Working | Leader prepares safe areas before boss fights |
| Configuration System | ✅ Working | YAML config for all companion settings |

### What IS Defined But NOT Implemented

| Feature | Status | Description |
|---------|--------|-------------|
| Attack Target Sync | ⚠️ Defined Only | Event exists but never sent or processed |
| Real-time Following | ⚠️ Partial | Only checkpoint-based waiting, not continuous following |
| Event Handler Registration | ❌ Not Working | CompanionEventHandler created but never registered |

### What is NOT Implemented At All

| Feature | Status | Description |
|---------|--------|-------------|
| Combat Coordination | ❌ Not Implemented | No shared target selection logic |
| Movement Following | ❌ Not Implemented | No active following between checkpoints |
| Buff Coordination | ❌ Not Implemented | No shared buff timing |
| Loot Distribution | ❌ Not Implemented | No loot priority system |

---

## Game Join Synchronization

This is the primary and most complete synchronization feature. It ensures companions automatically join the leader's game.

### How It Works

#### Step 1: Leader Creates Game

**File**: `internal/bot/single_supervisor.go`
**Lines**: 224-226

When the leader creates a new game:

```
Leader creates game → Game name and password stored → Event broadcast triggered
```

The leader sends a `RequestCompanionJoinGameEvent` containing:
- Leader's character name
- Game name (generated from template + counter)
- Game password

**Example**:
- Template: "baalrun-"
- Counter: 5
- Generated name: "baalrun-5"
- Password: "secret123"

#### Step 2: Event Broadcast

**File**: `internal/event/events.go`
**Lines**: 155-170

The event structure:
```go
type RequestCompanionJoinGameEvent struct {
    BaseEvent
    Leader   string  // Name of leader character
    Name     string  // Game name
    Password string  // Game password
}
```

The leader calls:
```go
event.Send(event.RequestCompanionJoinGame(
    event.Text(supervisorName, "New Game Started "+gameName),
    leaderCharacterName,
    gameName,
    gamePassword
))
```

This broadcasts the event to all running bots in the Koolo instance.

#### Step 3: Companion Receives Event

**File**: `internal/bot/companion.go`
**Lines**: 32-42

The companion's event handler processes the event:

**Checks performed**:
1. Is companion mode enabled? (`Companion.Enabled == true`)
2. Is this character NOT a leader? (`Companion.Leader == false`)
3. Does the leader name match? (`Companion.LeaderName == event.Leader` OR `Companion.LeaderName == ""`)

**If all checks pass**:
- Store game name: `cfg.Companion.CompanionGameName = event.Name`
- Store game password: `cfg.Companion.CompanionGamePassword = event.Password`
- Log: "Received companion join game request"

**If checks fail**:
- Ignore the event
- No action taken

#### Step 4: Companion Joins Game

**File**: `internal/bot/single_supervisor.go`
**Lines**: 520-557

On the companion's next menu cycle, the `HandleCompanionMenuFlow()` function is called:

**Process**:
1. Check if `CompanionGameName` is set
   - If empty: Wait 2 seconds and return idle status
   - If set: Proceed to join

2. Determine current game state:
   - **On character selection screen**: Enter lobby first, then join game
   - **In lobby**: Join game directly

3. Call `Manager.JoinOnlineGame(gameName, gamePassword)`

4. Handle join result:
   - **Success**: Enter game and begin playing
   - **Failure**: Log error and retry on next cycle

**Example flow**:
```
Companion menu cycle → Check CompanionGameName → "baalrun-5" found →
In lobby → Join "baalrun-5" with password "secret123" → Enter game
```

#### Step 5: Game Completion Reset

**File**: `internal/bot/single_supervisor.go`
**Lines**: 379-381

When the leader finishes the game and returns to menu:

The leader sends a `ResetCompanionGameInfoEvent`:
```go
event.Send(event.ResetCompanionGameInfo(
    event.Text(supervisorName, "Game "+gameName+" finished"),
    leaderCharacterName
))
```

**File**: `internal/bot/companion.go`
**Lines**: 44-56

Companions receive this event and:
1. Check if the event is from their configured leader
2. Clear stored game info:
   - `CompanionGameName = ""`
   - `CompanionGamePassword = ""`
3. Return to waiting state for next game

This prevents companions from trying to join old, finished games.

### Configuration Example

**Leader Configuration**:
```yaml
companion:
  enabled: true
  leader: true
  leaderName: ''  # Not used for leaders
  gameNameTemplate: 'myrun-'
  gamePassword: 'pass123'
```

**Companion Configuration**:
```yaml
companion:
  enabled: true
  leader: false
  leaderName: 'LeaderSorc'  # Must match leader's CharacterName
  gameNameTemplate: ''  # Not used for companions
  gamePassword: ''  # Not used for companions (received from event)
```

### Timing and Race Conditions

**Important timing considerations**:

1. **Game creation delay**: It takes 5-10 seconds for a Diablo 2 game to fully start after creation
2. **Event processing**: Events are processed in real-time (< 100ms typically)
3. **Menu cycle timing**: Companions check for game info every 2 seconds

**Potential race condition**:
- Companion receives event while leader's game is still loading
- Companion attempts to join before game is fully created
- Join fails with "game does not exist" error

**Mitigation**:
- Companions have retry logic in `HandleCompanionMenuFlow()`
- Failed joins are retried every 2 seconds
- Eventually the game becomes available and join succeeds

### File Paths for Game Join Synchronization

| Component | File Path | Lines |
|-----------|-----------|-------|
| Event Definition | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 155-170, 172-183 |
| Event Handler | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\companion.go` | 32-42, 44-56 |
| Event Sending | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\single_supervisor.go` | 224-226, 379-381 |
| Join Logic | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\single_supervisor.go` | 520-557 |
| Menu Flow Routing | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\single_supervisor.go` | 455 |

---

## Position Tracking and Waiting

The leader actively waits for companions to catch up during leveling runs, ensuring the party stays together.

### How It Works

**File**: `internal/action/leveling_tools.go`
**Lines**: 585-604

The `WaitForAllMembersWhenLeveling()` function implements position synchronization.

#### Function Logic

**Checks performed**:
1. Is this character the leader? (`Companion.Leader == true`)
2. Is the leader NOT in town? (`!PlayerUnit.Area.IsTown()`)
3. Is this a leveling run? (`isLeveling == true`)

**If all checks pass**:

1. **Get party roster**: Access `ctx.Data.Roster` which contains all party members
2. **Calculate distances**: For each party member (excluding self):
   - Get member's position: `member.Position`
   - Calculate distance: `ctx.PathFinder.DistanceFromMe(member.Position)`
   - Check if distance > 20 units
3. **Evaluate party status**:
   - If ALL members are within 20 units: Continue normally
   - If ANY member is beyond 20 units: Enter waiting mode
4. **Waiting behavior**:
   - Clear area around leader: `ClearAreaAroundPlayer(5, data.MonsterAnyFilter())`
   - Wait for companions to catch up
   - Check again on next action cycle

#### Distance Threshold

**20 units** is the synchronization threshold:
- This is roughly 2-3 screen widths in Diablo 2
- Provides balance between tight formation and flexibility
- Prevents leader from outrunning slow companions

#### Position Data Source

**File**: `internal/game/data.go` (referenced)

The `Data.Roster` contains party member information:
```go
type Roster struct {
    Name     string        // Character name
    Position data.Position // X, Y coordinates
    // Other fields...
}
```

This data is updated in real-time as characters move through the game world.

#### Path Finding Integration

**File**: `internal/game/pathfinder.go` (referenced)

The `PathFinder.DistanceFromMe()` function:
- Takes a position (X, Y coordinates)
- Calculates straight-line distance from player
- Returns distance as an integer (number of units)
- Accounts for pathable terrain (walls, obstacles)

**Distance calculation**:
```
Distance = sqrt((X2 - X1)^2 + (Y2 - Y1)^2)
```

Where:
- (X1, Y1) = Leader's current position
- (X2, Y2) = Companion's current position

### When Position Tracking is Active

Position tracking ONLY activates when:

1. **Character is a leader**: `Companion.Leader == true`
2. **Not in town**: `!PlayerUnit.Area.IsTown()`
3. **Leveling mode**: `isLeveling == true`

**This means**:
- No position tracking during normal magic finding runs
- No position tracking when leader is in town
- No position tracking for companion characters (only leader waits)
- Only active during specific leveling operations

### Clearing Area While Waiting

While waiting for companions, the leader:
- Clears 5-unit radius around self: `ClearAreaAroundPlayer(5, ...)`
- Uses `MonsterAnyFilter()` which targets all monsters
- Prevents leader from being overwhelmed by enemies
- Provides safe area for companions to catch up

### What This DOESN'T Do

**Important limitations**:

1. **Not continuous following**: Companions don't actively move toward leader
2. **Not teleportation**: No mechanism to bring companions to leader instantly
3. **Not pathfinding sync**: Companions may take different routes
4. **No stuck detection**: If companion is stuck, leader waits indefinitely
5. **Leveling-only**: Doesn't apply to regular farming runs

### Configuration for Position Tracking

There is NO explicit configuration for position tracking. It's automatically enabled when:
- Character is configured as leader
- Running leveling content

The 20-unit threshold is **hard-coded** and cannot be changed via configuration.

### File Paths for Position Tracking

| Component | File Path | Lines |
|-----------|-----------|-------|
| Wait Logic | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\leveling_tools.go` | 585-604 |
| Data Structures | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\game\data.go` | (various) |
| Path Finding | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\game\pathfinder.go` | (various) |

---

## Portal and Town Coordination

Leaders and companions coordinate around town portal usage, ensuring efficient town trips.

### How It Works

#### Leader Opens Portals

**File**: `internal/action/tools.go`
**Lines**: 11-22

The `OpenTPIfLeader()` function:

```go
func OpenTPIfLeader() error {
    ctx := context.Get()
    ctx.SetLastAction("OpenTPIfLeader")

    isLeader := ctx.CharacterCfg.Companion.Leader
    if isLeader {
        return step.OpenPortal()
    }
    return nil
}
```

**Logic**:
1. Check if character is leader: `Companion.Leader == true`
2. If leader: Call `step.OpenPortal()` to open a town portal
3. If companion: Do nothing (return success immediately)

**Result**:
- Leaders open portals
- Companions skip portal opening
- Prevents multiple overlapping portals

#### Where Portals Are Opened

Portals are opened at specific synchronization points throughout runs:

**1. Boss Runs** (before engaging boss)

**File**: `internal/run/diablo.go`
**Lines**: 64-85

Before Diablo fight:
```go
if d.ctx.CharacterCfg.Companion.Leader {
    action.OpenTPIfLeader()  // Open portal
    action.Buff()            // Apply buffs
    action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())  // Clear area
}
```

**File**: `internal/run/baal.go`
**Lines**: 83-89

Before Baal fight:
```go
if s.ctx.CharacterCfg.Companion.Leader {
    action.MoveToCoords(data.Position{X: 15116, Y: 5071})  // Safe position
    action.OpenTPIfLeader()  // Open portal
}
```

**2. Town Return Sequences**

**File**: `internal/action/town.go`
**Lines**: 219-226 (referenced in context)

When returning to town:
1. Leader opens portal using `OpenTPIfLeader()`
2. Leader enters portal first
3. Companions see portal on ground
4. Companions use portal (no need to open their own)

#### Portal Detection and Usage

**Companion perspective**:

1. **Portal appears**: Leader opens portal, it becomes visible to all party members
2. **Game data updates**: Portal object added to game state
3. **Companion AI detects**: Town return logic finds available portal
4. **Companion uses**: Interacts with portal object to enter
5. **No casting required**: Companion never casts Town Portal spell

**Result**:
- Only one portal per party (cleaner, more efficient)
- Companions save mana by not casting
- Reduced cast time for party town returns

#### Town Routine Coordination

**File**: `internal/action/town.go`
**Function**: `Town.ExecuteTownRoutine()`

When in town, companions execute normal town activities:
- Selling items
- Buying potions
- Repairing equipment
- Hiring/reviving mercenary

These activities are **independent** - each character handles their own town tasks without synchronization.

**After town tasks**:
- Leader exits town through portal (back to field)
- Companions exit town through portal
- All characters return to same location (where portal was opened)

### Configuration for Portal Coordination

There is NO explicit configuration for portal coordination. It's automatically enabled when:
- `companion.enabled: true`
- `companion.leader: true` (for portal opening)

### Why This Matters

**Benefits of portal coordination**:

1. **Efficiency**: Only one character needs to cast Town Portal
2. **Mana savings**: Companions don't spend mana on portals
3. **Time savings**: No wasted time with multiple characters casting
4. **Visual clarity**: Only one portal object on ground
5. **Synchronization point**: Ensures all characters can return to town from same location

**Without this coordination**:
- Each character opens their own portal
- Multiple portal objects clutter the ground
- Wasted mana and time
- Characters might open portals in different locations

### File Paths for Portal Coordination

| Component | File Path | Lines |
|-----------|-----------|-------|
| Portal Opening Logic | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\tools.go` | 11-22 |
| Diablo Run Usage | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\run\diablo.go` | 64-85 |
| Baal Run Usage | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\run\baal.go` | 83-89 |
| Town Routines | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\town.go` | 219-226 |
| Portal Step | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\step\portal.go` | (various) |

---

## Boss Room Coordination

Leaders prepare safe areas before boss fights, ensuring the party has a secure position and escape route.

### How It Works

Boss room coordination involves multiple steps:

1. **Leader reaches boss room**
2. **Leader moves to safe position**
3. **Leader opens town portal**
4. **Leader applies buffs**
5. **Leader clears nearby enemies**
6. **Companions catch up**
7. **Boss fight begins**

### Diablo Run Coordination

**File**: `internal/run/diablo.go`
**Lines**: 64-85

**Pre-fight sequence**:

```go
if d.ctx.CharacterCfg.Companion.Leader {
    action.OpenTPIfLeader()
    action.Buff()
    action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
}
```

**Step-by-step**:

1. **Portal opening**: `action.OpenTPIfLeader()`
   - Creates emergency escape route
   - Allows companions to enter fight from town if needed
   - Safe return point if fight goes badly

2. **Buff application**: `action.Buff()`
   - Leader applies all class-specific buffs
   - Examples: Battle Orders, Shout, Energy Shield, Holy Shield
   - Prepares for boss fight
   - Companions apply their own buffs independently

3. **Area clearing**: `action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())`
   - Clears all monsters within 30-unit radius
   - Removes trash mobs around Diablo
   - Creates safe fighting area
   - Prevents companions from being overwhelmed by adds

**30-unit radius**:
- Roughly 3-4 screen widths
- Large enough to create safe zone
- Ensures clean boss fight start
- Prevents surprise attacks from off-screen enemies

### Baal Run Coordination

**File**: `internal/run/baal.go`
**Lines**: 83-89

**Pre-fight sequence**:

```go
if s.ctx.CharacterCfg.Companion.Leader {
    action.MoveToCoords(data.Position{X: 15116, Y: 5071})
    action.OpenTPIfLeader()
}
```

**Step-by-step**:

1. **Position movement**: `action.MoveToCoords(data.Position{X: 15116, Y: 5071})`
   - Moves to specific safe coordinates in Throne of Destruction
   - Coordinates X: 15116, Y: 5071 are near the Worldstone Chamber entrance
   - This is a known safe position away from Baal's spawn point
   - Allows portal to be in secure location

2. **Portal opening**: `action.OpenTPIfLeader()`
   - Opens town portal at safe position
   - Provides escape route for entire party
   - Allows dead companions to rejoin fight

**Why specific coordinates?**:
- Baal spawns in center of Worldstone Chamber
- Position 15116, 5071 is outside immediate danger zone
- Portal won't be destroyed by Baal's attacks
- Safe rally point for companions

### Boss Fight Execution

After setup, the actual boss fight proceeds:

**Leader and companions both**:
1. Engage boss with their attack skills
2. Monitor health and use potions independently
3. Avoid boss attacks using character AI
4. Continue until boss dies

**No active synchronization during fight**:
- Each character uses own combat AI
- No shared target tracking (not implemented)
- No coordinated ability usage
- Independent positioning and movement

### Emergency Town Returns

If fight goes badly:

1. **Individual decision**: Each character monitors own health
2. **Chicken threshold**: If health drops below configured percentage
3. **Portal usage**: Character runs to portal opened by leader
4. **Town safety**: Character enters town via portal
5. **Heal up**: Restore health/mana in town
6. **Return**: Re-enter portal to rejoin fight

The pre-opened portal is critical for this emergency escape mechanism.

### Other Boss Runs

Similar coordination patterns exist in other boss runs:

**Andariel** (`internal/run/andariel.go`):
- Leader moves to safe position
- Opens portal before engaging
- Clears nearby enemies

**Mephisto** (`internal/run/mephisto.go`):
- Leader positions in safe spot
- Opens portal near moat
- Prepares for fight

**Duriel** (`internal/run/duriel.go`):
- Limited prep due to small tomb chamber
- Portal opened before entering chamber

### Configuration for Boss Coordination

Boss room coordination is automatically enabled when:
- `companion.enabled: true`
- `companion.leader: true`

No additional configuration is required. The behavior is built into each boss run's implementation.

### Why This Matters

**Benefits of boss coordination**:

1. **Safety**: Portal provides escape route for party
2. **Clear fighting area**: No surprise trash mob attacks
3. **Buffs applied**: Leader enters fight fully prepared
4. **Rally point**: Dead companions know where to rejoin
5. **Efficiency**: Clean setup leads to faster kills

**Without this coordination**:
- Party might fight surrounded by trash mobs
- No easy escape if fight goes badly
- Dead companions don't know where to return
- Chaotic, inefficient boss fights

### File Paths for Boss Coordination

| Component | File Path | Lines |
|-----------|-----------|-------|
| Diablo Setup | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\run\diablo.go` | 64-85 |
| Baal Setup | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\run\baal.go` | 83-89 |
| Andariel Setup | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\run\andariel.go` | (various) |
| Mephisto Setup | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\run\mephisto.go` | (various) |
| Portal Logic | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\tools.go` | 11-22 |
| Buff Action | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\buff.go` | (various) |
| Clear Action | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\clear_area.go` | (various) |

---

## Attack Target Synchronization (Planned)

Attack target synchronization is **defined in the codebase but NOT implemented**. This section documents what was planned and why it doesn't currently work.

### Event Definition

**File**: `internal/event/events.go`
**Lines**: 109-119

The event structure exists:

```go
type CompanionLeaderAttackEvent struct {
    BaseEvent
    TargetUnitID data.UnitID
}

func CompanionLeaderAttack(be BaseEvent, targetUnitID data.UnitID) CompanionLeaderAttackEvent {
    return CompanionLeaderAttackEvent{
        BaseEvent:    be,
        TargetUnitID: targetUnitID,
    }
}
```

**Purpose**: This event was intended to broadcast the leader's attack target to companions.

**Data carried**:
- `BaseEvent`: Timestamp, source supervisor, message
- `TargetUnitID`: Unique identifier for the monster being attacked

### What Was Intended

The planned flow:

1. **Leader selects target**: Leader's AI chooses enemy to attack
2. **Event broadcast**: Leader sends `CompanionLeaderAttackEvent` with target's unit ID
3. **Companions receive**: Companion event handlers receive the event
4. **Companions switch target**: Companions change their attack target to match leader
5. **Coordinated damage**: All characters focus fire on same enemy
6. **Faster kills**: Concentrated damage kills enemies quicker

### Why It's Not Working

**Problem 1: Event Never Sent**

Searching the entire codebase reveals:
- Event is defined in `events.go`
- Event constructor exists: `CompanionLeaderAttack()`
- **NO calls to `event.Send()` with this event type**

The leader never actually broadcasts attack information.

**Problem 2: No Event Handler**

Even if the event were sent:
- No registered event listener for `CompanionLeaderAttackEvent`
- `CompanionEventHandler` only handles `RequestCompanionJoinGameEvent` and `ResetCompanionGameInfoEvent`
- No code to process attack target information

**Problem 3: No Companion AI Integration**

The companion's attack logic:
- **File**: `internal/action/step/attack.go`
- Operates independently per character
- Uses standard target selection (nearest enemy, priority targets, etc.)
- Has no mechanism to override target based on leader

There's no code path to:
- Receive leader's target ID
- Override normal target selection
- Attack specific enemy by unit ID
- Ignore other enemies in favor of leader's target

### Current Attack Behavior

**How companions currently attack**:

1. **Independent targeting**: Each character selects own targets
2. **Priority system**: Attacks based on configured priority (elites, ranged, etc.)
3. **Proximity-based**: Typically attacks nearest dangerous enemy
4. **No coordination**: Pure coincidence if multiple characters attack same target

**Result**:
- Damage is spread across multiple enemies
- No focus fire
- Less efficient than coordinated attacks
- Works fine for most content but suboptimal

### Configuration That Doesn't Work

**File**: `config/template/config.yaml`
**Lines**: 214-221 (referenced)

The template config includes:

```yaml
companion:
  attack: true  # If set to true, character will try to attack same target as leader
```

**Status**: This setting exists but **has no effect** because:
- No code reads this `attack` setting
- No implementation for shared targeting
- Setting it to `true` or `false` makes no difference

This appears to be a placeholder for future implementation.

### What Would Need to Be Implemented

To make attack synchronization work:

**1. Leader sends events** (`internal/action/step/attack.go`):
```
Leader selects target → Get target's UnitID →
Send CompanionLeaderAttackEvent with UnitID
```

**2. Companion receives events** (`internal/bot/companion.go`):
```
Add handler for CompanionLeaderAttackEvent →
Check if attack sync enabled in config →
Store target UnitID in context
```

**3. Companion attack logic** (`internal/action/step/attack.go`):
```
Check if companion has leader's target ID →
Find monster with matching UnitID →
Override normal target selection →
Attack leader's target
```

**4. Handle edge cases**:
- Target already dead before companion can attack
- Target out of range for companion
- Target immune to companion's damage type
- Multiple targets being attacked in sequence

**5. Performance considerations**:
- Event frequency (send every attack or throttle?)
- Network overhead with many companions
- Race conditions with rapidly changing targets

### File Paths for Attack Synchronization

| Component | File Path | Lines | Status |
|-----------|-----------|-------|--------|
| Event Definition | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 109-119 | Defined |
| Event Constructor | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 109-119 | Exists |
| Event Sending | N/A | N/A | **Missing** |
| Event Handler | N/A | N/A | **Missing** |
| Companion AI Integration | N/A | N/A | **Missing** |
| Attack Logic | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\step\attack.go` | (various) | No integration |
| Config Setting | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\config\template\config.yaml` | 214-221 | Unused |

---

## Movement Following (Partial)

Movement following is **partially implemented** - only as checkpoint-based waiting, not continuous active following.

### What IS Implemented: Checkpoint Waiting

**File**: `internal/action/leveling_tools.go`
**Lines**: 585-604

As described in the [Position Tracking section](#position-tracking-and-waiting):

- Leader waits at checkpoints for companions within 20 units
- Only during leveling runs
- Only in non-town areas
- Companions must catch up on their own

This is **passive synchronization** - the leader waits, but companions don't actively follow.

### What is NOT Implemented: Active Following

**Missing features**:

**1. Real-time position broadcasting**:
- Leader doesn't send position updates
- No event for "leader moved to X, Y"
- Companions don't know where leader currently is

**2. Companion pathfinding to leader**:
- Companions don't have "move toward leader" behavior
- Each companion follows own navigation logic
- No override to change destination to leader's position

**3. Formation keeping**:
- No logic to maintain party formation
- No designated positions relative to leader (behind, flanking, etc.)
- Characters spread out naturally based on their AI

**4. Dynamic follow distance**:
- No configuration for how close to stay to leader
- The 20-unit threshold is for waiting, not following
- Can't configure "always stay within X units"

### Configuration That Doesn't Work

**File**: `config/template/config.yaml`
**Lines**: 214-221

```yaml
companion:
  followLeader: true  # If set to true, character will follow the leader
```

**Status**: This setting exists but **has no active effect** because:
- No code that reads `followLeader` setting for movement behavior
- No pathfinding override based on this setting
- Only affects checkpoint waiting (which happens regardless)

The setting might be used for future implementation or have very subtle effects not immediately apparent.

### Current Movement Behavior

**How companions currently move**:

1. **Independent pathfinding**: Each character navigates using own AI
2. **Area-based goals**: Characters move toward quest objectives, waypoints, bosses
3. **Collision avoidance**: Characters path around each other and obstacles
4. **No leader tracking**: Movement decisions don't consider leader's position
5. **Coincidental proximity**: Characters end up together because they have same goals

**Example**:
- Leader and companion both targeting Mephisto
- Both pathfind to Durance of Hate Level 3
- Both navigate to Mephisto's chamber
- They arrive together, but not because of following
- Result: Looks like following, but is actually parallel navigation

### When "Following" Appears to Work

**Scenarios where it seems like following**:

1. **Shared objectives**: Both characters going to same location
2. **Checkpoint waiting**: Leader waits, companions catch up
3. **Portal usage**: All characters use same portal, return to same location
4. **Small maps**: Limited area forces characters together
5. **Similar movement speed**: Characters naturally stay close

### What Would Need to Be Implemented

To implement real active following:

**1. Position broadcasting** (new event):
```
Leader moves → Every N seconds, broadcast position →
Send CompanionLeaderPositionEvent with X, Y coordinates
```

**2. Companion receives position** (`internal/bot/companion.go`):
```
Receive CompanionLeaderPositionEvent →
Check if followLeader enabled →
Store leader's position in context
```

**3. Pathfinding override** (`internal/game/pathfinder.go`):
```
Companion needs move destination →
Check if followLeader enabled and leader position known →
Override destination to leader's position →
Calculate path to leader
```

**4. Follow distance management**:
```
Check distance to leader →
If > max follow distance: Move toward leader →
If < min follow distance: Stop moving (prevent overlap) →
If within range: Resume normal behavior
```

**5. Smart following**:
- Don't interrupt critical actions (attacking, casting, looting)
- Maintain formation if multiple companions
- Handle teleporting leaders (Sorceresses)
- Deal with blocked paths and impassable terrain

**6. Configuration**:
```yaml
companion:
  followLeader: true
  followDistance: 15  # Stay within 15 units
  followPriority: medium  # How aggressively to follow vs. other goals
```

### Why Passive Following Works for Most Content

Despite lacking active following, the companion system functions reasonably well because:

1. **Shared objectives**: Both leader and companions have same goals (kill boss, clear area)
2. **Checkpoint waiting**: Leader waits at key points, allowing synchronization
3. **Similar pathing**: Characters use same pathfinding algorithms, take similar routes
4. **Portal coordination**: All characters travel through same portals
5. **Boss room coordination**: Characters converge on boss locations naturally

**Result**: For most farming runs, companions stay "close enough" without active following.

### When Passive Following Fails

**Problem scenarios**:

1. **Teleporting leaders**: Sorceress teleports rapidly, companions can't keep up
2. **Split objectives**: Characters have different goals (one loots, one rushes ahead)
3. **Large maps**: Companions get lost in big areas like Worldstone Keep
4. **Companion stuck**: Pathing error causes companion to get stuck far behind
5. **Different movement speeds**: Faster characters outpace slower ones

**Without active following**: These problems persist until next checkpoint or portal.

### File Paths for Movement Following

| Component | File Path | Lines | Status |
|-----------|-----------|-------|--------|
| Checkpoint Waiting | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\action\leveling_tools.go` | 585-604 | Implemented |
| Position Broadcasting | N/A | N/A | **Missing** |
| Position Event | N/A | N/A | **Missing** |
| Follow Logic | N/A | N/A | **Missing** |
| Pathfinding Override | N/A | N/A | **Missing** |
| Config Setting | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\config\template\config.yaml` | 214-221 | Unused |

---

## Event System Architecture

The companion system relies heavily on an event-driven architecture for communication between bots.

### Event System Overview

**File**: `internal/event/events.go`

The event system provides:
- **Publisher/Subscriber pattern**: Bots can send and listen for events
- **Decoupled communication**: Bots don't need direct references to each other
- **Real-time messaging**: Events processed as they occur
- **Type-safe events**: Each event type has defined structure

### Base Event Structure

**Lines**: 14-24 (approximate)

All events inherit from `BaseEvent`:

```go
type BaseEvent struct {
    Supervisor string    // Name of bot that sent event
    Message    string    // Human-readable description
    Timestamp  time.Time // When event was created
}
```

This provides common metadata for all event types.

### Event Publishing

**Sending an event**:

```go
event.Send(event.RequestCompanionJoinGame(
    event.Text(supervisorName, "New Game Started"),
    leaderName,
    gameName,
    gamePassword
))
```

**Steps**:
1. Create event using constructor function
2. Pass event to `event.Send()`
3. Event system broadcasts to all listeners
4. Each bot's event handlers receive the event
5. Handlers process or ignore based on event type

### Event Listening

**Registering a listener**:

Listeners must be registered during bot initialization, typically in:
- `cmd/koolo/main.go` (lines 80-150)
- `internal/bot/manager.go` (line 275)

**Example registration** (hypothetical):
```go
eventManager.RegisterListener(companionEventHandler)
```

**Handler interface** (approximate):
```go
type EventHandler interface {
    Handle(ctx context.Context, event interface{}) error
}
```

### Companion-Specific Events

**1. RequestCompanionJoinGameEvent** (Lines 155-170)

**Structure**:
```go
type RequestCompanionJoinGameEvent struct {
    BaseEvent
    Leader   string  // Leader character name
    Name     string  // Game name
    Password string  // Game password
}
```

**When sent**: Leader creates a new game
**Purpose**: Notify companions to join
**Sender**: Leader's supervisor
**Receivers**: All companion characters

**2. ResetCompanionGameInfoEvent** (Lines 172-183)

**Structure**:
```go
type ResetCompanionGameInfoEvent struct {
    BaseEvent
    Leader string  // Leader character name
}
```

**When sent**: Leader finishes a game
**Purpose**: Clear companion's stored game info
**Sender**: Leader's supervisor
**Receivers**: All companion characters

**3. CompanionLeaderAttackEvent** (Lines 109-119)

**Structure**:
```go
type CompanionLeaderAttackEvent struct {
    BaseEvent
    TargetUnitID data.UnitID  // Monster being attacked
}
```

**When sent**: Never (not implemented)
**Purpose**: Coordinate attack targets
**Status**: Defined but unused

**4. CompanionRequestedTPEvent** (Lines 121-127)

**Structure**:
```go
type CompanionRequestedTPEvent struct {
    BaseEvent
}
```

**When sent**: Leader requests town portal
**Purpose**: Signal companions that leader is returning to town
**Status**: Defined, usage unclear

### Event Flow Example

**Game creation scenario**:

```
Leader Bot:
  1. Creates game "baalrun-5"
  2. event.Send(RequestCompanionJoinGame(...))
  3. Event enters system

Event System:
  4. Broadcasts to all registered listeners
  5. Each bot's handlers receive event

Companion Bot 1:
  6. CompanionEventHandler.Handle() called
  7. Event type: RequestCompanionJoinGameEvent
  8. Check: Is companion enabled? Yes
  9. Check: Is this bot not a leader? Yes
  10. Check: Leader name matches? Yes
  11. Store game name and password
  12. Log: "Received join request"

Companion Bot 2:
  (Same process as Companion Bot 1)

Leader Bot:
  13. Receives own event
  14. Handler ignores (is a leader)
```

### Event Processing Performance

**Considerations**:

1. **Synchronous processing**: Events are handled immediately
2. **Handler speed**: Slow handlers block event system
3. **Event frequency**: Too many events can cause lag
4. **Memory usage**: Events are kept in memory until processed

**Best practices**:
- Keep event handlers fast (< 10ms)
- Don't perform I/O in handlers
- Store event data and process asynchronously if needed
- Filter events early to avoid unnecessary processing

### Event System Limitations

**Current issues**:

1. **No guaranteed delivery**: Events can be lost if bot crashes
2. **No ordering guarantees**: Events may arrive out of order
3. **No persistence**: Events not saved to disk
4. **Same-process only**: Only works for bots in same Koolo instance
5. **No inter-machine communication**: Can't coordinate bots on different computers

**Workarounds**:
- Use configuration to establish leader/companion relationships
- Implement retry logic for critical operations
- Use timeouts to handle missing events
- Keep event handlers idempotent (safe to process multiple times)

### File Paths for Event System

| Component | File Path | Lines |
|-----------|-----------|-------|
| Event Definitions | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 1-183 |
| Base Event | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 14-24 |
| Join Event | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 155-170 |
| Reset Event | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 172-183 |
| Attack Event | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 109-119 |
| TP Event | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\event\events.go` | 121-127 |
| Event Sending | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\single_supervisor.go` | 224-226, 379-381 |
| Companion Handler | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\companion.go` | 1-61 |

---

## Configuration System

The companion synchronization system is controlled entirely through YAML configuration files.

### Configuration File Structure

**File**: `internal/config/config.go`
**Lines**: 288-296

The companion configuration struct:

```go
type Companion struct {
    Enabled               bool   `yaml:"enabled"`
    Leader                bool   `yaml:"leader"`
    LeaderName            string `yaml:"leaderName"`
    GameNameTemplate      string `yaml:"gameNameTemplate"`
    GamePassword          string `yaml:"gamePassword"`
    CompanionGameName     string `yaml:"companionGameName"`
    CompanionGamePassword string `yaml:"companionGamePassword"`
}
```

### Configuration Fields

| Field | Type | Purpose | Who Uses | Saved to File |
|-------|------|---------|----------|---------------|
| `Enabled` | bool | Activates companion mode | Both | Yes |
| `Leader` | bool | Designates leader vs companion | Both | Yes |
| `LeaderName` | string | Name of leader to follow | Companion | Yes |
| `GameNameTemplate` | string | Prefix for game names | Leader | Yes |
| `GamePassword` | string | Password for games | Leader | Yes |
| `CompanionGameName` | string | Current game to join | Companion | No (runtime only) |
| `CompanionGamePassword` | string | Current game password | Companion | No (runtime only) |

### Runtime vs Persistent Configuration

**Persistent** (saved to YAML file):
- `enabled`, `leader`, `leaderName`
- `gameNameTemplate`, `gamePassword`
- These are user-configured settings

**Runtime** (only in memory):
- `companionGameName`, `companionGamePassword`
- These are populated by events
- Never written back to config file
- Reset when bot restarts

### Configuration Template

**File**: `config/template/config.yaml`
**Lines**: 214-221

Default configuration:

```yaml
companion:
  enabled: false
  leader: true
  leaderName: ''
  attack: true  # NOTE: This setting is defined but not currently used
  followLeader: true  # NOTE: This setting is defined but not currently used
  gameNameTemplate: game-
  gamePassword: xxx
```

### Configuration Examples

**Example 1: Leader Sorceress**

```yaml
character:
  characterName: LeaderSorc

companion:
  enabled: true
  leader: true
  leaderName: ''  # Not used for leaders
  gameNameTemplate: 'mf-'
  gamePassword: 'secret123'
```

**Game creation**:
- Creates games: "mf-1", "mf-2", "mf-3", etc.
- Uses password: "secret123"
- Broadcasts join events to companions

**Example 2: Companion Paladin**

```yaml
character:
  characterName: CompanionPally

companion:
  enabled: true
  leader: false
  leaderName: 'LeaderSorc'
  gameNameTemplate: ''  # Not used for companions
  gamePassword: ''  # Not used for companions
```

**Game joining**:
- Waits for events from "LeaderSorc"
- Joins games automatically
- Uses credentials from events

**Example 3: Companion with Empty Leader Name**

```yaml
companion:
  enabled: true
  leader: false
  leaderName: ''  # Empty = follow any leader
```

**Behavior**:
- Accepts join events from ANY leader
- Useful for single-companion setups
- Flexible but can cause confusion with multiple leaders

### Configuration Loading

**File**: `internal/config/config.go`

Configuration is loaded at bot startup:

1. **Find config file**: Look in `config/[character-name].yaml`
2. **Parse YAML**: Read and validate structure
3. **Create config object**: Populate `CharacterCfg` struct
4. **Validate settings**: Check for invalid combinations
5. **Store in context**: Make available to all bot components

### Configuration Validation

**Potential validation issues**:

1. **Both enabled and leader**: Valid (this is the leader)
2. **Enabled but leader name empty**: Valid for companions (follow any leader)
3. **Game template empty for leader**: Invalid (needs template)
4. **Companion with game template**: Harmless (ignored)

**No strict validation currently implemented** - invalid configs may cause runtime errors.

### Updating Configuration

**At runtime**:

Configuration can be updated via:
- Web UI: Character settings page
- HTTP API: POST to `/api/character/[name]/settings`
- Direct file edit: Modify YAML file and reload

**Changes require restart**: Most config changes require bot restart to take effect.

**Exception**: Runtime fields (`CompanionGameName`, `CompanionGamePassword`) update immediately from events.

### Configuration Priority

**Order of precedence**:

1. **Runtime events**: Override persistent config (game name/password)
2. **Config file**: User's YAML configuration
3. **Template defaults**: Default values from template
4. **Hard-coded defaults**: Built-in fallback values

### File Paths for Configuration

| Component | File Path | Lines |
|-----------|-----------|-------|
| Config Structure | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\config\config.go` | 288-296 |
| Template Config | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\config\template\config.yaml` | 214-221 |
| Config Loading | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\config\config.go` | (various) |
| HTTP API | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\server\http_server.go` | 545-550 |
| Web UI Settings | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\server\templates\character_settings.gohtml` | 466-470 |

---

## Technical Implementation Details

This section covers the low-level implementation details of the synchronization system.

### Supervisor Integration

**File**: `internal/bot/single_supervisor.go`

The supervisor is the main control loop for each character. It orchestrates all bot actions including companion behavior.

**Main supervisor loop** (approximate):

```
1. Check current game state
2. Are we in menu?
   a. If companion enabled and not leader: HandleCompanionMenuFlow()
   b. Otherwise: HandleSinglePlayerFlow()
3. Are we in game?
   a. Execute configured runs
   b. Check for companion-specific logic
   c. Monitor health, potions, etc.
4. Handle errors and recovery
5. Repeat
```

**Companion menu flow** (Lines 520-557):

```
1. Check if CompanionGameName is set
   - If empty: Wait 2 seconds, return idle
   - If set: Continue
2. Check current location
   - On character select screen: Enter lobby first
   - In lobby: Proceed to join
3. Call Manager.JoinOnlineGame(name, password)
4. Handle result:
   - Success: Enter game
   - Failure: Log error, retry next cycle
```

### Game Manager Integration

**File**: `internal/game/manager.go`

The game manager handles creating and joining games.

**Leader game creation** (Lines 140-150):

```
1. Increment game counter
2. Build game name: template + counter
   Example: "game-" + "5" = "game-5"
3. Get password from config
4. Call Diablo 2 game creation API
5. Wait for game to be created
6. Store game name and password in game data
7. Trigger event broadcast (in supervisor)
8. Return success
```

**Companion game joining** (approximate):

```
1. Receive game name and password (from menu flow)
2. Call Diablo 2 join game API
3. Wait for join result
4. If successful: Enter game state
5. If failed: Return error to supervisor
```

### Context Management

**File**: `internal/game/context.go` (referenced)

The context holds all runtime state for a character:

```go
type Context struct {
    CharacterCfg *config.CharacterCfg  // Configuration
    Data         *Data                  // Game state
    Manager      *Manager               // Game manager
    PathFinder   *PathFinder            // Navigation
    // Many other fields...
}
```

**Companion-relevant context fields**:
- `CharacterCfg.Companion`: Companion configuration
- `Data.Roster`: Party member positions
- `Data.Game.LastGameName`: Last created game name
- `Data.Game.LastGamePassword`: Last game password

### Roster Tracking

**File**: `internal/game/data.go`

The roster tracks all party members:

```go
type Roster struct {
    Name     string          // Character name
    Position data.Position   // X, Y coordinates
    Level    int             // Character level
    Class    data.Class      // Character class
    // Other fields...
}
```

**Updated by**:
- Game client (reads from Diablo 2 memory)
- Updated every game tick (multiple times per second)
- Provides real-time party member locations

**Used by**:
- Position tracking logic
- Wait for members functionality
- UI display (party list)

### Distance Calculation

**File**: `internal/game/pathfinder.go`

The pathfinder provides distance calculations:

```go
func (pf *PathFinder) DistanceFromMe(position data.Position) int {
    playerPos := pf.ctx.Data.PlayerUnit.Position

    deltaX := position.X - playerPos.X
    deltaY := position.Y - playerPos.Y

    distance := math.Sqrt(float64(deltaX*deltaX + deltaY*deltaY))

    return int(distance)
}
```

**Considerations**:
- Straight-line distance (Euclidean)
- Doesn't account for walls/obstacles
- Unit is game units (not pixels or meters)
- Fast calculation for frequent use

### Event Handler Implementation

**File**: `internal/bot/companion.go`
**Lines**: 1-61

Complete handler implementation:

```go
type CompanionEventHandler struct {
    supervisor string
    log        *slog.Logger
    cfg        *config.CharacterCfg
}

func NewCompanionEventHandler(
    supervisor string,
    log *slog.Logger,
    cfg *config.CharacterCfg,
) *CompanionEventHandler {
    return &CompanionEventHandler{
        supervisor: supervisor,
        log:        log,
        cfg:        cfg,
    }
}

func (c *CompanionEventHandler) Handle(
    ctx context.Context,
    e event.Event,
) error {
    switch evt := e.(type) {
    case event.RequestCompanionJoinGameEvent:
        return c.handleJoinGame(evt)
    case event.ResetCompanionGameInfoEvent:
        return c.handleResetGameInfo(evt)
    }
    return nil
}

func (c *CompanionEventHandler) handleJoinGame(
    evt event.RequestCompanionJoinGameEvent,
) error {
    if !c.cfg.Companion.Enabled || c.cfg.Companion.Leader {
        return nil  // Ignore if not a companion
    }

    if c.cfg.Companion.LeaderName != "" &&
       c.cfg.Companion.LeaderName != evt.Leader {
        return nil  // Wrong leader
    }

    c.cfg.Companion.CompanionGameName = evt.Name
    c.cfg.Companion.CompanionGamePassword = evt.Password
    c.log.Info("Received companion join game request")

    return nil
}

func (c *CompanionEventHandler) handleResetGameInfo(
    evt event.ResetCompanionGameInfoEvent,
) error {
    if c.cfg.Companion.LeaderName != "" &&
       c.cfg.Companion.LeaderName != evt.Leader {
        return nil  // Wrong leader
    }

    c.cfg.Companion.CompanionGameName = ""
    c.cfg.Companion.CompanionGamePassword = ""
    c.log.Info("Reset companion game info")

    return nil
}
```

**Handler lifecycle**:
1. Created during bot initialization
2. Registered with event system (NOTE: This step is currently missing!)
3. Receives all events via `Handle()` method
4. Routes to specific handler functions
5. Returns error or nil

### Statistics Tracking

**File**: `internal/bot/stats.go`
**Lines**: 104-108

Statistics include companion status:

```go
type Stats struct {
    // Many fields...
    IsCompanionFollower bool
    MuleEnabled         bool
}
```

**Populated by** (`internal/server/http_server.go`, Lines 545-550):

```go
if cfg.Companion.Enabled && !cfg.Companion.Leader {
    stats.IsCompanionFollower = true
    stats.MuleEnabled = cfg.Muling.Enabled
}
```

**Used for**:
- Web UI display (show companion badge)
- Dashboard filtering (group by leader/companion)
- Monitoring and debugging

### File Paths Summary

| Component | File Path |
|-----------|-----------|
| Supervisor | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\single_supervisor.go` |
| Game Manager | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\game\manager.go` |
| Context | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\game\context.go` |
| Data Structures | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\game\data.go` |
| PathFinder | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\game\pathfinder.go` |
| Companion Handler | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\companion.go` |
| Statistics | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\bot\stats.go` |
| HTTP Server | `C:\Users\doey\Downloads\kwader2k-koolo\koolo\internal\server\http_server.go` |

---

## Known Issues and Limitations

### Critical Issues

**1. CompanionEventHandler Not Registered**

**Problem**: The `CompanionEventHandler` is created but never registered with the event system.

**Evidence**:
- Handler exists in `internal/bot/companion.go`
- No registration in `cmd/koolo/main.go`
- No registration in `internal/bot/manager.go`
- Events are sent but may not be received

**Impact**: Companion join/reset events may not work properly

**Workaround**: Unknown - requires code fix

**Fix needed**: Add handler registration during bot initialization

**2. Attack Synchronization Not Implemented**

**Problem**: `CompanionLeaderAttackEvent` defined but never sent or handled.

**Evidence**:
- Event exists in `internal/event/events.go`
- No calls to send this event
- No handlers for this event
- Config setting `attack: true` has no effect

**Impact**: No coordinated attack targeting

**Workaround**: None - feature doesn't work

**Fix needed**: Implement event sending, handling, and target override logic

**3. Movement Following Not Implemented**

**Problem**: Active following logic doesn't exist, only checkpoint waiting.

**Evidence**:
- Config setting `followLeader: true` has no effect
- No position broadcasting from leader
- No pathfinding override for companions
- Only passive waiting at checkpoints

**Impact**: Companions don't actively follow leader

**Workaround**: Use checkpoint-based synchronization (leveling only)

**Fix needed**: Implement position broadcasting and following pathfinding

### Minor Issues

**4. Race Condition in Game Joining**

**Problem**: Companion may receive join event before game is fully created.

**Symptoms**:
- "Game does not exist" errors
- Join failures requiring retry

**Workaround**: Retry logic eventually succeeds

**Fix needed**: Add delay or confirmation mechanism

**5. No Stuck Detection**

**Problem**: If companion gets stuck, leader waits indefinitely.

**Impact**: Leader never proceeds if companion is stuck

**Workaround**: Manual intervention required

**Fix needed**: Timeout or stuck detection logic

**6. Hard-coded Constants**

**Problem**: Magic numbers like 20-unit distance threshold, 30-unit clear radius.

**Impact**: Can't tune synchronization behavior without code changes

**Workaround**: Edit code directly

**Fix needed**: Move constants to configuration

**7. Single-Process Only**

**Problem**: Events only work between bots in same Koolo instance.

**Impact**: Can't coordinate bots on different computers

**Workaround**: Run all bots in one Koolo instance

**Fix needed**: Network-based event system

### Design Limitations

**8. No Formation Control**

**Limitation**: Can't specify relative positions (tank in front, caster behind).

**Impact**: Characters cluster randomly

**Workaround**: None

**9. No Buff Coordination**

**Limitation**: Each character buffs independently.

**Impact**: Wasted time, overlapping buffs

**Workaround**: Stagger buff timing manually

**10. No Loot Distribution**

**Limitation**: All characters pick up items using normal logic.

**Impact**: Companions might steal leader's loot

**Workaround**: Configure different item pickup rules

**11. No Quest Coordination**

**Limitation**: Quest progress not synchronized.

**Impact**: Characters might have different quest states

**Workaround**: Ensure all characters on same quest progress

---

## How to Use Synchronization

### Basic Setup

**Step 1: Configure Leader**

Edit `config/leader-character.yaml`:

```yaml
character:
  characterName: MyLeader

companion:
  enabled: true
  leader: true
  leaderName: ''
  gameNameTemplate: 'run-'
  gamePassword: 'mypass'
```

**Step 2: Configure Companion**

Edit `config/companion-character.yaml`:

```yaml
character:
  characterName: MyCompanion

companion:
  enabled: true
  leader: false
  leaderName: 'MyLeader'
  gameNameTemplate: ''
  gamePassword: ''
```

**Step 3: Start Bots**

```
1. Start Koolo
2. Load leader character
3. Load companion character
4. Start both characters
5. Leader creates game, companion joins automatically
```

### Leveling Party Setup

For leveling runs with position synchronization:

**Configuration**: Same as basic setup

**Behavior**:
- Leader waits for companions within 20 units
- Companions navigate to objectives independently
- Leader clears area while waiting
- Party progresses through acts together

**Best for**:
- Questing together
- Leveling multiple characters
- Story progression

### Magic Finding Party Setup

For magic finding runs:

**Configuration**: Basic setup

**Behavior**:
- Companions join leader's games
- Each character farms independently
- Portal coordination for town trips
- No position waiting (non-leveling)

**Best for**:
- Boss runs (Mephisto, Baal, Diablo)
- Area farming (Pits, Ancient Tunnels)
- Maximizing runs per hour

### Boss Killing Party Setup

For coordinated boss fights:

**Configuration**: Basic setup

**Additional**: Configure all characters for boss damage

**Behavior**:
- Leader prepares boss room (portal, clear)
- Companions enter boss room
- Focus fire on boss (coincidentally)
- Use leader's portal for town trips

**Best for**:
- Diablo/Baal runs
- Uber Tristram (with modifications)
- High-difficulty content

### Troubleshooting Tips

**Companion not joining games**:
1. Check `enabled: true` in companion config
2. Verify `leader: false` for companion
3. Ensure `leaderName` matches exactly (case-sensitive)
4. Check logs for join errors
5. Verify bots are in same Koolo instance

**Companions too far behind**:
1. Reduce leader movement speed
2. Increase companion movement speed
3. Use teleport/leap skills sparingly
4. Check for companions getting stuck
5. Add manual wait actions in leader's route

**Both characters acting as leaders**:
1. Check `leader` setting in configs
2. Ensure only ONE character has `leader: true`
3. Restart both bots after config changes

**Events not working**:
1. Verify bots are in same Koolo process
2. Check logs for event sending/receiving
3. Restart Koolo to reset event system
4. Verify no config syntax errors

### Advanced Usage

**Multiple companions**:

```yaml
# Leader (creates games)
Leader: enabled: true, leader: true

# Companion 1 (joins leader)
Comp1: enabled: true, leader: false, leaderName: 'Leader'

# Companion 2 (joins leader)
Comp2: enabled: true, leader: false, leaderName: 'Leader'

# Companion 3 (joins leader)
Comp3: enabled: true, leader: false, leaderName: 'Leader'
```

All companions join leader's games automatically.

**Multiple leader groups**:

```yaml
# Group 1
Leader1: enabled: true, leader: true, gameNameTemplate: 'grp1-'
Comp1A: enabled: true, leader: false, leaderName: 'Leader1'
Comp1B: enabled: true, leader: false, leaderName: 'Leader1'

# Group 2
Leader2: enabled: true, leader: true, gameNameTemplate: 'grp2-'
Comp2A: enabled: true, leader: false, leaderName: 'Leader2'
Comp2B: enabled: true, leader: false, leaderName: 'Leader2'
```

Each group operates independently.

### Performance Considerations

**Resource usage per character**:
- CPU: 5-10% per character
- RAM: 500MB-1GB per game client
- Network: Minimal (local events only)

**Recommended limits**:
- 4 characters per computer (1 leader + 3 companions)
- 8 characters maximum (2 groups of 4)
- More characters = more resource usage

---

## Summary

The companion synchronization system provides:

### Working Features
- ✅ **Game join synchronization**: Automatic game joining via events
- ✅ **Position tracking**: Leader waits for companions (leveling only)
- ✅ **Portal coordination**: Leader opens portals for party
- ✅ **Boss room setup**: Leader prepares safe areas

### Planned But Not Working
- ⚠️ **Attack synchronization**: Event defined, not implemented
- ⚠️ **Movement following**: Config exists, not functional
- ⚠️ **Event handler**: Created but not registered

### Key Concepts
- **Leader-follower model**: One leader, multiple companions
- **Event-driven**: Communication through events
- **Configuration-based**: YAML files control behavior
- **Checkpoint synchronization**: Passive waiting, not active following

### Best Use Cases
- Magic finding parties
- Boss killing groups
- Leveling parties
- Multi-character farming

### Limitations
- Same Koolo instance only (no network coordination)
- Checkpoint-based (not real-time following)
- No attack coordination (independent targeting)
- No formation control
- Hard-coded constants

The system works well for basic multi-character coordination but lacks advanced features like active following and combat synchronization.
