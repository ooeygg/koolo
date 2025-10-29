# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Koolo is a bot for Diablo II: Resurrected that reads game memory and injects input (clicks/keystrokes) to automate gameplay. The bot supports multiple character builds (Sorceress, Paladin, Assassin, Druid, Necromancer, Barbarian, etc.) and various run types (boss runs, leveling, farming). The project is written in Go and uses obfuscation (Garble) for distribution.

## Build System

### Building the Project
```bash
# Standard build (preferred method)
better_build.bat

# This script:
# - Validates Go 1.24 and Garble 0.14.2 are installed
# - Builds obfuscated executable using Garble
# - Copies tools, config templates, and assets to build/ directory
# - Preserves existing Settings.json if present
```

### Requirements
- **Go 1.24** (NOT 1.25 - version 1.24 is explicitly required)
- **Garble 0.14.2** (NOT 0.15.x - older version required for compatibility)
- Install Garble: `go install mvdan.cc/garble@v0.14.2`

### Key Build Notes
- The build produces a unique executable name (GUID-based) in the `build/` directory
- Obfuscation preserves UI and critical packages (see `GOGARBLE` in better_build.bat)
- Anti-virus software may interfere with compilation - add project to exclusions if needed

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/pather/astar

# Run tests with verbose output
go test -v ./...
```

Note: There are minimal tests currently (mainly in internal/pather/astar). Most functionality is tested through runtime execution.

## Architecture

### Core Components

1. **Supervisor System** (`internal/bot/supervisor.go`)
   - Manages bot lifecycle (start, stop, pause)
   - Each character instance runs under a Supervisor
   - Coordinates game window management and priority handling

2. **Character System** (`internal/character/`)
   - Each character build implements the `Character` interface
   - Character files define combat rotation, skill usage, and build-specific logic
   - Examples: `blizzard_sorceress.go`, `hammerdin.go`, `mosaic.go`
   - Leveling variants exist for supported classes: `sorceress_leveling.go`, `paladin_leveling.go`, etc.

3. **Run System** (`internal/run/`)
   - Each run type (boss, area, leveling) is a separate file
   - Runs define pathing, objectives, and completion conditions
   - Examples: `mephisto.go`, `baal.go`, `travincal.go`, `leveling_act1.go`

4. **Action System** (`internal/action/`)
   - Reusable high-level actions: `clear_area.go`, `item_pickup.go`, `stash.go`
   - Low-level steps in `internal/action/step/`: `cast.go`, `attack.go`, `interact_npc.go`
   - Actions use the context to access game state and character capabilities

5. **Context System** (`internal/context/context.go`)
   - Central state container passed through all bot operations
   - Contains: `Data` (game state), `Char` (character instance), `PathFinder`, `HealthManager`, etc.
   - Priority system controls execution (Normal, Background, Pause, Stop)
   - Use `context.Get()` to access current context

6. **Game Interface** (`internal/game/`)
   - `memory_reader.go`: Reads game memory using d2go library
   - `memory_injector.go`: Injects input into game
   - `hid.go`: Human Interface Device simulation (mouse/keyboard)
   - `manager.go`: Coordinates game state reading and updates

7. **Pathfinding** (`internal/pather/`)
   - A* pathfinding implementation in `astar/`
   - Path calculation, obstacle detection, and movement logic
   - Integrates with map data from external `koolo-map.exe` tool

8. **Pickit System** (`internal/pickit/`)
   - NIP file parser for item filtering rules
   - Item database with stat evaluation
   - Auto-equip logic with item scoring

9. **Config System** (`internal/config/`)
   - YAML-based configuration (`koolo.yaml`)
   - Per-character settings in `config/template/`
   - Game settings, run configuration, and bot behavior

10. **Event System** (`internal/event/`)
    - Event-driven architecture for notifications
    - Integrations: Discord, Telegram, drop logging
    - Events: item drops, deaths, game completion, etc.

### Data Flow

1. **Main Loop**: `cmd/koolo/main.go` starts web UI, supervisor manager, and scheduler
2. **Supervisor** creates a Bot instance with character config and game reader
3. **Bot** reads game state via `MemoryReader` (d2go wrapper)
4. **Character** implementation decides actions based on game state
5. **Actions** execute via HID (keyboard/mouse injection) and memory injection
6. **Event System** broadcasts state changes to listeners (Discord, Telegram, logs)

### External Dependencies

- **d2go library** (`github.com/kwader2k/d2go`): D2R memory structures and game data
  - Provides: `data.Monster`, `data.Item`, `area.ID`, memory reading primitives
  - This is a forked version (see go.mod replace directive)
- **koolo-map.exe** (`tools/koolo-map.exe`): External map server providing collision data
- **handle64.exe** (`tools/handle64.exe`): Windows utility for process handle management

## Development Patterns

### Adding a New Character Build

1. Create file in `internal/character/` (e.g., `new_build.go`)
2. Implement the `Character` interface with combat logic
3. Define skill rotation, buffing, and stat requirements
4. Add character initialization in character factory
5. Create corresponding config template in `config/template/`

### Adding a New Run

1. Create file in `internal/run/` (e.g., `new_area.go`)
2. Implement path to objective, clear logic, and exit conditions
3. Register run in `internal/run/run.go` factory
4. Add run configuration option in config system

### Working with Game State

```go
// Access current context
ctx := context.Get()

// Read game data
player := ctx.Data.PlayerUnit
monsters := ctx.Data.Monsters.Enemies()
items := ctx.Data.Items

// Use character actions
ctx.Char.KillMonsterSequence(...)
ctx.Char.BuffSkills()

// Pathfinding
path, _ := ctx.PathFinder.GetPath(targetPosition)
```

### Memory Injection vs HID

- **Memory Injection**: Used for teleport, direct state manipulation (requires memory_injector)
- **HID (Keyboard/Mouse)**: Used for normal movement, skill casting, clicking
- Always restore memory after injection: `ctx.MemoryInjector.RestoreMemory()`

## Configuration

### Character Configuration Location
- Template configs: `config/template/{CharacterName}/`
- User configs generated in: `build/config/{CharacterName}/`
- Pickit rules (NIP files): `config/{CharacterName}/pickit/*.nip`

### Key Settings
- `Settings.json`: Global bot settings (paths, game executable)
- `koolo.yaml`: Per-supervisor settings (Discord, Telegram, game password)
- Character YAML: Build-specific settings (skills, stats, runs, leveling)

## Leveling System

The leveling system is a work-in-progress feature that automates character progression from level 1:

- Leveling runs in `internal/run/leveling_act*.go` (act1-act5)
- Character-specific leveling configs: `sorceress_leveling.go`, `paladin_leveling.go`, etc.
- Enable via character config: `leveling: true` in Run Settings
- Auto-skill allocation via `SkillPoints()` method in character implementation
- See `docs/` for detailed leveling explanations

## Important Notes

- The project requires Diablo II: LOD 1.13c for game data extraction
- Game must run at 1280x720 windowed mode
- Only English language supported (prevents language-related bugs)
- Memory addresses and structures from d2go may break with game patches
