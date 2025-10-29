# Companion System - Complete Documentation

## Table of Contents
1. [Overview](#overview)
2. [What is the Companion System?](#what-is-the-companion-system)
3. [Companion Roles](#companion-roles)
4. [Configuration](#configuration)
5. [How the System Works](#how-the-system-works)
6. [Event-Driven Communication](#event-driven-communication)
7. [Game Creation and Joining Flow](#game-creation-and-joining-flow)
8. [Companion Behavior](#companion-behavior)
9. [Mercenary vs Companion System](#mercenary-vs-companion-system)
10. [Web Interface](#web-interface)
11. [Technical Implementation](#technical-implementation)
12. [Troubleshooting](#troubleshooting)

---

## Overview

The companion system in Koolo is a multiplayer coordination feature that allows multiple bot characters to work together in the same game. One character acts as a "leader" who creates games, while other characters act as "companions" or "followers" who automatically join the leader's games.

Think of it like a party system where one player hosts the game and others join automatically without needing to manually coordinate game names and passwords.

---

## What is the Companion System?

The companion system enables you to run multiple Diablo 2 characters simultaneously where:

- **One character is the leader**: This character creates new games using a template name (like "game-1", "game-2", etc.)
- **Other characters are companions**: These characters automatically join the games that the leader creates
- **Communication is automatic**: The leader broadcasts game information to companions through an event system
- **Synchronization is optional**: Companions can optionally attack the same targets as the leader and follow the leader's movement

**Important Note**: The companion system is different from the mercenary system. Mercenaries are NPC followers that you hire in-game (like Act 2 mercenaries with auras). The companion system is for coordinating multiple player characters that you're running as bots.

---

## Companion Roles

### Leader Role

When a character is configured as a leader:

1. **Creates games** using a template name and counter (example: "game-1", "game-2", "game-3")
2. **Broadcasts game information** to all companion bots when a new game is created
3. **Signals game completion** so companions know not to join old games
4. **Runs normally** with whatever run configuration you've set up (Mephisto runs, Baal runs, etc.)

**Configuration for Leader:**
```yaml
companion:
  enabled: true
  leader: true
  leaderName: ''  # Not used for leaders
  gameNameTemplate: game-
  gamePassword: mypassword
```

### Companion/Follower Role

When a character is configured as a companion:

1. **Waits for leader's game creation events** instead of creating its own games
2. **Automatically joins** the leader's game when notified
3. **Can attack the same targets** as the leader (optional)
4. **Can follow the leader's movement** (optional)
5. **Returns to town** when needed (no potions, mercenary dies, equipment broken)

**Configuration for Companion:**
```yaml
companion:
  enabled: true
  leader: false
  leaderName: 'MyLeaderCharacterName'  # Name of the leader to follow
  attack: true  # Attack same targets as leader
  followLeader: true  # Follow leader's position
  gameNameTemplate: ''  # Not used for companions
  gamePassword: ''  # Not used for companions (receives from leader)
```

---

## Configuration

### Configuration File Location

The companion settings are in your character's configuration file, typically located at:
- `config/[character-name].yaml`

### Configuration Options

| Option | Type | Description | Leader Uses | Companion Uses |
|--------|------|-------------|-------------|----------------|
| `enabled` | boolean | Enables companion mode | Yes | Yes |
| `leader` | boolean | Sets character as leader or follower | Yes (true) | Yes (false) |
| `leaderName` | string | Name of leader character to follow | No | Yes |
| `attack` | boolean | Whether to attack same targets as leader | No | Yes |
| `followLeader` | boolean | Whether to follow leader's movement | No | Yes |
| `gameNameTemplate` | string | Prefix for game names (e.g., "game-") | Yes | No |
| `gamePassword` | string | Password for created games | Yes | No |

### Example Configurations

**Leader Configuration:**
```yaml
companion:
  enabled: true
  leader: true
  leaderName: ''
  attack: false
  followLeader: false
  gameNameTemplate: baalrun-
  gamePassword: secretpass123
```

**Companion Configuration:**
```yaml
companion:
  enabled: true
  leader: false
  leaderName: 'MySorceress'
  attack: true
  followLeader: true
  gameNameTemplate: ''
  gamePassword: ''
```

---

## How the System Works

### Basic Flow

1. **Startup**: Both leader and companion bots start up
2. **Leader Creates Game**: Leader creates a game named "game-1" with password "mypass"
3. **Leader Broadcasts**: Leader sends an event to all companions: "Join game-1 with password mypass"
4. **Companions Receive**: All companion bots receive this event
5. **Companions Filter**: Each companion checks if the event is from their configured leader
6. **Companions Store Info**: Matching companions save the game name and password
7. **Companions Join**: On their next menu cycle, companions join the stored game
8. **Game Runs**: Leader and companions complete the run together
9. **Game Ends**: Leader broadcasts a "game finished" event
10. **Reset**: Companions clear stored game info and wait for next game
11. **Repeat**: Leader creates "game-2" and the cycle repeats

### Timing Considerations

- **Companion wait time**: Companions may need to wait a few seconds in the menu for the leader to fully create the game
- **Game creation delay**: There's typically a 5-10 second delay between the leader creating a game and it being joinable
- **Event latency**: Events are processed in real-time, but there can be slight delays based on system performance

---

## Event-Driven Communication

The companion system uses an internal event system to communicate between bots. This means the bots don't need to be on the same computer or even share configuration files - they just need to be running in the same Koolo instance and able to send/receive events.

### Key Events

#### 1. RequestCompanionJoinGameEvent

**Sent by**: Leader
**Sent when**: Leader successfully creates a new game
**Contains**:
- Leader character name
- Game name
- Game password

**Purpose**: Tells companions "I just created a game, here's how to join it"

#### 2. ResetCompanionGameInfoEvent

**Sent by**: Leader
**Sent when**: Game finishes and leader returns to menu
**Contains**:
- Leader character name

**Purpose**: Tells companions "Don't try to join my old game anymore, I'm done with it"

#### 3. CompanionLeaderAttackEvent

**Sent by**: Leader
**Sent when**: Leader attacks a target
**Contains**:
- Target unit ID

**Purpose**: Tells companions what enemy to attack for synchronized combat

#### 4. CompanionRequestedTPEvent

**Sent by**: Leader
**Sent when**: Leader opens a town portal
**Purpose**: Signals companions that leader is returning to town

### Event Processing

When a companion receives a `RequestCompanionJoinGameEvent`:

1. **Check if enabled**: Is companion mode enabled?
2. **Check if not leader**: Am I a follower, not a leader?
3. **Check leader name**: Is this event from my configured leader? (If `leaderName` is empty, accept any leader)
4. **Store game info**: Save the game name and password
5. **Wait for menu**: On next menu cycle, attempt to join the game

When a companion receives a `ResetCompanionGameInfoEvent`:

1. **Check leader name**: Is this from my leader?
2. **Clear stored info**: Erase stored game name and password
3. **Return to waiting**: Wait for next game creation event

---

## Game Creation and Joining Flow

### Leader's Game Creation Process

Located in the game manager (`internal/game/manager.go`):

1. **Increment counter**: Keep track of how many games have been created (starts at 1)
2. **Build game name**: Combine template + counter (e.g., "game-" + "1" = "game-1")
3. **Get password**: Use configured password from companion settings
4. **Create game**: Actually create the game in Diablo 2
5. **Broadcast event**: Send `RequestCompanionJoinGameEvent` to all companions
6. **Run normally**: Execute configured runs (Mephisto, Baal, etc.)
7. **Finish game**: Exit game when runs are complete
8. **Broadcast reset**: Send `ResetCompanionGameInfoEvent` to clear companion info
9. **Repeat**: Go back to step 1 with incremented counter

### Companion's Game Joining Process

Located in the supervisor (`internal/bot/single_supervisor.go`):

1. **Wait for event**: Stay on character selection or lobby screen
2. **Receive event**: Get `RequestCompanionJoinGameEvent` from leader
3. **Store game info**: Save `CompanionGameName` and `CompanionGamePassword`
4. **Enter lobby** (if on character selection screen):
   - Ensure online connection
   - Navigate to lobby
5. **Join game**: Call `Manager.JoinOnlineGame(gameName, gamePassword)`
6. **Handle errors**: If join fails (game full, doesn't exist, etc.), retry or wait
7. **Play game**: Once in game, follow leader and execute companion behavior
8. **Game ends**: Exit to menu when game finishes
9. **Clear info**: Receive `ResetCompanionGameInfoEvent` and clear stored info
10. **Repeat**: Wait for next event

### Menu Flow Handler

The `HandleCompanionMenuFlow()` function is called repeatedly while the companion is on the menu:

**Checks**:
- Is there a stored game name? (If not, do nothing)
- Are we on the character selection screen? (If yes, enter lobby first)
- Are we in the lobby? (If yes, join the game)

This ensures companions will continuously attempt to join the leader's game until successful.

---

## Companion Behavior

### Combat Synchronization

When `attack: true` is configured:

- Companion receives `CompanionLeaderAttackEvent` events from leader
- Companion targets the same enemy unit ID as the leader
- Provides coordinated damage and faster kill times
- Useful for builds that complement each other (tank + damage dealer, etc.)

**Use Cases**:
- Boss killing (all characters focus fire on boss)
- Clearing specific dangerous enemies first
- Coordinating crowd control

### Movement Following

When `followLeader: true` is configured:

- Companion attempts to stay near the leader's position
- Follows leader's pathing through areas
- Helps keep the party together
- Prevents companions from wandering off

**Use Cases**:
- Keeping fragile characters close to tanks
- Ensuring aura-based builds benefit the whole party
- Preventing companions from getting stuck or lost

### Independent Behavior

Even with `attack` and `followLeader` disabled:

- Companions still execute their own AI
- Companions can engage enemies independently
- Companions follow their own pathing logic
- Useful for farming runs where characters split up

### Back to Town Triggers

Companions will return to town independently when:

- **No healing potions**: `backToTown.noHpPotions = true`
- **No mana potions**: `backToTown.noMpPotions = true`
- **Mercenary dies**: `backToTown.mercDied = true`
- **Equipment broken**: `backToTown.equipmentBroken = true`

These are configured separately in the `backToTown` section of the config.

---

## Mercenary vs Companion System

It's important to understand the difference between these two systems:

### Mercenary System

**What it is**:
- Hiring and managing the in-game NPC followers (Act 2 mercs, Act 1 rogue, etc.)

**Features**:
- Hire mercenary at the start of a run
- Monitor mercenary health
- Give mercenary healing/rejuvenation potions
- Choose specific mercenary types (Frozen Aura, etc.)
- Revive mercenary if it dies
- Emergency exit game if mercenary health too low

**Configuration Example**:
```yaml
character:
  useMerc: true
  mercHealingPotionAt: 50
  mercRejuvPotionAt: 25
  mercChickenAt: 10
  shouldHireAct2MercFrozenAura: true
```

**Location**:
- Each character has their own mercenary
- Mercenary follows that specific character
- Independent of companion system

### Companion System

**What it is**:
- Coordinating multiple player characters (bots) to play together

**Features**:
- Multiple characters in the same game
- Leader creates games, companions join
- Optional combat and movement synchronization
- Event-based communication

**Configuration Example**:
```yaml
companion:
  enabled: true
  leader: false
  leaderName: 'Sorceress'
  attack: true
  followLeader: true
```

**Location**:
- Multiple separate characters/bots
- Each character is a full player character
- Coordinated through companion system

### Using Both Together

You can (and often should) use both systems together:

- **Leader character**: Has companion system enabled as leader + uses mercenary
- **Companion character**: Has companion system enabled as follower + uses mercenary

This gives you:
- 1 leader player character
- X companion player characters
- X+1 mercenaries (one per player character)

For a 2-character party, that's 4 total units (2 players + 2 mercs).

---

## Web Interface

### Dashboard

The Koolo web interface provides companion management features:

**Location**: Usually accessible at `http://localhost:8087` (or configured port)

**Features**:
- View companion status for each character
- See if a character is a leader or follower
- Manually trigger companion to join a game
- Monitor companion statistics

### Character Settings Page

The character settings page includes companion configuration:

**Fields**:
- Enable companion mode (checkbox)
- Set as leader (checkbox)
- Leader name (text input)
- Attack same target (checkbox)
- Follow leader (checkbox)
- Game name template (text input)
- Game password (text input)

**Manual Join**:
- "Join Companion Game" button for manual override
- Allows testing or manual coordination

### API Endpoints

The HTTP API provides programmatic access:

- **POST /api/companion-join**: Manually request companion to join leader's game

**Form Fields**:
- `companionEnabled`: Enable/disable companion mode
- `companionLeader`: Set as leader or follower
- `companionLeaderName`: Name of leader character
- `companionGameNameTemplate`: Game name prefix
- `companionGamePassword`: Password for games

---

## Technical Implementation

### Code Organization

**Core Files**:
- `internal/bot/companion.go`: Event handler for companion events
- `internal/bot/single_supervisor.go`: Supervisor logic for game joining
- `internal/event/events.go`: Event definitions
- `internal/config/config.go`: Configuration structure
- `internal/game/manager.go`: Game creation with companion events
- `internal/server/http_server.go`: Web API endpoints

### Configuration Structure

The companion configuration is stored in a struct:

```go
type Companion struct {
    Enabled               bool
    Leader                bool
    LeaderName            string
    GameNameTemplate      string
    GamePassword          string
    CompanionGameName     string  // Runtime only (not saved to file)
    CompanionGamePassword string  // Runtime only (not saved to file)
}
```

**Note**: `CompanionGameName` and `CompanionGamePassword` are populated at runtime when events are received. They are not saved to the configuration file.

### Event Handler

The `CompanionEventHandler` processes incoming events:

**Initialization**:
- Created with references to supervisor, logger, and config
- Registered to listen for companion events

**Event Handling**:
- Receives `RequestCompanionJoinGameEvent`:
  - Checks if companion is enabled and not a leader
  - Verifies leader name matches (or accepts any if empty)
  - Stores game name and password in runtime config
- Receives `ResetCompanionGameInfoEvent`:
  - Checks if event is from configured leader
  - Clears stored game name and password

### Game Name Generation

Game names are generated using a template and counter:

```
gameName = gameNameTemplate + counter
```

**Examples**:
- Template: "game-", Counter: 1 → "game-1"
- Template: "baalrun-", Counter: 5 → "baalrun-5"
- Template: "mf-", Counter: 42 → "mf-42"

The counter increments after each game, ensuring unique game names.

### Supervisor Integration

The single character supervisor has special logic for companions:

**Main Loop**:
1. Check if character is on menu
2. If companion mode enabled and not leader: Call `HandleCompanionMenuFlow()`
3. If leader or normal mode: Call `HandleSinglePlayerFlow()`

**HandleCompanionMenuFlow()**:
- Only executes if `CompanionGameName` is set (from event)
- Handles navigation from character select → lobby → game
- Calls `Manager.JoinOnlineGame(gameName, password)`

### Statistics Tracking

Companion status is tracked in statistics:

- `IsCompanionFollower`: Boolean flag indicating if character is a companion
- Displayed in web dashboard
- Used for filtering and display logic

---

## Troubleshooting

### Common Issues

#### Companion Not Joining Games

**Possible Causes**:
1. **Leader name mismatch**: Companion's `leaderName` doesn't match leader's character name
   - Solution: Check spelling and capitalization of leader name

2. **Companion not enabled**: `enabled: false` in configuration
   - Solution: Set `enabled: true`

3. **Both set as leaders**: Both characters have `leader: true`
   - Solution: Set companion to `leader: false`

4. **Game full**: Diablo 2 has an 8-player limit
   - Solution: Reduce number of companions

5. **Event system issue**: Events not being sent/received
   - Solution: Check logs for event errors, restart bots

#### Companion Joins Old Games

**Possible Causes**:
1. **Reset event not received**: `ResetCompanionGameInfoEvent` wasn't processed
   - Solution: Check if leader is properly sending reset events

2. **Timing issue**: Companion joining before leader fully exits
   - Solution: Add delay in companion join logic

#### Companion Not Attacking Leader's Target

**Possible Causes**:
1. **Attack disabled**: `attack: false` in configuration
   - Solution: Set `attack: true`

2. **Event not being sent**: Leader not sending `CompanionLeaderAttackEvent`
   - Solution: Check leader's attack logic and event sending

3. **Target already dead**: Target dies before companion can attack
   - Solution: Normal behavior, companion will attack next target

#### Companion Not Following Leader

**Possible Causes**:
1. **Follow disabled**: `followLeader: false` in configuration
   - Solution: Set `followLeader: true`

2. **Pathing issues**: Companion stuck or blocked
   - Solution: Check logs for pathing errors, adjust companion AI

3. **Distance too large**: Companion too far behind leader
   - Solution: Adjust movement speed or wait logic

### Debug Tips

1. **Check logs**: Look for companion-related log messages
   - Event sending/receiving
   - Game join attempts
   - Configuration loading

2. **Monitor web dashboard**: Watch companion status in real-time
   - Shows if companion is waiting, joining, or in-game
   - Displays current game name (if stored)

3. **Test with manual join**: Use web interface to manually trigger join
   - Helps isolate if issue is with events or join logic

4. **Verify configuration**: Double-check YAML syntax and values
   - Use YAML validator
   - Check indentation (YAML is whitespace-sensitive)

5. **Test with one companion**: Start with just leader + 1 companion
   - Add more companions after basic setup works

6. **Check game name availability**: Ensure games are actually being created
   - Look at in-game lobby
   - Verify no conflicts with existing games

### Performance Considerations

1. **Multiple companions**: More companions = more resource usage
   - CPU: Each bot runs AI independently
   - Memory: Each game client uses RAM
   - Network: More game traffic

2. **Event processing**: Events are processed in real-time
   - Minimal overhead for small number of bots
   - Can add up with many companions

3. **Game creation delay**: Leader should wait before companions join
   - Prevents join failures
   - Configurable in timing settings

---

## Advanced Usage

### Multiple Leader Groups

You can run multiple independent leader-companion groups:

**Group 1**:
- Leader: "Sorceress" (gameNameTemplate: "mf-")
- Companion: "Paladin" (leaderName: "Sorceress")

**Group 2**:
- Leader: "Barbarian" (gameNameTemplate: "baal-")
- Companion: "Druid" (leaderName: "Barbarian")

Each group operates independently with different game names.

### Empty Leader Name

Setting `leaderName: ''` (empty string) makes a companion follow ANY leader:

```yaml
companion:
  enabled: true
  leader: false
  leaderName: ''  # Follow any leader
```

**Use Cases**:
- Single companion that follows whoever creates games first
- Testing multiple leader configurations
- Flexible party composition

**Caution**: Can cause confusion if multiple leaders are active simultaneously.

### Dynamic Leader Switching

Since leader info is stored at runtime, you can theoretically switch leaders mid-session:

1. Stop current leader
2. Start new leader with same gameNameTemplate
3. Companions will automatically follow new leader

**Note**: This is not officially supported and may cause issues.

### Staggered Companion Joins

To prevent all companions from joining simultaneously:

- Use different timing settings for each companion
- Add delays in companion join logic
- Useful for preventing server load spikes

---

## Future Enhancements

Based on code analysis, some features may be planned but not yet implemented:

### Companion-Specific Runs

In `internal/run/run.go`, there's commented-out code for companion-specific runs:

```go
//if cfg.Companion.Enabled && !cfg.Companion.Leader {
//	return []Run{Companion{baseRun: baseRun}}
//}
```

This suggests there may be plans for companions to have different run behaviors than the leader.

### Enhanced Synchronization

Potential future features:
- Buff coordination (leader casts Battleorders, companions wait)
- Loot distribution (companions don't pick up leader's loot)
- Portal management (companions use leader's portals)
- Quest progression synchronization

---

## Summary

The companion system is a powerful feature for running multiple bot characters together:

**Key Points**:
- **Leader-Follower Model**: One character creates games, others join
- **Event-Driven**: Communication through internal event system
- **Automatic Coordination**: Companions join leader's games without manual intervention
- **Optional Synchronization**: Combat and movement can be coordinated
- **Separate from Mercenaries**: Mercenaries are in-game NPCs, companions are player characters
- **Configurable**: Flexible settings for different party compositions

**Best Practices**:
- Test with one companion before adding more
- Use clear, unique game name templates
- Monitor logs and dashboard for issues
- Set appropriate back-to-town triggers
- Consider resource usage with multiple bots

**Common Use Cases**:
- Magic finding with multiple characters
- Boss runs with coordinated damage
- Leveling parties
- Support character following damage dealer
- Tank + DPS combinations

The companion system enables sophisticated multi-character automation while keeping configuration simple and coordination automatic.
