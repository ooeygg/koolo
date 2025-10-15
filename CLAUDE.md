# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Koolo is a bot for Diablo II: Resurrected that reads game memory and interacts with the game by injecting clicks/keystrokes. The project is written in Go and targets Windows only.

**Important**: This is a game automation tool that can result in bans. The codebase involves memory reading/injection and game automation, which should only be used for educational purposes in offline mode.

## Build System

### Requirements
- **Go 1.24** (not 1.25 - this is critical)
- Garble obfuscation tool: `go install mvdan.cc/garble@v0.14.2`
- Windows-only (uses Windows-specific APIs via `github.com/lxn/win`)

### Build Commands
- **Primary build**: `better_build.bat` - This is the recommended build script
  - Creates obfuscated executable with Garble
  - Handles tool folder copying
  - Manages configuration files (Settings.json, koolo.yaml)
  - Outputs to `build/` directory with unique GUID-based exe name
- **Legacy build**: `build.bat` - Older script, use `better_build.bat` instead

### Garble Configuration
The `GOGARBLE` environment variable excludes specific packages from obfuscation:
- `internal/server*` - Web server needs to remain unobfuscated
- `internal/event*` - Event system
- `github.com/inkeliz/gowebview*` - WebView dependency

## Architecture

### Core Components

#### Bot System (`internal/bot/`)
- **Supervisor**: Main orchestrator that manages bot lifecycle
  - `baseSupervisor` handles character selection, online connectivity, window positioning
  - Manages game state transitions and pause/resume
  - Located in `supervisor.go`
- **StatsHandler**: Tracks run statistics (games completed, items found, etc.)
- **Companion**: Multi-bot coordination system (currently not fully working per README)
  - Leaders create games, companions join them
  - Events: `RequestCompanionJoinGameEvent`, `ResetCompanionGameInfoEvent`

#### Character System (`internal/character/`)
Characters implement the `context.Character` interface with methods like:
- `KillCountess()`, `KillAndariel()`, `KillMephisto()`, etc. - Boss-specific kill logic
- `KillMonsterSequence()` - Generic monster killing with immunity checks
- `BuffSkills()`, `PreCTABuffSkills()` - Character buffing
- `CheckKeyBindings()` - Validates required skills are bound

**Character Classes**:
- Sorceress variants: `BlizzardSorceress`, `NovaSorceress`, `LightningSorceress`, `HydraOrbSorceress`, `FireballSorceress`
- Paladin: `Hammerdin`, `Foh`
- Assassin: `Trapsin`, `MosaicSin`
- Barbarian: `Berserker` (Berserk barb with hork/Find Item)
- Druid: `WindDruid`
- Amazon: `Javazon`

**Leveling Characters**: Special implementations with `LevelingCharacter` interface:
- `SorceressLeveling`, `PaladinLeveling`, `AssassinLeveling`
- Add `StatPoints()`, `SkillPoints()`, `SkillsToBind()` methods for auto-leveling

#### Run System (`internal/run/`)
Each run is a separate type implementing the `Run` interface:
```go
type Run interface {
    Name() string
    Run() error
}
```

Runs are built via `BuildRuns()` which maps config run names to run instances:
- Boss runs: `Countess`, `Andariel`, `Summoner`, `Duriel`, `Mephisto`, `Diablo`, `Baal`, etc.
- Area runs: `Pit`, `AncientTunnels`, `Travincal`, `Cows`, `TerrorZone`, etc.
- Leveling: `Leveling` (uses act-specific files: `leveling_act1.go` through `leveling_act5.go`)
- Quests: `Quests` (automated quest completion)

#### Action System (`internal/action/`)
Reusable actions that characters and runs use:
- **Movement**: `move.go` - Pathfinding and movement
- **Combat**: `clear_area.go`, `clear_level.go` - Area clearing logic
- **Item Management**: `item_pickup.go`, `stash.go`, `vendor.go`, `identify.go`
- **Inventory**: `belt_manage.go`, `refill_belt_from_inventory.go`
- **Equipment**: `autoequip.go` - Auto-equip system with tier-based scoring
- **Crafting**: `cube_recipes.go`, `runeword_maker.go`, `gambling.go`
- **Town Actions**: `town.go`, `tp_actions.go`, `repair.go`, `revive_merc.go`
- **Leveling Tools**: `leveling_tools.go` - Quest item handling, stat/skill allocation
- **Step Package** (`step/`): Atomic game actions like `interact_object.go`, `attack.go`, `set_skill.go`

#### Game Interface (`internal/game/`)
- **Memory Reading**: `memory_reader.go` - Reads game state from D2R memory
- **Memory Injection**: `memory_injector.go` - Injects code into game process
- **HID**: `keyboard.go`, `mouse.go`, `hid.go` - Input injection to game window
- **Data**: `data.go` - Game data structures wrapper around `d2go` library
- **Map Client**: `map_client/` - Communicates with external koolo-map.exe tool for pathfinding data
- **Grid System**: `grid.go` - Inventory/stash grid management
- **Utilities**: `screenshot.go`, `crash_detector.go`, `handle_killer.go`

#### Pathfinding (`internal/pather/`)
- Uses A* algorithm (`astar/` package)
- `path_finder.go` - Main pathfinding logic using map data from map client
- Works with external `koolo-map.exe` tool (in `tools/` directory)

#### Configuration (`internal/config/`)
- `config.go` - Main configuration structures
  - `KooloCfg` - Global Koolo settings (D2 paths, Discord/Telegram, logging)
  - `CharacterCfg` - Per-character configuration (class, runs, health settings, inventory, game settings)
