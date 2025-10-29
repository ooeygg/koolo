# Detailed Explanation of sorceress_leveling.go

This document provides a comprehensive explanation of everything that happens in the `internal/character/sorceress_leveling.go` file, written in plain English.

## Overview

This file implements the complete combat AI and character progression system for a Sorceress character during the leveling phase in Diablo 2. The Sorceress transitions from fire-based skills (Fire Bolt and Fire Ball) in early levels to cold-based skills (Blizzard and Glacial Spike) at level 24, with specialized boss-killing strategies using Static Field.

---

## Constants and Configuration Values

The file begins by defining several important constants that control combat behavior:

### Distance Constants

- **SorceressLevelingMaxAttacksLoop (50)**: The maximum number of attack attempts allowed on a regular monster before giving up and moving on.
- **SorceressLevelingMaxBossAttacksLoop (300)**: The maximum number of attack attempts allowed on boss monsters (much higher since bosses are tougher).
- **SorceressLevelingMinDistance (10)**: The minimum distance the Sorceress should maintain from enemies while attacking (for safety).
- **SorceressLevelingMaxDistance (45)**: The maximum distance at which the Sorceress can attack enemies effectively.
- **SorceressLevelingMeleeDistance (2)**: The distance used when forced to use melee attacks (very close range).
- **SorceressLevelingSafeDistanceLevel (24)**: A level threshold (not actually used in the current code).
- **SorceressLevelingThreatDistance (15)**: The distance used to detect threats.
- **AndarielRepositionLength (9)**: Used for Andariel fight positioning (not directly used).
- **SorceressLevelingDangerDistance (4)**: The distance at which enemies are considered dangerously close.
- **SorceressLevelingSafeDistance (6)**: The distance considered safe from enemies.
- **StaticFieldEffectiveRange (4)**: The maximum distance at which Static Field can reliably hit enemies.

---

## Data Structure: SorceressLeveling

This is the main struct that holds all the state for the Sorceress character:

### Base Components

- **BaseCharacter**: Embedded struct that provides common character functionality (like logger, data access, pathfinding, etc.)

### Combat Tracking Maps

- **blizzardCasts**: A map tracking how many times Blizzard has been cast on each SuperUnique monster (identified by their Unit ID).
- **blizzardPhaseCompleted**: A map tracking whether the initial Blizzard casting phase is completed for each monster.
- **staticPhaseCompleted**: A map tracking whether the Static Field phase is completed for each boss monster.
- **blizzardInitialCastCompleted**: A map tracking whether the initial Blizzard cast has been completed for boss fights.

### Andariel-Specific State

- **andarielMoves**: A counter tracking how many times the character has repositioned during the Andariel fight.
- **andarielSafePositions**: A list of predetermined safe positions to move to during the Andariel fight.

---

## Skill Allocation Sequences

The file defines two different skill point allocation sequences depending on the character's level.

### Fire Skill Sequence (Levels 1-23)

This sequence is used from level 1 until level 24. It focuses on fire skills:

**Level-by-Level Breakdown:**

