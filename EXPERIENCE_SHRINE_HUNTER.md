# Experience Shrine Hunter System

## Overview

The Experience Shrine Hunter is an automated feature that searches for and activates Experience Shrines at the start of every game for Lightning Sorceress and Nova Sorceress characters. This provides an XP boost for efficient leveling and run optimization.

## How It Works

### Activation
The shrine hunt runs automatically at the start of **every game** for:
- **Lightning Sorceress** (`lightsorc`)
- **Nova Sorceress** (`nova`)

### Search Strategy
The bot uses a fast, optimized search pattern:

1. **Waypoint-Based Search**: Only searches near waypoints (50-unit radius)
2. **Limited Checks**: Checks maximum 3 shrines per waypoint before moving on
3. **Smart Sorting**: Always checks closest shrines first
4. **Quick Completion**: Designed to complete in under 30 seconds

### Search Order
The bot searches Act 1 waypoints in this order:
1. **Cold Plains** waypoint
2. **Stony Field** waypoint
3. **Dark Wood** waypoint
4. **Black Marsh** waypoint

### Behavior at Each Waypoint
1. Teleport to waypoint
2. Scan all shrines within 50 units
3. Sort shrines by distance (closest first)
4. Check up to 3 shrines:
   - If Experience Shrine found → activate it and return to town
   - If not Experience Shrine → check next shrine
   - After 3 shrines → move to next waypoint
5. Continue until Experience Shrine found or all waypoints checked
6. Return to town and proceed with normal PreRun routine

## Implementation Details

### Files Modified
- **`internal/action/town.go`**: Added shrine hunt call in `PreRun()` function
- **`internal/action/shrine_hunt.go`**: New file containing shrine hunting logic

### Key Functions

#### `FindExperienceShrineAct1()`
Main entry point for the shrine hunt. Orchestrates the waypoint search loop.

```go
func FindExperienceShrineAct1() error
```

#### `quickSearchAroundWaypoint()`
Performs fast shrine scanning within 50 units of current waypoint.

```go
func quickSearchAroundWaypoint(searchArea area.ID) (bool, error)
```

#### `interactWithExperienceShrine()`
Handles shrine interaction and activation.

```go
func interactWithExperienceShrine(shrine *data.Object) error
```

### Configuration Constants

```go
const searchRadius = 50          // Search within 50 units of waypoint
const maxShrinesToCheck = 3      // Check max 3 shrines per waypoint
```

## Companion System Integration

### Companion Behavior
When using the Companion system, both Leader and Companion characters will hunt for Experience Shrines:

#### Leader Character
- Hunts for Experience Shrine at game start
- Creates game after shrine activation
- Waits for companions to join

#### Companion Character
- Hunts for Experience Shrine at game start (independently)
- Joins leader's game after shrine hunt completes
- Both leader and companion get XP boost

### Companion Timing
The shrine hunt happens **before** companion synchronization:
1. Leader creates game
2. Leader hunts for shrine
3. Companion joins game
4. Companion hunts for shrine (in their own instance)
5. Both characters proceed with runs

This ensures both characters maximize XP gains without interfering with each other's shrine hunts.

## Performance Optimization

### Why It's Fast
- **No Area Clearing**: Doesn't traverse entire areas or clear rooms
- **Memory Reading**: Instantly reads all shrines from game memory
- **Distance Filtering**: Only considers shrines within 50 units
- **Early Termination**: Stops after finding Experience Shrine
- **Limited Checks**: Maximum 3 shrines per waypoint prevents time waste

### Expected Performance
- **Typical duration**: 15-30 seconds
- **Best case**: 5-10 seconds (shrine at first waypoint)
- **Worst case**: 30-40 seconds (no shrine found, all waypoints checked)

## Debug Information

### Logging
The shrine hunt provides detailed logging:

```
[INFO] Nova detected - starting Experience Shrine hunt
[DEBUG] Checking for Experience Shrine near Cold Plains waypoint
[DEBUG] Quick searching 50 units around waypoint in Cold Plains
[DEBUG] Found 4 shrines within 50 units of waypoint
[DEBUG] Found shrine type ExperienceShrine at distance 23
[INFO] Experience Shrine found at position {X:12345, Y:23456} (distance: 23)!
[INFO] Experience Shrine found and activated near Cold Plains waypoint!
```

### Debug Conditions
PreRun logs the conditions being checked:

```
[DEBUG] PreRun conditions - firstRun: false, isLevelingChar: false, class: lightsorc
```

## Troubleshooting

### Shrine Hunt Not Running

**Check these conditions:**
1. Character class must be `lightsorc` or `nova`
2. Character must NOT be a leveling character
3. Check debug logs for "PreRun conditions" message

**Enable debug logging** in `config/koolo.yaml`:
```yaml
debug:
  log: true
```

### Shrine Not Found

**Common reasons:**
- RNG: Not all games spawn Experience Shrines near waypoints
- Search radius: Shrines beyond 50 units won't be detected
- Shrine limit: Bot stops after checking 3 shrines per waypoint

**Adjustments:**
To increase search radius or shrine check limit, modify constants in `shrine_hunt.go`:
```go
const searchRadius = 75          // Increase from 50 to 75
const maxShrinesToCheck = 5      // Increase from 3 to 5
```

### Performance Issues

If shrine hunt takes too long:
1. **Reduce waypoints**: Remove Dark Wood/Black Marsh from search order
2. **Reduce shrine checks**: Lower `maxShrinesToCheck` to 2
3. **Reduce search radius**: Lower `searchRadius` to 40

Edit `shrine_hunt.go` line 24:
```go
areasToSearch := []area.ID{area.ColdPlains, area.StonyField}  // Only 2 waypoints
```

## Future Enhancements

### Potential Improvements
- **Configuration Toggle**: Add config option to enable/disable shrine hunt
- **Class Expansion**: Add support for other character classes
- **Smart Waypoint Order**: Learn which waypoints have better shrine spawn rates
- **Shrine Caching**: Remember shrine locations across multiple games
- **Multi-Shrine Support**: Option to search for other beneficial shrines

### Advanced Configuration (Future)
```yaml
character:
  shrine_hunt:
    enabled: true
    classes: ["lightsorc", "nova", "blizz"]
    waypoints: ["ColdPlains", "StonyField"]
    search_radius: 50
    max_shrines_per_waypoint: 3
    timeout_seconds: 30
```

## Code Example

### Adding Shrine Hunt to New Character Class

To enable shrine hunting for another character class, edit `internal/action/town.go`:

```go
className := strings.ToLower(ctx.CharacterCfg.Character.Class)
if !isLevelingChar && (className == "lightsorc" || className == "nova" || className == "blizz") {
    ctx.Logger.Info(fmt.Sprintf("%s detected - starting Experience Shrine hunt", ctx.CharacterCfg.Character.Class))
    if err := FindExperienceShrineAct1(); err != nil {
        ctx.Logger.Warn(fmt.Sprintf("Experience Shrine hunt encountered an error: %v, continuing with normal PreRun routine", err))
    }
}
```

### Custom Shrine Search Function

To search for different shrine types:

```go
// Search for Skill Shrine instead
if sd.shrine.Shrine.ShrineType == object.SkillShrine {
    ctx.Logger.Info("Skill Shrine found!")
    // Activate shrine...
}
```

## Summary

The Experience Shrine Hunter provides automatic XP boost acquisition for Lightning and Nova Sorceress characters with:
- ✅ Fast 15-30 second search time
- ✅ Runs at the start of every game
- ✅ Works with Companion system
- ✅ No impact on normal bot operations
- ✅ Detailed debug logging
- ✅ Easy to configure and extend

This feature ensures your Sorceress characters start every run with maximum XP efficiency!