- `runs.go` - Run type constants (`CountessRun`, `AndarielRun`, etc.)
- Configuration files stored in `config/` directory:
  - `koolo.yaml` - Global settings
  - `config/<character_name>/config.yaml` - Per-character YAML
  - `config/<character_name>/pickit/*.nip` - Item pickit rules (NIP format)
  - `config/<character_name>/pickit_leveling/*.nip` - Leveling-specific pickit

#### Context (`internal/context/`)
- `character.go` - Defines `Character` and `LevelingCharacter` interfaces
- Context aggregates: GameReader, HID, Logger, CharacterCfg, Data, MemoryInjector

#### Event System (`internal/event/`)
- Event-driven architecture for bot notifications
- `events.go` - Event type definitions
- `listener.go` - Central event dispatcher
- Events: GameCreated, GameFinished, ItemDropped, RunFinished, GamePaused, etc.

#### Remote Integration (`internal/remote/`)
- **Discord** (`discord/`): Discord bot integration for monitoring/control
- **Telegram** (`telegram/`): Telegram bot integration
- **Droplog** (`droplog/`): Item drop logging to files

#### Server (`internal/server/`)
- Local web server (port 8087) for the WebView-based UI
- REST API for bot control (start/stop/pause, stats retrieval)

#### Town Management (`internal/town/`)
- Per-act town navigation: `A1.go` through `A5.go`
- NPC locations, waypoint positions, stash/vendor interactions

#### Health Management (`internal/health/`)
- `health_manager.go` - Auto-potion usage, chicken (exit game on low HP)
- `belt_manager.go` - Belt potion management

### Main Entry Point (`cmd/koolo/`)
- `main.go` - Application entry point
  - Starts WebView UI (1280x720 window)
  - Initializes supervisor manager, event listener, scheduler
  - Starts Discord/Telegram bots if enabled
  - Sets up web server on port 8087
  - Uses errgroup for goroutine management with panic recovery

## Key Dependencies

- `github.com/hectorgimenez/d2go` (replaced with fork: `github.com/kwader2k/d2go`) - D2R game data structures and memory reading
- `github.com/inkeliz/gowebview` - WebView for UI
- `github.com/lxn/win` - Windows API bindings
- `github.com/bwmarrin/discordgo` - Discord integration
- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram integration
- `gopkg.in/yaml.v3` - YAML configuration parsing

## Development Patterns

### Adding a New Character Class
1. Create file in `internal/character/` (e.g., `new_class.go`)
2. Implement `context.Character` interface
3. Embed `BaseCharacter` for common functionality
4. Add boss kill methods (`KillAndariel()`, etc.)
5. Implement `KillMonsterSequence()` with immunity handling using `preBattleChecks()`
6. Register in `BuildCharacter()` switch statement
7. For leveling: implement `context.LevelingCharacter` interface with stat/skill allocation

### Adding a New Run
1. Create file in `internal/run/` (e.g., `my_run.go`)
2. Implement `Run` interface with `Name()` and `Run()` methods
3. Use actions from `internal/action/` for movement, combat, looting
4. Add run constant to `internal/config/runs.go`
5. Register in `BuildRuns()` switch statement in `run.go`

### Adding Configuration Options
1. Add fields to `CharacterCfg` in `internal/config/config.go`
2. Use YAML tags for serialization
3. Update `config/template/config.yaml` with defaults
4. Access via `ctx.CharacterCfg` in character/run implementations

### Character-Specific Configuration
Many classes have nested config sections in `CharacterCfg.Character`:
- `BerserkerBarb.FindItemSwitch` - Enable weapon swap for Find Item
- `NovaSorceress.BossStaticThreshold` - HP% to start using Static Field
- `BlizzardSorceress.UseMoatTrick` - Use moat trick on Mephisto
- `MosaicSin.UseTigerStrike`, `UseCobraStrike`, etc. - Which finishers to use

## Testing

No automated test suite exists. The only test file is `internal/pather/astar/astar_test.go` for A* pathfinding.

## Common Tasks

### Running the Bot
1. Build with `better_build.bat`
2. Execute `build/<guid>.exe`
3. Configure via web UI at http://localhost:8087
4. Ensure D2 LOD 1.13c is installed (required for memory offsets)

### Adding Discord/Telegram Integration
- Configure in `config/koolo.yaml`
- Event handlers in `internal/remote/discord/` and `internal/remote/telegram/`
- Events are sent via `event.Send()` throughout the codebase

### Debugging
- Enable `debug.log: true` in `config/koolo.yaml`
- Logs written to `LogSaveDirectory` (default: `logs/`)
- Use `ctx.Logger.Debug()`, `ctx.Logger.Info()`, etc. throughout code
- Screenshots can be enabled with `debug.screenshots: true`

## Code Organization Notes

- **Windows-specific**: All Windows API calls go through `internal/utils/winproc/`
- **Game interaction**: HID input via `internal/game/hid.go`, never direct Windows API in other files
- **Pathfinding data**: External tool `tools/koolo-map.exe` provides map/collision data via localhost websocket
- **NIP files**: Pickit rules use NIP format (parsed by `d2go/pkg/nip`)
- **Difficulty settings**: Use `github.com/hectorgimenez/d2go/pkg/data/difficulty` constants
- **Area IDs**: Use `github.com/hectorgimenez/d2go/pkg/data/area` constants
- **Skill IDs**: Use `github.com/hectorgimenez/d2go/pkg/data/skill` constants

## Important Constraints

1. **Always use Go 1.24** - Version 1.25 has compatibility issues
2. **Windows only** - Uses Windows-specific memory reading and input injection
3. **D2 LOD 1.13c required** - Memory offsets depend on this version
4. **Game must be 1280x720 windowed** - Fixed resolution requirement
5. **English game language** - Prevents language-related bugs
6. **Companion mode currently broken** - Multi-bot coordination is WIP per README
