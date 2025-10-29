# Detailed Explanation of leveling_act1.go

This document provides a comprehensive explanation of everything that happens in the `internal/run/leveling_act1.go` file, written in plain English.

## Overview

This file contains the logic for automatically leveling a character through Act 1 of Diablo 2. It handles quest completion, farming strategies, equipment purchasing, and progression through different difficulty levels (Normal, Nightmare, and Hell).

---

## Main Function: act1()

The `act1()` function is the central orchestrator for all Act 1 leveling activities. Here's what happens step by step:

### Initial Setup and Validation

1. **Location Check**: The function first checks if the player is currently in the Rogue Encampment (the Act 1 town). If the player is not there, the function exits immediately and does nothing.

2. **Quest Log Update**: The game's quest log is updated to get the current state of all quests.

3. **Level 1 Character Setup**: The function checks the player's current level. If the player is exactly level 1, it logs an info message saying "Player level is 1. Setting Leveling Config" and calls a special setup function to configure all the settings for a brand new character.

4. **Post-Level 1 Preparations**: If the player is higher than level 1, the function does two things:
   - Refills potions at a vendor
   - Ensures all skill keybindings are properly set up (if this fails, it logs an error)

### Quest Completion and Farming Strategy

The function now enters the main quest and farming logic, which follows this priority order:

#### Fast Travel to Act 4 (If Applicable)

5. **Act 4 Shortcut for Farming**: If the Sisters to the Slaughter quest (killing Andariel) is already completed, the function attempts to use a waypoint to travel directly to The Pandemonium Fortress in Act 4. This is useful if the player has already completed Act 1 and is farming the Mausoleum area. If the player dies during Mausoleum farming, they would respawn in Act 1, so this logic helps them quickly return to Act 4. If the waypoint works, the function exits. If the waypoint fails, it logs that it will fall back to manual portal entry.

#### Low Gold Farming

6. **Nightmare Gold Farming**: If the player has less than 50,000 gold and is playing on Nightmare difficulty, the function runs the Stony Field farming route to earn gold.

7. **Hell Gold Farming**: If the player has less than 50,000 gold and is playing on Hell difficulty:
   - If they have less than 5,000 gold, the function sets a special configuration where the character clears enemies in a 20-unit radius (called "clear path distance"). This "eco run" setting applies even to sorceresses, who normally wouldn't engage enemies while teleporting.
   - Then it runs the Mausoleum farming route to earn gold and experience.

8. **Hell Leveling**: If the difficulty is Hell and the player is level 75 or lower, the function runs the Mausoleum farming route regardless of gold.

#### Quest Progression

9. **Den of Evil Quest**: If the Den of Evil quest is not completed and the difficulty is not Hell, the function completes the Den of Evil quest. This quest gives a skill point reward.

10. **Early Normal Gold Farming**: In Normal difficulty, if the player has less than 300 gold and hasn't completed the Sisters Burial Grounds quest (Blood Raven), the function runs a special early gold farm in Tristram.

11. **Blood Raven Quest**: In Normal difficulty, if the Sisters Burial Grounds quest (killing Blood Raven) is not completed, the function runs the Blood Raven encounter to complete it.

12. **Mercenary Hiring**: After the Blood Raven quest is completed, if the character configuration shows they're not using a mercenary, the function:
    - Sets the configuration to enable mercenary use
    - Interacts with Kashya (the NPC who gives you the mercenary reward)
    - Saves the updated character configuration
    - Logs an error if saving fails

#### Equipment Purchases

13. **Belt Upgrade**: In Normal difficulty, if the player has more than 3,000 gold and is between level 9 and 11, the function attempts to gamble for a 9-slot belt from Gheed. This provides more potion storage capacity.

#### Cain Rescue Quest

14. **Entering Tristram for Cain**: If the Search for Cain quest has one of several "in progress" statuses (specifically status 2, 3, or 4, which correspond to different stages of the quest) and the difficulty is not Hell, the function runs the Tristram area to rescue Cain.

15. **Completing Cain Quest**: If Cain is not in town, the quest is not marked as completed, and the difficulty is not Hell, the function completes the rescue Cain quest sequence (which involves talking to Akara).

#### Leveling Farm Routes

16. **Early Tristram Farming (Levels 1-5)**: In Normal difficulty, if the player is below level 12:
    - The function sets the clear path distance based on character class (4 units for sorceresses, 20 units for others)
    - Saves the configuration
    - If the player is below level 6, it runs Tristram and exits the function
    - If the player is level 6 or higher (but still below 12), it runs Tristram but continues with other activities