1. **Level 2, Point 1**: Frozen Armor (defensive buff)
2. **Level 2, Point 2**: Fire Bolt (first Den of Evil skill point)
3. **Levels 3-11**: Continue investing in Fire Bolt (reaches 9 points total)
4. **Level 12**: Switch to Fire Ball (requires level 12)
5. **Level 13**: Fire Ball (second point)
6. **Level 13 (Rada reward)**: Fire Bolt (can't put into Fire Ball due to level restrictions)
7. **Levels 14-16**: Continue Fire Ball (reaches 4 points)
8. **Level 17**: Telekinesis (prerequisite for Teleport)
9. **Level 18**: Teleport (critical mobility skill)
10. **Levels 19-29**: Continue maxing Fire Ball (reaches 17 points)
11. **Level 30**: Fire Mastery (increases fire damage)
12. **Levels 31-35**: Continue Fire Ball to max (reaches 20 points)

**Total Points at Level 35:**
- Fire Bolt: 10 points
- Fire Ball: 20 points
- Fire Mastery: 1 point
- Frozen Armor: 1 point
- Telekinesis: 1 point
- Teleport: 1 point

### Blizzard Skill Sequence (Level 24+)

This sequence is used after the character respecs at level 24. It focuses on cold skills:

**Phase 1 - Utility and Prerequisites (9 points):**
1. Static Field (utility for bosses)
2. Telekinesis (utility, prerequisite for Teleport)
3. Teleport (mobility)
4. Frozen Armor (defense)
5. Warmth (mana regeneration)
6. Frost Nova (prerequisite for Blizzard)
7. Ice Bolt (prerequisite and synergy for Blizzard)
8. Ice Blast (prerequisite and synergy for Blizzard)
9. Glacial Spike (prerequisite and synergy for Blizzard)

**Phase 2 - First Blizzard Point (1 point):**
10. Blizzard (main attack skill unlocked)

**Phase 3 - Synergy Building (17 points total so far):**
11-17. Glacial Spike (7 more points, total 8)
18-26. Ice Blast (9 more points, total 10)

**Phase 4 - Maxing Blizzard (36 points total):**
27-35. Blizzard (9 more points, total 10)
36. Cold Mastery (first point in mastery)

**Phase 5 - More Blizzard (46 points total):**
37-46. Blizzard (10 more points, total 20 - MAXED)

**Phase 6 - Maxing Cold Mastery (62 points total):**
47-62. Cold Mastery (16 more points, total 17)

**Phase 7 - Maxing Glacial Spike Synergy:**
63-67. Glacial Spike (5 points)
68-71. Lightning prerequisites (Charged Bolt, Lightning, Chain Lightning, Energy Shield)
72-81. Glacial Spike (10 more points, total 20 - MAXED)

**Phase 8 - Maxing Ice Blast Synergy:**
82-92. Ice Blast (11 more points, total 20 - MAXED)

**Phase 9 - Maxing Ice Bolt Synergy:**
93-111. Ice Bolt (19 more points, total 20 - MAXED)

**Phase 10 - Finishing Cold Mastery:**
112-114. Cold Mastery (3 more points, total 20 - MAXED)

**Final Build at High Level:**
- Blizzard: 20 points (main attack)
- Cold Mastery: 20 points (enemy resistance reduction)
- Glacial Spike: 20 points (secondary attack and synergy)
- Ice Blast: 20 points (synergy)
- Ice Bolt: 20 points (synergy)
- Various utility skills: 1 point each

---

## Helper Functions

### ShouldIgnoreMonster()

**Purpose**: Determines if a monster should be ignored during combat.

**Logic**: Currently returns false for all monsters, meaning the Sorceress doesn't skip any enemies.

### isPlayerDead()

**Purpose**: Checks if the player character is dead.

**Logic**: Checks if the player's health percentage is zero or less. Returns true if dead, false if alive.

### CheckKeyBindings()

**Purpose**: Verifies that all required skills are properly bound to keyboard keys.

**Logic**:
1. Creates an empty list of required keybindings (currently the list is empty, so no bindings are actually required).
2. Creates an empty list to track missing bindings.
3. Loops through each required skill and checks if it has a keybinding.
4. If a required skill is missing a binding, adds it to the missing list.
5. If there are missing bindings, logs a debug message listing them.
6. Returns the list of missing bindings.

**Current Behavior**: Since the required list is empty, this function always returns an empty list. It's a placeholder for potential future requirements.

### findDangerousMonsters()

**Purpose**: Identifies all monsters that are dangerously close to the player.

**Logic**:
1. Creates an empty list to store dangerous monsters.
2. Loops through all monsters in the current game.
3. For each monster, checks two conditions:
   - The monster has health greater than zero (is alive)
   - The distance from the player to the monster is less than the danger distance threshold (4 units)
4. If both conditions are met, adds the monster to the dangerous list.
5. Returns the complete list of dangerous monsters.

**Use Case**: This helps the AI decide when to reposition to avoid being overwhelmed by nearby enemies.

---

## Main Combat Function: KillMonsterSequence()

This is the heart of the Sorceress combat AI. It's a complex, intelligent combat loop that handles attacking a specific monster with adaptive strategies. Here's the complete breakdown:

### Function Parameters

1. **monsterSelector**: A function that returns the Unit ID of the monster to attack and a boolean indicating if it was found.
2. **skipOnImmunities**: A list of resistances (fire, cold, lightning, etc.) that should cause the function to skip the monster.

### Initialization and Setup

**Attack Loop Counter**:
- Initializes a counter to track how many attack loops have been completed (starts at 0).
- Tracks the previous monster's Unit ID to detect when switching targets.
- Records the last time the character repositioned (for cooldown purposes).

**Andariel Safe Positions**:
- Defines three predetermined safe positions for the Andariel fight:
  - Position 1: (22547, 9591)
  - Position 2: (22547, 9600)
  - Position 3: (22547, 9609)

**Static Field Target List**:
- Creates a map of boss monsters that should be attacked with Static Field:
  - Andariel
  - Duriel
  - Izual
  - Diablo
  - Baal (BaalCrab form)
  - Ancient Barbarian (all three versions)

### Main Combat Loop

The function enters an infinite loop that continues until the monster dies or max attacks are reached.

#### Loop Iteration Start

1. **Priority Check**: Checks if the bot should pause (for user intervention or higher priority tasks).

2. **Increment Counter**: Increases the completed attack loops counter by 1.

3. **Logging**: Logs how many attack loops have been completed.

4. **Death Check**: Checks if the player is dead. If yes, logs a message and exits the function immediately.

5. **Monster Selection**: Calls the monster selector function to get the current target.
   - If no monster is found, exits the function (monster is dead or gone).

6. **Target Change Detection**: Checks if the current target is different from the previous target:
   - If different (new target), resets all tracking:
     - Attack loop counter back to 0
     - Blizzard cast tracking map cleared
     - Blizzard phase completion map cleared
     - Static Field phase completion map cleared
     - Initial Blizzard cast map cleared
     - Andariel move counter reset to 0

#### Pre-Battle Checks

7. **Pre-Battle Validation**: Calls a pre-battle check function that validates the target:
   - Checks immunities against the skip list
   - If checks fail, exits the function

8. **Monster Retrieval**: Finds the actual monster object by its Unit ID:
   - If not found, logs "Target monster not found or died"
   - Waits 500 milliseconds
   - Exits the function

#### Special Case: Andariel Repositioning (Normal Difficulty Only)

9. **Andariel Distance Check**: If fighting Andariel on Normal difficulty:
   - Calculates the distance to Andariel
   - If distance is less than minimum safe distance (10 units) AND there are still safe positions to move to:
     - Gets the next safe position from the predefined list
     - Increments the move counter
     - Logs the repositioning action
     - Moves to that safe position
     - Waits 200 milliseconds
     - Continues to the next loop iteration (re-evaluates the situation)

#### Monster Analysis and Target Prioritization

10. **Immunity Detection**: Determines the monster's immunity status:
    - **isColdImmuneNotLightImmune**: True if the monster is immune to cold but not immune to lightning
    - **isBossTarget**: True if the monster is in the Static Field target list
    - **isFireImmune**: True if the monster is immune to fire

11. **Proximity Threat Detection**: If the current target is NOT a boss and NOT cold immune:
    - Loops through all other monsters in the area
    - Skips the current target monster
    - Calculates the distance from the player to both the current target and each other monster
    - Sets a proximity threshold to minimum distance plus 5 (15 units total)
    - If another monster is:
      - Closer to the player than the current target
      - Within the proximity threshold
    - Logs that a new, closer monster was detected
    - Re-prioritizes to the closer threat
    - Continues to the next loop iteration

12. **Max Attack Loop Check**: Determines the maximum attacks allowed:
    - If the target is a boss: 300 attacks maximum
    - If the target is a regular monster: 50 attacks maximum
    - If the completed attack loops exceed this maximum:
      - Logs that max attacks were reached
      - Exits the function (gives up on this monster)

#### Skill Level and Attack Options Setup

13. **Get Player Level**: Retrieves the player's current level.

14. **Attack Distance Configuration**: Sets up attack options:
    - Default attack option: maintain distance between 10 and 45 units
    - Glacial Spike attack option: same distance (10-45 units)

#### Repositioning for Safety (Non-Cold-Immune Monsters)

15. **Danger Detection**: For non-cold-immune monsters, checks for dangerous enemies nearby:
    - Calls a function to check if any enemy is within danger distance (4 units)
    - Also checks if enough time has passed since the last reposition (1 second cooldown)
    - If both conditions are met:
      - Updates the last reposition timestamp
      - **Death Check**: Checks if player died during this time
      - **Target Revalidation**: Re-selects the target monster (might have changed)
      - **Target Retrieval**: Gets the current target monster object
      - Logs the dangerous monster detection with its distance
      - **Safe Position Calculation**: Finds a safe position that is:
        - At least 4 units away from the dangerous monster
        - At least 6 units away in a safe direction
        - Between 10 and 45 units from the target monster
      - If a safe position is found:
        - Logs the safe position coordinates
        - Checks if Teleport is available (skill level > 0)
        - Checks if Teleport is keybind and not on cooldown
        - **Death Check**: Another death check before teleporting
        - Moves to the safe position
        - Waits 200 milliseconds
        - Continues to the next loop iteration
      - If Teleport is not available or on cooldown:
        - Logs that repositioning is not possible
      - If no safe position is found:
        - Logs that a safe position could not be calculated

### Boss Combat Strategy

16. **Boss Target Confirmation**: Re-checks if the current target is a boss.

17. **Skill Availability Checks**:
    - Checks if Static Field is learned (skill level > 0)
    - Checks if Static Field has a keybinding
    - Checks if Blizzard has a keybinding

18. **Boss Initial Blizzard Cast**: For boss targets:
    - Checks if the initial Blizzard has been cast yet
    - If NOT yet cast:
      - Checks if Blizzard is bound and the player is not on cooldown
      - Checks if the player is not currently casting
      - If ready:
        - Logs "Boss: Casting initial Blizzard"
        - Casts Blizzard on the boss
        - Marks initial Blizzard as completed
        - Waits 100 milliseconds
        - Continues to next loop iteration
      - If player is busy casting:
        - Logs "Boss: Player busy, waiting for initial Blizzard cast"
        - Waits 50 milliseconds
        - Continues to next loop iteration

19. **Boss Static Field Phase**: After the initial Blizzard:
    - Checks if the Static Field phase is already completed
    - If completed:
      - Logs "Boss: Static Field phase already completed"
      - Proceeds to main attack logic
    - If NOT completed:
      - Calculates the monster's current health percentage
      - Determines the required health threshold based on difficulty:
        - **Normal/Nightmare**: 40% health
        - **Hell**: 70% health (Static Field is less effective in Hell)
      - If the monster's health is ABOVE the threshold:
        - Checks if Static Field is available and bound
        - If yes:
          - Calculates distance to the monster
          - If distance is greater than Static Field range (4 units):
            - Logs "Boss: Too far for Static Field, repositioning closer"
            - Moves toward the monster
            - Waits 100 milliseconds
            - Continues to next loop iteration
          - If within range and player is not casting:
            - Logs "Boss: Using Static Field on target"
            - Casts Static Field (range 0-4 units)
            - Waits 100 milliseconds
            - Continues to next loop iteration (re-check health)
          - If player is busy casting:
            - Logs "Boss: Player busy, skipping Static Field for this tick"
            - Waits 50 milliseconds
            - Continues to next loop iteration
        - If Static Field is not available:
          - Marks Static Field phase as completed
          - Logs "Boss: Static Field skill not available. Transitioning to main attack"
          - Falls through to main attack logic
      - If the monster's health is AT OR BELOW the threshold:
        - Marks Static Field phase as completed
        - Logs "Boss: Static Field threshold reached. Transitioning to main attack"
        - Falls through to main attack logic

### Cold Immune Monster Strategy

20. **Cold Immune (Non-Boss) Combat**: For cold immune monsters that aren't bosses:
    - Checks if the mercenary is alive
    - If mercenary is dead:
      - Logs "Mercenary is dead, skipping attack on Cold Immune monster"
      - Exits the function (can't kill cold immune without help)
    - Initializes the tracking maps if they're null
    - Determines how many Blizzards to cast based on monster type:
      - **SuperUnique**: 2 Blizzard casts required
      - **Regular**: 1 Blizzard cast required
    - Gets the current Blizzard cast count for this monster
    - If not enough Blizzards cast yet and phase not completed:
      - Checks if Blizzard is available, bound, and not on cooldown
      - If ready:
        - Logs "CI: Casting initial Blizzard (cast X/Y)"
        - Casts Blizzard on the monster
        - Increments the cast counter
        - Waits 100 milliseconds
        - Continues to next loop iteration
    - If enough Blizzards cast:
      - Marks Blizzard phase as completed
    - **Telestomp with Static Field**: After Blizzards are cast:
      - Checks if Static Field is available and bound
      - If yes:
        - Calculates distance to monster
        - If distance is greater than Static Field range (4 units):
          - Logs "CI: Too far for Static Field, repositioning closer"
          - Moves toward the monster
          - Waits 100 milliseconds
          - Continues to next loop iteration
        - If within range and player is not casting:
          - Logs "CI: Using Static Field until dead"
          - Casts Static Field
          - Waits 100 milliseconds
          - Continues to next loop iteration
        - If player is busy:
          - Logs "CI: Player busy, skipping Static Field"
          - Waits 50 milliseconds
      - If Static Field not available:
        - Logs "CI: Static Field skill not available. Transitioning to main attack"
        - Falls through to main attack logic

### Main Attack Logic

This section handles the standard attack patterns for all monsters:

21. **Fire Immune Early Level Handling**: If the player is below level 12 AND the monster is fire immune:
    - Logs "Under level 12 and facing a Fire Immune monster, using primary attack"
    - Uses the primary attack (physical attack, melee range 1-2 units)
    - Records the previous Unit ID
    - Waits 50 milliseconds
    - Continues to next loop iteration

22. **Low Mana Handling**: If mana is below 15% AND player is below level 12:
    - **Death Check**: Checks if player died
    - Logs "Low mana, using primary attack (left-click skill, e.g., Attack/Fire Bolt)"
    - Uses primary attack (melee range 1-2 units)
    - Records the previous Unit ID
    - Continues to next loop iteration

23. **Blizzard Attack (Level 24+)**: If the player has Blizzard learned:
    - Checks if Blizzard has a keybinding
    - If yes:
      - **Cooldown Handling**: If Blizzard is on cooldown:
        - Checks if Glacial Spike is learned
        - If yes:
          - **Death Check**: Checks if player died
          - If player is not currently casting:
            - Logs "Blizzard on cooldown, attempting to cast Glacial Spike (Main)"
            - Uses Glacial Spike as primary attack (2 attacks, uses shift-click for stationary casting, distance 10-45 units)
          - If player is busy:
            - Logs "Player is busy, waiting to cast Glacial Spike (Main)"
            - Waits 50 milliseconds
      - **Blizzard Casting**: If Blizzard is NOT on cooldown:
        - **Death Check**: Checks if player died
        - If player is not currently casting:
          - Logs "Using Blizzard (Main)"
          - Casts Blizzard as secondary attack (distance 10-45 units)
        - If player is busy:
          - Logs "Player is busy, waiting to cast Blizzard (Main)"
          - Waits 50 milliseconds

24. **Fire Skill Fallback (Below Level 24)**: If Blizzard is not available:
    - Starts with the default attack skill (physical attack)
    - **Meteor Check**: If Meteor is learned and has a keybinding:
      - Sets current attack skill to Meteor
    - **Fire Ball Check**: Else if Fire Ball is learned and has a keybinding:
      - Sets current attack skill to Fire Ball
    - **Fire Bolt Check**: Else if Fire Bolt has a keybinding:
      - Sets current attack skill to Fire Bolt
    - If a fire skill was found:
      - **Death Check**: Checks if player died
      - If player is not currently casting:
        - Logs "Using [skill name] (fallback)"
        - Casts the selected fire skill as secondary attack (distance 10-45 units)
      - If player is busy:
        - Logs "Player is busy, skipping [skill name] (fallback) for this tick"
        - Waits 50 milliseconds
    - If NO skills are available:
      - **Death Check**: Checks if player died
      - Logs "No secondary skills available, using primary attack (fallback)"
      - Uses physical primary attack (melee range 1-2 units)

25. **Loop Completion**: After the attack:
    - Records the current monster's Unit ID as the previous Unit ID
    - Waits 50 milliseconds
    - Continues to the next loop iteration

---

## Monster-Specific Kill Functions

### killMonsterByName()

**Purpose**: Provides a wrapper around KillMonsterSequence for killing a specific named monster.

**Parameters**:
1. **id**: The NPC ID of the monster to kill
2. **monsterType**: The type of monster (Unique, SuperUnique, etc.)
3. **skipOnImmunities**: List of immunities to skip

**Logic**:
1. Enters an infinite loop.
2. Searches for the monster by ID and type.
3. If monster is found:
   - Checks if its health is zero or less
   - If dead, logs a message and breaks the loop
   - If alive, calls KillMonsterSequence with a monster selector function that:
     - Searches for the monster by ID and type
     - Returns its Unit ID if found
   - If KillMonsterSequence returns an error:
     - Logs the error
     - Breaks the loop
   - Waits 20 milliseconds to prevent busy-looping
4. If monster is not found:
   - Logs "Monster not found. Assuming it's dead or vanished"
   - Breaks the loop
5. Returns nil (no error)

**Use Case**: Simplifies killing named bosses and specific monsters.

---

## Buff and Skill Configuration Functions

### BuffSkills()

**Purpose**: Returns a list of buff skills the Sorceress should cast before combat.

**Logic**:
1. Creates an empty skill list.
2. **Energy Shield Check**: If Energy Shield has a keybinding:
   - Adds Energy Shield to the list
3. **Armor Selection**: Checks for armor skills in order of preference:
   - Chilling Armor (best)
   - Shiver Armor (second best)
   - Frozen Armor (basic)
   - If any armor skill has a keybinding:
     - Adds that armor to the list
     - Returns the list immediately (only use one armor skill)
4. Returns the complete buff list.

**Result**: Ensures the Sorceress always has defensive buffs active.

### PreCTABuffSkills()

**Purpose**: Returns a list of skills to cast before using Call to Arms (CTA) weapon swap.

**Logic**: Returns an empty list (no pre-CTA buffs needed for leveling Sorceress).

---

## Character Progression Functions

### ShouldResetSkills()

**Purpose**: Determines if the character should use Akara's free respec to switch from fire to cold skills.

**Logic**:
1. Gets the player's current level.
2. Checks three conditions:
   - Player is exactly level 24
   - Fire Ball skill level is greater than 7
   - Fire Bolt skill level is greater than 7
3. If all conditions are met:
   - Logs "Respecing to Blizzard: Level 32+ and FireBall level > 9" (note: the log message has a typo—it says level 32+ but checks for 24)
   - Returns true
4. Otherwise, returns false.

**Use Case**: Automatically triggers the respec at level 24 to switch from the fire build to the cold build.

### SkillsToBind()

**Purpose**: Determines which skills should be bound to keyboard keys based on the character's current level and skills.

**Returns**: Two values:
1. The main/primary attack skill
2. A list of all skills that should be keybind

**Logic**:

1. Gets the player's current level.

2. **Starting Bindings**: Creates a base list with Fire Bolt.

3. **Level-Based Additions**:
   - **Level 2+**: Adds Frozen Armor
   - **Level 6+**: Adds Static Field
   - **Level 18+**: Adds Teleport
   - **Level 24+**: Adds Blizzard

4. **Skill Availability Checks** (if below level 24):
   - If Meteor is learned: Adds Meteor
   - Else if Hydra is learned: Adds Hydra
   - Else if Fire Ball is learned: Adds Fire Ball

5. **Utility Skill Checks**:
   - If Energy Shield is learned: Adds Energy Shield
   - If Battle Command is learned: Adds Battle Command (from Call to Arms runeword)
   - If Battle Orders is learned: Adds Battle Orders (from Call to Arms runeword)

6. **Main Skill Selection**:
   - If level 24 or higher: Main skill is Glacial Spike
   - Otherwise: Main skill is basic Attack

7. **Town Portal Tome Check**:
   - Searches inventory for a Tome of Town Portal
   - If found: Adds it to the skill bindings

8. Logs the main skill and all bindings.

9. Returns the main skill and the complete binding list.

**Use Case**: Ensures the Sorceress always has the appropriate skills accessible via keyboard shortcuts.

### StatPoints()

**Purpose**: Defines the priority order for allocating stat points as the character levels up.

**Returns**: A list of stat allocation targets in order.

**The Allocation Sequence**:

The system alternates between Vitality (health) and Strength (gear requirements), with emphasis on survivability:

1. **Vitality: 20** then **Strength: 20** (early survivability)
2. **Vitality: 30** then **Strength: 30**
3. **Vitality: 40** then **Strength: 40**
4. **Vitality: 50** then **Strength: 50**
5. **Vitality: 100** then **Strength: 95** (heavy focus on health)
6. **Vitality: 250** then **Strength: 156** (enough strength for most gear)
7. **Vitality: 999** (dump all remaining points into Vitality for maximum survivability)

**Strategy**: Prioritizes Vitality for survivability while ensuring enough Strength to equip important gear like Spirit shields (156 strength requirement).

### SkillPoints()

**Purpose**: Returns the appropriate skill point allocation sequence based on the character's current level.

**Logic**:
1. Gets the player's current level.
2. If level is less than 24:
   - Returns the fire skill sequence
3. If level is 24 or higher:
   - Returns the blizzard skill sequence
4. The system automatically allocates points following the returned sequence.

**Use Case**: Ensures skill points are spent optimally for the character's current progression stage.

---

## Boss-Specific Kill Functions

Each boss has a specialized function that handles the unique mechanics and strategies for that encounter.

### KillCountess()

**Purpose**: Kills the Countess (a SuperUnique monster in the Forgotten Tower).

**Logic**: Calls killMonsterByName with:
- NPC ID: DarkStalker (Countess's internal name)
- Monster Type: SuperUnique
- Skip Immunities: None (nil)

**Strategy**: Uses the standard combat logic with 2 initial Blizzard casts (since she's SuperUnique) followed by Static Field spam if she's cold immune.

### KillAndariel()

**Purpose**: Kills Andariel, the Act 1 boss.

**Special Logic**:
1. If playing on Hell difficulty:
   - Saves the current "back to town" configuration
   - Enables the "return to town if mercenary dies" setting
   - Sets up a deferred function to restore the original settings after the fight
   - Logs "Restored original back-to-town checks after Duriel fight" (note: this is a typo in the log message—should say Andariel)
2. Calls killMonsterByName with:
   - NPC ID: Andariel
   - Monster Type: Unique
   - Skip Immunities: None

**Strategy**:
- On Normal: Uses the special repositioning logic in KillMonsterSequence that moves through three safe positions
- On Hell: Protects the mercenary by returning to town if the merc dies
- Uses boss combat strategy: initial Blizzard → Static Field to 40%/70% → main Blizzard rotation

### KillSummoner()

**Purpose**: Kills the Summoner in the Arcane Sanctuary (Act 2).

**Special Logic**:
1. Saves the current "back to town" configuration.
2. Disables all "return to town" triggers:
   - No HP potions: false (don't return)
   - No MP potions: false (don't return)
   - Equipment broken: false (don't return)
   - Mercenary died: false (don't return)
3. Sets up a deferred function to restore the original settings after the fight.
4. Logs "Restored original back-to-town checks after Mephisto fight" (note: typo—should say Summoner).
5. Calls killMonsterByName with:
   - NPC ID: Summoner
   - Monster Type: Unique
   - Skip Immunities: None

**Strategy**: Prevents premature town returns during this fight, ensuring the Summoner is fully killed before leaving.

### KillDuriel()

**Purpose**: Kills Duriel, the Act 2 boss.

**Logic**: Calls killMonsterByName with:
- NPC ID: Duriel
- Monster Type: Unique
- Skip Immunities: None

**Strategy**: Uses boss combat strategy with Static Field. Duriel is a very tanky boss, so the extended attack loop (300 attempts) is important.

### KillCouncil()

**Purpose**: Kills the Council Members in Travincal (Act 3).

**Special Logic**: This uses KillMonsterSequence directly with a custom monster selector:

1. **Monster Selector Function**:
   - Creates an empty list of council members
   - Loops through all monsters in the game
   - Adds any monster with these names to the list:
     - CouncilMember
     - CouncilMember2
     - CouncilMember3
   - Sorts the list by distance from the player (closest first)
   - Returns the Unit ID of the closest council member
   - If no council members found, returns false

2. **Skip Immunities**: None

**Strategy**: Kills council members one at a time, always prioritizing the closest one. This prevents being overwhelmed by multiple council members at once.

### KillMephisto()

**Purpose**: Kills Mephisto, the Act 3 boss. This is the most complex boss fight implementation.

The function has two completely different strategies based on configuration:

#### Strategy 1: Static Field Pre-Damage (If UseStaticOnMephisto is enabled)

This aggressive strategy gets Mephisto to low health before using the moat trick:

1. **Setup**:
   - Configures attack options (distance 10-45 units)
   - Configures Static Field range (0-4 units)
   - Moves to position (17565, 8065) near Mephisto

2. **Initial Monster Check**:
   - Searches for Mephisto
   - If not found, logs an error and exits

3. **Initial Blizzard**:
   - If Blizzard is available:
     - Logs "Applying initial Blizzard cast"
     - Casts Blizzard on Mephisto
     - Waits 300 milliseconds for the chill effect to apply

4. **Aggressive Static Field Phase**:
   - Checks if Static Field is available and bound
   - If yes:
     - Logs "Starting aggressive Static Field phase on Mephisto"
     - Determines the target health percentage based on difficulty:
       - **Normal/Nightmare**: 40%
       - **Hell**: 70%
     - Sets a maximum of 50 Static Field casts
     - Enters a loop (up to 50 iterations):
       - **Monster Check**: Finds Mephisto
       - If Mephisto is dead or gone, breaks the loop
       - **Health Check**: Calculates Mephisto's current health percentage
       - If health is at or below the threshold:
         - Logs that the threshold was reached
         - Breaks the loop
       - **Distance Check**: Calculates distance to Mephisto
       - If distance is greater than Static Field range (4 units) AND Teleport is available:
         - Logs "Mephisto too far for Static Field, repositioning closer"
         - Teleports toward Mephisto
         - Waits 150 milliseconds
         - Continues to next iteration
       - **Static Field Cast**: If within range and not currently casting:
         - Logs "Using Static Field on Mephisto"
         - Casts Static Field
         - Waits 150 milliseconds
       - If player is busy casting:
         - Waits 50 milliseconds
       - Increments the attack counter
   - If Static Field not available:
     - Logs "Static Field not available or bound, skipping Static Phase"

5. **Repositioning**:
   - Moves to position (17563, 8072)
   - If movement fails, returns the error

#### Strategy 2: Moat Trick (If UseMoatTrick is enabled)

This is the classic Mephisto cheese strategy where you attack him from across the moat where he can't reach you:

1. **Standard Kill**: If the moat trick is NOT enabled:
   - Simply calls killMonsterByName with standard parameters

2. **Moat Trick Implementation**: If enabled:
   - **Setup**:
     - Gets the game context
     - Sets attack distance to 15-80 units (very long range)
     - Enables force attack (attacks even if you can't see the target clearly)
     - Sets up a deferred function to disable force attack when done
   - **Initial Positioning**:
     - Waits 350 milliseconds
     - Moves to position (17563, 8072)
     - Waits 350 milliseconds
   - **Movement Sequence**: Executes a precise sequence of movements:
     - Move to (17575, 8086), wait 350ms
     - Move to (17584, 8088), wait 1200ms
     - Move to (17600, 8090), wait 550ms
     - Move to (17609, 8090), wait 2500ms
     - These positions navigate around the moat to the safe spot
   - **Area Clearing**:
     - Clears all enemies within 10 units of position (17609, 8090)
     - Ensures no monsters interrupt the moat trick
   - **Final Positioning**:
     - Moves to (17609, 8090) - the moat trick position
   - **Attack Loop**:
     - Sets maximum attacks to 100
     - Initializes attack counter to 0
     - Enters loop (up to 100 attacks):
       - **Priority Check**: Checks if bot should pause
       - **Monster Check**: Searches for Mephisto
       - If not found, exits the function (he's dead)
       - **Cooldown Handling**: If Blizzard is on cooldown:
         - Uses Glacial Spike as primary attack (2 casts, shift-click)
         - Waits 50 milliseconds
       - **Blizzard Cast**: If Blizzard is available:
         - Casts Blizzard on Mephisto
         - Waits 100 milliseconds
       - Increments attack counter
   - Returns nil (success)

**Strategy Summary**: The aggressive Static Field approach is faster but riskier. The moat trick is slower but completely safe as Mephisto cannot attack you from across the moat.

### KillIzual()

**Purpose**: Kills Izual in the Plains of Despair (Act 4).

**Logic**: Calls killMonsterByName with:
- NPC ID: Izual
- Monster Type: Unique
- Skip Immunities: None

**Strategy**: Uses boss combat strategy with Static Field. Izual is a challenging boss with high health and resistances.

### KillDiablo()

**Purpose**: Kills Diablo, the Act 4 final boss.

**Special Logic**:

1. **Timeout Setup**:
   - Sets a 20-second timeout for finding Diablo
   - Records the start time
   - Initializes a "Diablo found" flag as false

2. **Diablo Detection Loop**:
   - Enters an infinite loop
   - **Timeout Check**: If 20 seconds have passed AND Diablo was never found:
     - Logs "Diablo was not found, timeout reached"
     - Returns nil (exits)
   - **Diablo Search**: Searches for Diablo (NPC ID: Diablo, Type: Unique)
   - If not found OR Diablo's health is zero or less:
     - If Diablo was previously found (meaning he died):
       - Returns nil (success)
     - If Diablo was never found (still searching):
       - Waits 200 milliseconds
       - Continues to next loop iteration
   - **Diablo Found**: If Diablo is alive:
     - Sets the "Diablo found" flag to true
     - Logs "Diablo detected, attacking"
     - **Configuration Backup**: Saves the current "back to town" settings
     - **Configuration Override**: Disables all town return triggers:
       - No HP potions: false
       - No MP potions: false
       - Equipment broken: false
       - Mercenary died: false
     - **Deferred Restoration**: Sets up cleanup to restore settings and log when done
     - **Combat**: Calls killMonsterByName with:
       - NPC ID: Diablo
       - Monster Type: Unique
       - Skip Immunities: None
     - Returns the result

**Strategy**: Waits patiently for Diablo to spawn (he has an animation sequence), then commits fully to the fight without returning to town prematurely.

### KillPindle()

**Purpose**: Kills Pindleskin (Defiled Warrior) in the Halls of Anguish (Act 5).

**Logic**: Calls killMonsterByName with:
- NPC ID: DefiledWarrior
- Monster Type: SuperUnique
- Skip Immunities: Uses the immunities configured in the Pindleskin game settings

**Strategy**: Respects the skip immunity configuration, allowing the bot to avoid Pindleskin if he has immunities the Sorceress can't handle.

### KillNihlathak()

**Purpose**: Kills Nihlathak in the Halls of Vaught (Act 5).

**Logic**: Calls killMonsterByName with:
- NPC ID: Nihlathak
- Monster Type: SuperUnique
- Skip Immunities: None

**Strategy**: Uses SuperUnique combat logic (2 initial Blizzards) followed by standard rotation. Nihlathak's Corpse Explosion is dangerous, so the high health from stat allocation is important.

### KillAncients()

**Purpose**: Kills all three Ancient Barbarians at the Arreat Summit (Act 5).

**Special Logic**:

1. **Configuration Backup**:
   - Saves the current "back to town" configuration

2. **Configuration Override**:
   - Disables all town return triggers:
     - No HP potions: false
     - No MP potions: false
     - Equipment broken: false
     - Mercenary died: false

3. **Ancient Killing Loop**:
   - Gets all elite monsters in the area (the three Ancients)
   - For each elite monster:
     - Searches for it by name as a SuperUnique
     - If not found, skips to the next
     - Moves to position (10062, 12639) - a safe fighting position
     - Calls killMonsterByName on that Ancient
     - Continues to the next Ancient

4. **Configuration Restoration**:
   - Restores the original "back to town" configuration
   - Logs "Restored original back-to-town checks after Ancients fight"

5. Returns nil (success)

**Strategy**: Fights each Ancient one at a time from a safe position, with disabled town returns to ensure all three are killed before leaving.

### KillBaal()

**Purpose**: Kills Baal, the Act 5 final boss.

**Logic**: Calls killMonsterByName with:
- NPC ID: BaalCrab (Baal's second form after he transforms)
- Monster Type: Unique
- Skip Immunities: None

**Strategy**: Uses boss combat strategy with Static Field. The extended 300 attack loop is crucial for this very tanky boss.

---

## Runeword Configuration

### GetAdditionalRunewords()

**Purpose**: Returns a list of additional runewords that are specifically useful for a caster character like the Sorceress.

**Logic**:
1. Calls a helper function that returns common caster runewords (likely includes things like Spirit, Insight, Lore, etc.)
2. Returns that list

**Use Case**: Combined with the base runeword list, this ensures the Sorceress can create all the runewords beneficial for a caster build.

---

## Summary

The `sorceress_leveling.go` file implements a sophisticated, adaptive AI for leveling a Sorceress character in Diablo 2. Key features include:

### Combat Intelligence
- **Adaptive Strategy**: Automatically switches between fire and cold builds at level 24
- **Boss Mechanics**: Specialized strategies for each boss with Static Field pre-damage
- **Immunity Handling**: Uses Blizzard + Static Field for cold immune monsters (with mercenary support)
- **Safety First**: Constant repositioning to maintain safe distance from enemies
- **Priority Targeting**: Automatically switches to closer threats when detected

### Character Progression
- **Automatic Respec**: Triggers at level 24 to switch from fire to cold build
- **Optimized Stat Allocation**: Balances Vitality and Strength for survivability and gear requirements
- **Dynamic Skill Binding**: Adjusts keybindings as new skills become available
- **Buff Management**: Automatically maintains Energy Shield and armor buffs

### Boss Strategies
- **Andariel**: Repositioning through safe positions on Normal
- **Mephisto**: Optional moat trick or aggressive Static Field approach
- **Diablo**: Patient detection with no premature town returns
- **Baal & Others**: Extended attack loops with Static Field optimization
- **Council**: Distance-based prioritization to avoid being overwhelmed

The system is designed to handle the entire leveling process from level 1 to endgame, with minimal user intervention required.