17. **Countess Farming for Runes**: The function farms the Countess in several scenarios:
    - **Normal Difficulty**: If the Cain quest is completed, the player is level 6 or higher, AND either:
      - The player is below level 12, OR
      - The player is below level 16 AND playing as a paladin or necromancer
    - When farming Countess as a sorceress, the clear path distance is set to 15 units
    - **Nightmare Difficulty**: If the player is below level 50, the Den of Evil is completed, AND a special check determines that the character still needs specific runes (the function calls `shouldFarmCountessForRunes()`)

#### Act Progression

18. **Moving to Act 2 or Fighting Andariel**: Finally, the function checks if the Sisters to the Slaughter quest (killing Andariel) is completed:
    - **If completed**: The function calls `goToAct2()` to transition to Act 2
    - **If not completed**: The function prepares for the Andariel fight:
      - In Normal difficulty, it adjusts the clear path distance (7 units for sorceresses, 15 for others)
      - Changes the belt configuration to carry 2 columns of healing potions and 2 columns of mana potions
      - Saves the configuration
      - Runs the Andariel encounter

---

## Supporting Function: setupLevelOneConfig()

This function is called when a brand new level 1 character is detected. It centralizes all the initial configuration setup for optimal leveling. Here's everything it configures:

### Game Settings

- Sets the game difficulty to Normal
- Enables automatic skill point allocation
- Enables automatic key binding for skills
- Enables automatic equipment equipping
- Enables the runeword maker system
- Configures which runewords can be made (retrieved from `GetRunewords()`)
- Enables teleport usage for movement
- Disables mercenary use initially (you don't have one yet)
- Disables stashing items to shared stash
- Enables using Cain to identify items
- Keeps the mini panel open (doesn't auto-close it)
- Sets maximum game length to 1200 seconds (20 minutes)

### Health and Potion Settings

- Uses healing potions when health drops to 40 percent
- Uses mana potions when mana drops to 25 percent
- Doesn't use rejuvenation potions for life (set to 0)
- "Chickens" (exits the game to safety) when health drops to 7 percent
- For mercenary: uses rejuvenation potions at 40 percent health
- For mercenary: doesn't chicken (set to 0)
- For mercenary: uses healing potions at 25 percent health
- Belt configuration: 2 columns of healing potions, 2 columns of mana potions
- Inventory carrying: 4 healing potions, 8 mana potions, 0 rejuvenation potions

### Gambling and Town Behavior

- Enables gambling
- Returns to town when out of health potions
- Returns to town when out of mana potions
- Doesn't return to town if mercenary dies
- Doesn't return to town if equipment breaks

### Cube Recipes

- Enables cube recipes
- Enables specific recipes: "Perfect Amethyst"

### Area-Specific Settings

**Tristram:**
- Doesn't clear the portal area
- Focuses on elite monster packs

**Pit:**
- Moves through Black Marsh
- Opens chests
- Doesn't focus on elite packs only
- Doesn't only clear level 2

**Andariel:**
- Clears the entire room before fighting
- Uses antidote potions

**Mephisto (for later acts):**
- Doesn't kill council members
- Doesn't open chests
- Exits to Act 4 after the kill

### Inventory Management

- Locks the entire inventory (sets a 4x10 grid of locked slots) so items stay organized
- Sets minimum gold pickup threshold to 2,000 gold

### Shrine and Mercenary Settings

- Enables interaction with shrines for buffs
- Configures to hire an Act 2 mercenary with the Frozen Aura (Defensive mercenary from Nightmare)

### Saving Configuration

After all these settings are configured, the function saves the configuration to disk. If saving fails, it logs an error.

---

## Supporting Function: AdjustDifficultyConfig()

This function adjusts configuration settings based on the player's current level and difficulty. It's called to adapt the character's behavior as they progress. Here's what it does:

### Initial Level Check

1. **Skip for Level 1**: If the player is level 1, the function exits immediately because `setupLevelOneConfig()` already handled the initial setup.

### Universal Settings

2. **Teleport for Sorceresses**: If the character class is a sorceress in leveling mode, teleport is enabled.

3. **Runeword Updates**: Updates the list of enabled runewords based on character progression.

4. **Gold Pickup Adjustment**: Sets the minimum gold pickup threshold to 5,000 multiplied by the player's level. For example, a level 10 character won't pick up gold piles smaller than 50,000.

### Level 4-23 Settings

5. **Mid-Early Leveling**: If the player is between level 4 and 23:
   - Sets healing potion usage threshold to 85 percent (more conservative than level 1)
   - Sets clear path distance to 7 units for sorceresses, 15 units for others

### Level 24+ Settings (Varies by Difficulty)

6. **Normal Difficulty (Level 24+)**:
   - Belt configuration: 2 columns healing, 2 columns mana
   - Mercenary uses healing potions at 55 percent health
   - Mercenary doesn't use rejuvenation potions (set to 0)
   - Player uses healing potions at 85 percent
   - Chicken threshold raised to 30 percent (less paranoid)
   - Clear path distance set to 15 units

7. **Nightmare Difficulty (Level 24+)**:
   - Same configuration as Normal difficulty at level 24+

8. **Hell Difficulty (Level 24+)**:
   - Belt configuration: 2 columns healing, 1 column mana, 1 column rejuvenation (more defensive)
   - Mercenary uses healing potions at 80 percent health (more conservative)
   - Mercenary uses rejuvenation potions at 40 percent health
   - Player uses healing potions at 90 percent (very conservative)
   - Chicken threshold raised to 40 percent (more survivability focus)
   - Clear path distance: 0 for sorceresses (don't engage while teleporting if out of mana), 15 for others

### Saving Changes

9. After applying the appropriate difficulty settings, the function saves the configuration. If saving fails, it logs an error.

---

## Supporting Function: GetRunewords()

This function determines which runewords the character should be able to craft. Here's the logic:

1. **Base Runewords**: Starts with a list of essential runewords that work for most characters:
   - Ancients' Pledge (shield runeword for resistances)
   - Lore (helmet runeword)
   - Insight (polearm runeword for mercenary, provides meditation aura)
   - Smoke (armor runeword for resistances)
   - Treachery (armor runeword with fade proc)
   - Call to Arms (weapon runeword for battle orders buff)

2. **Ladder-Only Runewords**: If the character is playing on Ladder (not Non-Ladder):
   - Adds "Bulwark" to the list
   - Adds "Hustle" to the list
   - Logs an info message: "Ladder character detected. Adding Bulwark and Hustle runewords."

3. **Character-Specific Runewords**: The function checks if the character implements the LevelingCharacter interface. If it does:
   - Calls the character's `GetAdditionalRunewords()` method to get any class-specific runewords
   - Appends these additional runewords to the list

4. **Return**: Returns the complete list of enabled runewords

---

## Supporting Function: goToAct2()

This function handles the transition from Act 1 to Act 2. Here's the complete sequence:

1. **Logging**: Logs the message "Act 1 completed. Moving to Act 2."

2. **Return to Town**: Calls a function to ensure the character is in town (Rogue Encampment).

3. **Den of Evil Cleanup**: Before leaving Act 1, checks if the Den of Evil quest is completed:
   - If NOT completed AND difficulty is not Hell, completes the Den of Evil quest
   - If completing the quest returns an error, the function exits with that error

4. **Cain Rescue Cleanup**: Checks if Deckard Cain is in town:
   - If Cain is NOT in town AND difficulty is not Hell, completes the rescue Cain quest
   - If completing the quest returns an error, the function exits with that error

5. **Warriv Interaction**: Interacts with the NPC Warriv (the caravan merchant who travels between acts).

6. **Menu Navigation**: Simulates keyboard input to navigate Warriv's dialogue menu:
   - Presses HOME key (goes to top of menu)
   - Presses DOWN arrow key (selects "travel to Act 2" option)
   - Presses RETURN/ENTER key (confirms selection)

7. **Wait for Dialogue**: Waits 1000 milliseconds (1 second) for dialogue to process.

8. **Skip Cinematics**: Holds down the SPACE bar for 2000 milliseconds (2 seconds) to skip any transition cinematics or dialogue.

9. **Final Wait**: Waits another 1000 milliseconds (1 second) for the transition to complete.

10. **Success**: Returns no error, indicating successful transition to Act 2.

---

## Supporting Function: stonyField()

This function farms the Stony Field area for gold and experience. Here's what it does:

1. **Waypoint Travel**: Uses the waypoint system to travel to the Stony Field area.
   - If this fails (returns an error), the function exits with that error.

2. **Clear the Area**: Calls a function to clear all enemies in the current level:
   - The first parameter `false` indicates not to open containers/chests
   - Uses a monster filter that accepts any monster type (doesn't discriminate)

3. **Return**: Returns any error that occurred during the clearing process, or nil if successful.

---

## Supporting Function: isCainInTown()

This simple function checks whether Deckard Cain has been rescued and is present in town:

1. **Monster Search**: Searches the monster/NPC data for Deckard Cain (specifically NPC ID for Cain in Act 1, version 5).
   - The search looks for monster type "None" (meaning a regular NPC, not a hostile monster)

2. **Return**: Returns true if Cain is found, false if he's not present.

---

## Supporting Function: killRavenGetMerc()

This function efficiently finds and kills Blood Raven to complete the Sisters Burial Grounds quest. The strategy involves pathing near the Mausoleum entrance to find her. Here's the detailed process:

### Setup

1. **Context Capture**: Stores the context in a local variable for convenience.

2. **Action Logging**: Sets the last action to "killRavenGetMerc" for debugging/tracking purposes.

3. **Travel to Cold Plains**: Uses the waypoint to travel to Cold Plains.
   - If this fails, returns an error with message "failed to move to Cold Plains"

4. **Move to Burial Grounds**: Travels from Cold Plains to the Burial Grounds area.
   - If this fails, returns an error with message "failed to move to Burial Grounds"

### Configuration Backup and Modification

5. **Save Original Settings**: Stores the current "back to town" configuration so it can be restored later.

6. **Modify Settings for Fight**: Changes settings specifically for the Blood Raven fight:
   - Disables the "return to town when out of mana potions" behavior (sets to false)
   - Sets healing potion usage threshold to 55 percent (more aggressive)

7. **Deferred Restoration**: Sets up a deferred function (runs at the end, even if there's an error) that:
   - Restores the original back-to-town configuration
   - Logs the message "Restored original back-to-town checks after Blood Raven fight."

### Locating Blood Raven

8. **Get Area Data**: Retrieves the data for the Burial Grounds area.

9. **Find Blood Raven's Position**: Searches the area's NPC data for Blood Raven (NPC ID 805):
   - If not found OR if there are no known positions, logs "Blood Raven position not found" and exits

10. **Move to Blood Raven**: Moves the character to Blood Raven's first known position (she has a fixed spawn location in the area).

### Combat Loop

11. **Fight Blood Raven**: Enters an infinite loop that:
    - Searches the current monster list for Blood Raven (using her NPC ID and checking for monster type "None")
    - If she's not found, breaks out of the loop (meaning she's dead)
    - If she's found, calls the character's kill monster sequence, passing:
      - A function that returns Blood Raven's Unit ID and true (confirming she's the target)
      - No additional filter (nil)
    - The loop continues until Blood Raven is dead

12. **Completion**: Returns nil (no error), indicating the quest was successfully completed.

---

## Supporting Function: gambleAct1Belt()

This function handles purchasing a 9-slot belt from Gheed through gambling. This is important because a larger belt means more potion storage. Here's the complete process:

### Eligibility Checks

1. **Level Check**: Checks if the player is between level 9 and 10 (inclusive):
   - If below level 9 OR level 11 or higher, logs "Not level 9 to 11, skipping belt gamble" and exits
   - This creates a specific window where belt gambling is worthwhile

2. **Check Equipped Belt**: Loops through all equipped items looking for any of these belt types:
   - Belt (basic 9-slot belt)
   - HeavyBelt (9-slot belt)
   - PlatedBelt (9-slot belt)
   - If any of these are found, logs "Already have a 9 slot belt equipped, skipping" and exits

3. **Check Inventory Belt**: Loops through all items in the inventory looking for the same belt types:
   - If any are found, logs "Already have a 9 slot belt in inventory, skipping" and exits

4. **Gold Check**: Checks if the player has at least 3,000 gold:
   - If not, logs "Not enough gold to buy a belt, skipping" and exits

### Gambling Process

5. **Visit Gheed**: Logs "No 12 slot belt found, trying to buy one from Gheed" (note: this is a typo in the log message—it should say "9 slot belt")
   - Interacts with Gheed, the gambling NPC
   - If interaction fails, returns the error
   - Sets up a deferred function to close all menus when done

6. **Open Gambling Window**: Simulates keyboard navigation:
   - Presses HOME (goes to top of menu)
   - Presses DOWN twice (navigates to gambling option)
   - Presses RETURN (selects gambling)
   - Waits 1000 milliseconds (1 second) for the window to open

7. **Verify Menu**: Checks if the NPC shop menu is actually open:
   - If not, logs a debug message "failed opening gambling window"

### Purchase Loop

8. **Define Target**: Creates a list with one item: "Belt"

9. **Infinite Gambling Loop**: Starts an infinite loop that:
   - **Search for Belt**: Loops through the list of items to gamble for:
     - Searches the vendor's inventory for an item named "Belt" at the vendor location
     - If found:
       - Calls a function to buy 1 of that item
       - Logs "Belt purchased, running AutoEquip"
       - Calls the auto-equip function to automatically equip the belt if it's better
       - If auto-equip fails, logs an error with the message "AutoEquip failed after buying belt"
       - Returns nil (success)

   - **Refresh If Not Found**: If no belt was found in the gambling window:
     - Logs "Desired items not found in gambling window, refreshing..."
     - Clicks the refresh button (position depends on whether using legacy graphics or modern graphics)
     - Waits 500 milliseconds (0.5 seconds)
     - Loop continues to check the refreshed inventory

This loop continues indefinitely until a belt is purchased.

---

## Supporting Function: atDistance()

This is a mathematical helper function that calculates a position at a specific distance along the line between two points. Here's how it works:

### Input Parameters

1. **start**: The starting position (has X and Y coordinates)
2. **end**: The ending position (has X and Y coordinates)
3. **distance**: The desired distance from the start point toward the end point

### Calculation Process

1. **Calculate Delta X and Y**: Determines the difference in X and Y coordinates:
   - dx = end.X - start.X (horizontal difference)
   - dy = end.Y - start.Y (vertical difference)
   - Converts these to floating-point numbers for precision

2. **Calculate Total Distance**: Uses the Pythagorean theorem to find the straight-line distance:
   - dist = square root of (dx² + dy²)

3. **Handle Zero Distance**: If the distance is zero (start and end are the same point):
   - Returns the start position unchanged

4. **Calculate Ratio**: Determines what fraction of the total distance to travel:
   - ratio = desired distance / actual distance between points

5. **Calculate New Position**: Applies the ratio to find the new coordinates:
   - newX = start.X + (dx × ratio)
   - newY = start.Y + (dy × ratio)

6. **Return Position**: Converts the floating-point coordinates back to integers and returns the new position.

### Example

If you have a start point at (0, 0) and an end point at (100, 0), and you want a position 30 units away from start:
- The function would return (30, 0)—a point 30 units along the line toward the end point.

---

## Supporting Function: shouldFarmCountessForRunes()

This function determines whether the character needs to farm the Countess for specific runes in Nightmare difficulty. It checks inventory and stash to see if the character has enough of each required rune. Here's the detailed logic:

### Required Runes Definition

1. **Define Requirements**: Creates a map of rune names to required quantities:
   - TalRune: needs 3
   - ThulRune: needs 2
   - OrtRune: needs 2
   - AmnRune: needs 2
   - TirRune: needs 1
   - SolRune: needs 3
   - RalRune: needs 2
   - NefRune: needs 2
   - ShaelRune: needs 3
   - IoRune: needs 1
   - EldRune: needs 1

These runes are used for important runewords like Insight, Spirit, Lore, and others that are critical for leveling.

### Inventory Check

2. **Create Owned Runes Map**: Initializes an empty map to track how many of each rune the character actually has.

3. **Get Items to Check**: Retrieves all items from three locations:
   - Player inventory
   - Personal stash
   - Shared stash

4. **Scan for Runes**: Logs "--- Checking for required runes ---" and then:
   - Loops through every item in those locations
   - Gets the item's name
   - Checks if that name matches any of the required runes
   - If it's a required rune:
     - Logs "Found a required rune: [rune name]. Incrementing count."
     - Increases the count for that rune in the owned runes map

5. **Log Final Counts**: Logs the complete map of owned rune counts for debugging.

### Comparison and Decision

6. **Check Each Requirement**: Loops through each required rune and its required quantity:
   - Compares the owned quantity to the required quantity
   - If the character has fewer runes than required:
     - Logs a message like "Missing runes, farming Countess. Need [X] of [rune name], but have [Y]."
     - Returns true (yes, farm Countess)

7. **All Requirements Met**: If the loop completes without finding any shortages:
   - Logs "All required runes are present. Skipping Countess farm."
   - Returns false (no need to farm Countess)

---

## Summary

The `leveling_act1.go` file creates an intelligent, adaptive leveling system that:

- Automatically configures settings for brand new characters
- Progresses through Act 1 quests in an optimized order
- Farms specific areas for gold, experience, and runes based on level and difficulty
- Purchases important equipment upgrades at the right time
- Adjusts gameplay settings dynamically based on character level and difficulty
- Handles transitions between acts seamlessly
- Manages different strategies for different character classes
- Ensures quest cleanup before moving to the next act

The system is designed to be hands-off, making decisions based on the character's current state, level, difficulty, gold, equipment, and quest progress to provide the most efficient leveling experience possible.
