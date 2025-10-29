# Detailed Explanation of leveling_act2.go

This document provides a comprehensive explanation of everything that happens in the `internal/run/leveling_act2.go` file, written in plain English.

## Overview

This file contains the logic for automatically leveling a character through Act 2 of Diablo 2. Act 2 is more complex than Act 1, involving multiple interlocking quests (finding the Horadric Cube, the Staff of Kings, the Amulet of the Viper, killing the Summoner, and finally defeating Duriel). The file orchestrates all these activities in the correct order while also handling gold farming, equipment purchases, and special character-specific needs.

---

## Main Function: act2()

The `act2()` function is the central orchestrator for all Act 2 leveling activities. It uses a priority-based system to decide what to do next. Here's the complete flow:

### Initial Setup and Validation

1. **Running Flag Check**:
   - The function starts by creating a local variable called `running` and setting it to false.
   - This variable is used to prevent the function from running multiple times simultaneously.

2. **Location and Running Check**:
   - Checks two conditions:
     - If the `running` flag is already true (meaning the function is already executing)
     - OR if the player is not in Lut Gholein (the Act 2 town)
   - If either condition is true, the function exits immediately and does nothing.

3. **Set Running Flag**:
   - Sets the `running` flag to true, indicating that the function is now executing.

4. **Quest Log Update**:
   - Updates the quest log to get the current state of all Act 2 quests.
   - The parameter `false` means it's not a verbose update.

### Equipment Purchase Phase

5. **Belt Purchase**:
   - Calls a helper function `buyAct2Belt()` to attempt to purchase a 12-slot belt.
   - If the belt purchase returns an error, the function exits with that error.

6. **Vendor Refill**:
   - Refills all potions at a vendor.
   - Parameters: `false` means don't force a refill, `true` means do buy potions.

7. **Necromancer Bone Wand Purchase**:
   - Calls a function to buy 2-socket Bone Wands specifically for Necromancers.
   - These wands are used to make the "White" runeword, which is excellent for Necromancer leveling.
   - If this fails, it logs an error but doesn't stop execution (the error is not returned).

8. **Get Player Level**:
   - Retrieves the player's current level for later checks.

### Priority 0: Act Completion Check (Highest Priority)

This is the first and most important check—if Act 2 is already done, move to Act 3.

9. **Seven Tombs Quest Completed Check**:
   - Checks if the Seven Tombs quest (killing Duriel) is completed AND the player is at least level 24.
   - If both conditions are true:
     - Logs "Act 2, The Seven Tombs quest completed. Moving to Act 3."
     - **Move to Meshif**: Moves to coordinates (5195, 5060) where Meshif (the ship captain) stands.
     - **Interact with Meshif**: Talks to Meshif.
     - **Menu Navigation**: Simulates keyboard input:
       - Presses HOME (goes to top of dialogue menu)
       - Presses DOWN (selects "travel to Act 3" option)
       - Presses RETURN/ENTER (confirms selection)
     - **Wait for Dialogue**: Waits 1000 milliseconds (1 second).
     - **Skip Cinematics**: Holds SPACE key for 2000 milliseconds (2 seconds) to skip the boat travel animation.
     - **Final Wait**: Waits another 1000 milliseconds.
     - **Exit Function**: Returns nil, indicating successful transition to Act 3.

### Gold Farming Logic

10. **Low Gold Detection**:
    - Calls a function to check if the player has low gold.
    - If gold is low, enters a difficulty-based farming strategy:
      - **Normal Difficulty**:
        - Calls a function to kill Radament for gold (Radament always drops gold piles).
        - Returns from the function after farming (will re-evaluate on next call).
      - **Hell Difficulty**:
        - Runs the Mausoleum farming route (in Act 1, which is more profitable).
        - After farming, uses the waypoint to return to Lut Gholein.
        - If the waypoint fails, logs an error message.
    - Note: If difficulty is Nightmare, no gold farming is done here.

### Priority 1: Post-Duriel Quest Turn-In

11. **Duriel Defeated But Not Reported Check**:
    - Checks if the Seven Tombs quest has status "InProgress6" (Duriel is dead but not yet turned in).
    - If this status is detected:
      - Logs "Act 2, The Seven Tombs quest completed. Need to talk to Meshif and then move to Act 3."
      - **Move to Jerhyn**: Moves to coordinates (5092, 5144).
      - **Interact with Jerhyn**: Talks to Jerhyn to officially complete the quest.
      - Logs "Act 2, The Seven Tombs quest completed. Moving to Act 3."
      - **Move to Meshif**: Moves to coordinates (5195, 5060).
      - **Interact with Meshif**: Talks to Meshif.
      - **Menu Navigation**: Same keyboard sequence as before (HOME, DOWN, RETURN).
      - **Wait for Dialogue**: Waits 1000 milliseconds.
      - **Skip Cinematics**: Holds SPACE for 2000 milliseconds.
      - **Final Wait**: Waits 1000 milliseconds.
      - **Exit Function**: Returns nil.

### Special Mercenary Hiring Logic (Nightmare Only)

12. **Frozen Aura Mercenary Hiring**:
    - This is a very complex sequence that only happens in Nightmare difficulty.
    - Checks three conditions:
      - Difficulty is Nightmare
      - The mercenary is currently alive (health percent > 0)
      - The configuration flag `ShouldHireAct2MercFrozenAura` is true
    - If all conditions are met, executes this sequence:

    **Step 1: Preparation**
    - Logs "Start Hiring merc with Frozen Aura"
    - **Drink All Potions**: Uses all potions in the inventory to make room for mercenary equipment.

    **Step 2: Unequip Current Mercenary**
    - Logs "Un-equipping merc"
    - Calls a function to remove all items from the current mercenary.
    - If unequipping fails, logs an error and returns that error.

    **Step 3: Graphics Mode Check**
    - The mercenary hiring list only works in Legacy graphics mode.
    - Checks if the game is NOT in legacy graphics mode.
    - If not in legacy mode:
      - Logs "Switching to legacy mode to hire merc"
      - Presses the legacy toggle key.
      - Waits 500 milliseconds for the graphics mode to switch.

    **Step 4: Interact with Mercenary Contractor**
    - Logs "Interacting with mercenary NPC"
    - Gets the town's mercenary contractor NPC (in Act 2, this is Greiz).
    - Interacts with that NPC.
    - If interaction fails, returns the error.
    - Simulates keyboard input: HOME, DOWN, RETURN (to open the hire mercenary dialogue).
    - Waits 2000 milliseconds (2 seconds) for the mercenary list to load.

    **Step 5: Get Mercenary List**
    - Logs "Getting merc list"
    - Reads the list of available mercenaries from game memory.

    **Step 6: Find Frozen Aura Mercenary**
    - Creates a variable to hold the mercenary to hire (starts as nil).
    - Loops through all available mercenaries in the list.
    - For each mercenary, checks if their skill ID is HolyFreeze (Frozen Aura).
    - If a mercenary with Holy Freeze is found:
      - Stores a reference to that mercenary.
      - Breaks out of the loop.
    - After the loop, if no mercenary with Frozen Aura was found:
      - Logs "No merc with Frozen Aura found, cannot hire"
      - Returns nil (exits without hiring).

    **Step 7: Hire the Mercenary**
    - Logs the mercenary's name and skill being hired.
    - Creates a keyboard sequence starting with HOME.
    - Loops through the mercenary's index position:
      - For each index, adds a DOWN arrow key press to the sequence.
      - This navigates to the correct position in the list.
    - Adds RETURN, UP, RETURN to the sequence:
      - First RETURN selects the mercenary.
      - UP moves to the "hire" option.
      - Second RETURN confirms the hiring.
    - Executes the complete keyboard sequence.

    **Step 8: Update Configuration**
    - Sets the `ShouldHireAct2MercFrozenAura` flag to false (so this doesn't happen again).

    **Step 9: Graphics Mode Restoration**
    - Checks if the game is NOT in classic mode AND is currently in legacy graphics mode.
    - If true:
      - Logs "Switching back to non-legacy mode"
      - Presses the legacy toggle key to switch back.
      - Waits 500 milliseconds.

    **Step 10: Save Configuration**
    - Saves the updated configuration to disk.
    - If saving fails, logs an error (but doesn't stop execution).

    **Step 11: Re-Equip Mercenary**
    - Logs "Merc hired successfully, re-equipping merc"
    - Calls the auto-equip function to give the new mercenary the best available gear.

### Priority 3: Pre-Duriel Readiness Check

This section checks if both prerequisite quests are complete and decides whether to fight Duriel or keep leveling.

13. **Quest Completion Status Check**:
    - Checks if the Horadric Staff quest is completed.
    - Checks if the Summoner quest is completed.

14. **Both Quests Complete Logic**:
    - If both quests are completed:
      - Logs "Horadric Staff and Summoner quests are completed. Proceeding to Duriel or leveling."
      - Logs the character's class.

      **Necromancer Special Check**:
      - If the character is a Necromancer:
        - Logs "Necromancer detected, checking for Bone Prison skill"
        - Checks if the Necromancer has the Bone Prison skill.
        - Logs the check results (whether they have it and what level).
        - If the Necromancer does NOT have Bone Prison OR the skill level is less than 1:
          - Logs "Necromancer needs Bone Prison skill before Duriel. Farming tombs."
          - Runs the Tal Rasha's Tombs farming route to level up and get Bone Prison.
          - Returns from the function (will re-evaluate next time).
        - If the Necromancer has Bone Prison:
          - Logs "Necromancer has Bone Prison, ready for Duriel."

      **All Other Classes Level Check**:
      - If the character is NOT a Necromancer:
        - Gets the player's current level.
        - Logs the level check.
        - If the level is less than 24:
          - Logs "Player not level 24, farming tombs before Duriel."
          - Runs the Tal Rasha's Tombs farming route to reach level 24.
          - Returns from the function.

      **Proceed to Duriel**:
      - If all readiness checks pass (level 24 for most classes, or Bone Prison for Necromancer):
        - Calls `prepareStaff()` to create the Horadric Staff from its components.
        - Calls `duriel()` to fight Duriel.
        - Returns the result.

### Quest Progression: Horadric Cube

15. **Horadric Cube Check**:
    - Searches the inventory and stash for the Horadric Cube.
    - If the cube is found:
      - Logs "Horadric Cube found, skipping quest"
    - If the cube is NOT found:
      - Logs "Horadric Cube not found, starting quest"
      - Calls a function to complete the Horadric Cube quest (go to Halls of the Dead, get the cube).
      - Returns from the function.

### Quest Progression: Radament (Conditional)

16. **Radament Quest Check**:
    - Gets the player's current level.
    - Checks two conditions (connected by OR):
      - Player level is less than 18
      - OR (Difficulty is Nightmare AND Radament quest is not completed)
    - If either condition is true:
      - Logs "Starting Radament."
      - Calls a function to kill Radament (this gives a skill point book).
      - Returns from the function.

### Book of Skill Usage

17. **Book of Skill Check**:
    - Searches the inventory for a "BookofSkill" item.
    - If found:
      - Logs "BookofSkill found in inventory. Using it..."
      - **Menu Preparation**: Closes all open menus.
      - **Open Inventory**: Presses the inventory keybinding.
      - **Get Book Position**: Calculates the screen coordinates for the book.
      - **Wait**: Waits 200 milliseconds for the inventory to open.
      - **Right Click Book**: Right-clicks the book to use it (grants a skill point).
      - **Close Menus**: Closes all menus.
      - Logs "Book of Skill used successfully."

### Quest Progression: Horadric Staff Quest

18. **Drognan Interaction for Staff Quest**:
    - Checks if the Horadric Staff quest has status "InProgress4".
    - If yes, interacts with Drognan (this advances certain quest dialogue states).

19. **Staff of Kings Search**:
    - Checks if the Horadric Staff quest is NOT completed.
    - If not completed:
      - Searches for the complete "HoradricStaff" in inventory, stash, or equipped.
      - Searches for "StaffOfKings" in inventory, stash, or equipped.
      - If either the Staff of Kings OR the complete Horadric Staff is found:
        - Logs "StaffOfKings found, skipping quest"
      - If neither is found:
        - Logs "StaffOfKings not found, starting quest"
        - Calls `findStaff()` to go to the Maggot Lair and get the Staff of Kings.
        - Returns from the function.

20. **Amulet of the Viper Search**:
    - Still within the "Horadric Staff quest not completed" check:
      - Searches for "AmuletOfTheViper" in inventory, stash, or equipped.
      - If the amulet OR the complete Horadric Staff is found:
        - Logs "Amulet of the Viper found, skipping quest"
      - If neither is found:
        - Logs "Amulet of the Viper not found, starting quest"
        - Calls `findAmulet()` to go to the Claw Viper Temple and get the amulet.
        - Returns from the function.

### Quest Progression: Summoner Quest

21. **Summoner Quest Not Completed Check**:
    - Checks if the Summoner quest is NOT completed AND the Seven Tombs quest has status "QuestNotStarted".
    - If both conditions are true:
      - Logs "Starting summoner quest (Summoner not yet completed)."
      - **Interact with Drognan**: Talks to Drognan (this may be required for quest progression).
      - **Run Summoner**: Calls the Summoner run function to navigate to the Arcane Sanctuary and kill the Summoner.
      - If the Summoner run returns an error, returns that error.

      **Journal Interaction Sequence** (note: the code comments indicate this block is temporary):
      - **Find Journal**: Searches for the "YetAnotherTome" object (the journal that the Summoner drops).
      - If not found:
        - Logs an error: "YetAnotherTome (journal) not found after Summoner kill. This is unexpected."
        - Returns the error.
      - If found:
        - Logs "Interacting with the journal to open the portal."
        - **Interact with Journal**: Interacts with the journal object.
        - The interaction waits until a "PermanentTownPortal" object appears (the red portal).
        - If interaction fails, returns the error.
        - Logs "Moving to red portal"
        - **Find Portal**: Gets the permanent town portal object.
        - **Use Portal**: Interacts with the portal.
        - The interaction waits until the player is in the Canyon of the Magi and inside the area boundaries.
        - If interaction fails, returns the error.

      **Waypoint Discovery**:
      - Logs "Discovering Canyon of the Magi waypoint."
      - Calls a function to discover and activate the waypoint in the Canyon of the Magi.
      - If discovery fails, returns the error.
      - Logs "Summoner quest chain (journal, portal, WP) completed."
      - Updates the quest log.
      - Returns nil (will re-evaluate on next call).

### Quest Progression: Tombs Exploration

22. **Summoner Complete But Tombs Not Started**:
    - Checks if the Summoner quest is completed AND the Seven Tombs quest has status "QuestNotStarted".
    - If both conditions are true:
      - **Run Tal Rasha's Tombs**: Explores the tombs to find the correct one (contains Duriel's chamber).
      - If the tomb run returns an error, returns that error.

      **Journal Interaction Sequence** (identical to the previous journal sequence):
      - This appears to be duplicate code (the comments say it can be removed).
      - Performs the exact same journal → portal → waypoint sequence as in step 21.

      **Post-Tombs Actions**:
      - Logs "Discovering Canyon of the Magi waypoint."
      - Discovers the waypoint.
      - Logs "Summoner quest chain (journal, portal, WP) completed."
      - Updates the quest log.
      - Returns nil.

### Final Priority: Duriel Readiness (Second Check)

23. **Seven Tombs Quest Started Check**:
    - Checks if the Seven Tombs quest does NOT have status "QuestNotStarted" (meaning it's been started).
    - If the quest is in progress:
      - Logs "Character class check" with the class name.

      **Necromancer Bone Prison Check** (identical to step 14):
      - If Necromancer:
        - Logs "Necromancer detected, checking for Bone Prison skill"
        - Checks for Bone Prison skill and its level.
        - Logs the check results.
        - If no Bone Prison or level less than 1:
          - Logs "Necromancer needs Bone Prison skill before Duriel. Farming tombs."
          - Runs Tal Rasha's Tombs.
          - Returns from the function.
        - Logs "Necromancer has Bone Prison, ready for Duriel."

      **Other Classes Level Check** (identical to step 14):
      - If NOT Necromancer:
        - Gets player level.
        - Logs the level check.
        - If level less than 24:
          - Logs "Player not level 24, farming tombs before Duriel."
          - Runs Tal Rasha's Tombs.
          - Returns from the function.

      **Proceed to Duriel**:
      - Calls `prepareStaff()`.
      - Calls `duriel()`.
      - Returns the result.

### Fallback

24. **No Action Triggered**:
    - If none of the above conditions triggered an action:
      - Logs a debug message: "act2() function completed, no specific action triggered this tick. Returning nil."
      - Returns nil.

---

## Supporting Function: findStaff()

This function navigates to the Maggot Lair and retrieves the Staff of Kings. Here's the complete sequence:

### Navigation Phase

1. **Waypoint to Far Oasis**:
   - Uses the waypoint system to travel to the Far Oasis.
   - If this fails, returns the error immediately.

2. **Enter Maggot Lair Level 1**:
   - Moves from the Far Oasis to the entrance of the Maggot Lair.
   - If this fails, returns the error.

3. **Descend to Level 2**:
   - Moves from Level 1 to Level 2 of the Maggot Lair.
   - If this fails, returns the error.

4. **Descend to Level 3**:
   - Moves from Level 2 to Level 3 (the deepest level).
   - If this fails, returns the error.

### Chest Location Phase

5. **Find Staff Chest**:
   - Calls a movement function that takes a position-finding callback.
   - The callback function:
     - Searches for the "StaffOfKingsChest" object.
     - If found:
       - Logs "Staff Of Kings chest found, moving to that room"
       - Returns the chest's position and true (found).
     - If not found:
       - Returns an empty position and false (not found).
   - The movement function navigates to the returned position.
   - If movement fails, returns the error.

### Combat Phase

6. **Clear Area (Non-Hell Only)**:
   - Checks if the difficulty is NOT Hell.
   - If not Hell:
     - Clears all monsters within 15 units of the player.
     - Uses a filter that accepts any monster type.

### Chest Interaction Phase

7. **Relocate Chest**:
   - Searches for the Staff of Kings chest object again (to get updated state).
   - If not found, returns an error.

8. **Open Chest**:
   - Interacts with the chest object.
   - The interaction waits until the chest becomes non-selectable (meaning it's been opened).
   - Specifically, it repeatedly checks:
     - Finds the chest object.
     - If found, checks if it's NOT selectable anymore.
     - Returns true when the chest is no longer selectable (opened).
   - If interaction fails, returns the error.

### Item Pickup Phase

9. **Wait for Loot**:
   - Waits 200 milliseconds for the staff to drop on the ground.

10. **Pick Up Staff**:
    - Calls the item pickup function.
    - The parameter `-1` means pick up all items in range (not just one specific item).

11. **Success**:
    - Returns nil (no error).

---

## Supporting Function: findAmulet()

This function navigates to the Claw Viper Temple and retrieves the Amulet of the Viper. Here's the complete sequence:

### Setup Phase

1. **Update Quest Log**:
   - Updates the quest log to ensure the game knows you're on this quest.

2. **Interact with Drognan**:
   - Talks to Drognan (may be required for quest state progression).

### Navigation Phase

3. **Waypoint to Lost City**:
   - Uses the waypoint to travel to the Lost City.
   - If this fails, returns the error.

4. **Move to Valley of Snakes**:
   - Moves from Lost City to the Valley of Snakes.
   - If this fails, returns the error.

5. **Enter Claw Viper Temple Level 1**:
   - Moves from Valley of Snakes to the temple entrance.
   - If this fails, returns the error.

6. **Descend to Level 2**:
   - Moves to the second level of the temple.
   - If this fails, returns the error.

### Altar Location Phase

7. **Find Tainted Sun Altar**:
   - Calls a movement function with a position-finding callback.
   - The callback function:
     - Searches for the "TaintedSunAltar" object (the altar where the amulet is placed).
     - If found:
       - Logs "Tainted Sun Altar found, moving to that room"
       - Returns the altar's position and true.
     - If not found:
       - Returns an empty position and false.
   - Navigates to the altar's position.
   - If movement fails, returns the error.

### Combat Phase (Commented Out)

8. **Area Clearing (Disabled)**:
   - There's a commented-out line that would clear monsters around the player.
   - This is currently disabled (probably because the boss fight is handled differently).

### Altar Interaction Phase

9. **Relocate Altar**:
   - Searches for the Tainted Sun Altar object.
   - If not found, returns an error.

10. **Interact with Altar**:
    - Interacts with the altar object.
    - The interaction waits until the altar becomes non-selectable (the amulet is removed).
    - The callback function:
      - Repeatedly finds the altar object.
      - If found:
        - Checks if it's NOT selectable (meaning interaction succeeded).
        - Logs "Interacted with Tainted Sun Altar" when it becomes non-selectable.
      - Returns true when non-selectable.
    - If interaction fails, returns the error.

### Quest Completion Phase

11. **Return to Town**:
    - Uses a town portal or waypoint to return to Lut Gholein.

12. **Interact with Drognan Again**:
    - Talks to Drognan.
    - The comment says "This stops us being blocked from getting into Palace"
    - This is necessary because talking to Drognan after getting the amulet triggers dialogue that allows you to enter Jerhyn's palace.

13. **Update Quest Log**:
    - Updates the quest log to reflect the amulet has been retrieved.

14. **Success**:
    - Returns nil (no error).

---

## Supporting Function: prepareStaff()

This function combines the Staff of Kings and Amulet of the Viper in the Horadric Cube to create the complete Horadric Staff. Here's the detailed logic:

### Check for Complete Staff

1. **Search for Horadric Staff**:
   - Searches for the complete "HoradricStaff" item in inventory, stash, and equipped slots.

2. **Staff Already Complete Check**:
   - If the complete staff is found:
     - Logs "Horadric Staff found!"
     - **Stash Location Check**:
       - Checks if the staff is in the stash (not inventory or equipped).
       - If it's in the stash:
         - Logs "It's in the stash, let's pick it up"
         - **Find Bank Object**: Searches for the stash (Bank) object.
         - If not found, logs "bank object not found"
         - **Open Stash**: Interacts with the bank object, waiting until the stash menu opens.
         - If opening fails, returns the error.
         - **Retrieve Staff**: Gets the screen coordinates for the staff in the stash.
         - Performs a Ctrl+Left-Click on the staff (this moves it from stash to inventory).
         - Waits 300 milliseconds for the transfer.
         - **Close Menus**: Closes all open menus.
         - Returns nil (staff is now in inventory, ready for Duriel).

### Component Gathering

3. **Search for Staff of Kings**:
   - Searches for "StaffOfKings" in inventory, stash, and equipped slots.
   - If NOT found:
     - Logs "Staff of Kings not found, skipping"
     - Returns nil (can't make the staff without this component).

4. **Search for Amulet**:
   - Searches for "AmuletOfTheViper" in inventory, stash, and equipped slots.
   - If NOT found:
     - Logs "Amulet of the Viper not found, skipping"
     - Returns nil (can't make the staff without this component).

### Cube Recipe Execution

5. **Add Items to Cube**:
   - Calls a function to place both the staff and amulet into the Horadric Cube.
   - If this fails, returns the error.

6. **Transmute**:
   - Calls a function to perform the cube transmutation.
   - This combines the Staff of Kings and Amulet of the Viper into the complete Horadric Staff.
   - If this fails, returns the error.

7. **Success**:
   - Returns nil (the Horadric Staff is now created).

---

## Supporting Function: duriel()

This function handles the entire Duriel boss fight sequence. Here's what happens:

### Preparation Phase

1. **Logging**:
   - Logs "Starting Duriel...."

2. **Thawing Potion Configuration**:
   - Sets the game configuration to use thawing potions during the Duriel fight.
   - Thawing potions provide cold resistance, which is crucial because Duriel deals cold damage and applies a chilling effect.

3. **Run Duriel Encounter**:
   - Calls the Duriel run function (this handles navigation to Duriel's chamber, the fight mechanics, etc.).
   - If the run returns an error, returns that error immediately.

### Post-Fight Verification

4. **Clear Remaining Enemies**:
   - Clears any remaining monsters within 30 units of the player.
   - Uses a special "Duriel filter" that only targets Duriel (in case he's still alive somehow).

5. **Duriel Death Check**:
   - Searches for Duriel in the monster list.
   - Checks three conditions (connected by OR):
     - Duriel is not found in the monster list
     - OR Duriel's life stat is zero or less (he's dead)
     - OR the Seven Tombs quest has status "InProgress3" (quest state after Duriel dies)
   - If ANY of these conditions are true, Duriel is confirmed dead:

### Tyrael Interaction

6. **Move to Tyrael**:
   - Moves to specific coordinates (22577, 15600).
   - This is where Tyrael appears after Duriel is defeated.

7. **Talk to Tyrael**:
   - Interacts with Tyrael (the angel who gives you the quest completion dialogue).

### Cleanup Phase

8. **Return to Town**:
   - Uses a town portal or waypoint to return to Lut Gholein.

9. **Update Quest Log**:
   - Updates the quest log to reflect that Duriel has been defeated.

10. **Success**:
    - Returns nil (Act 2 is essentially complete, though you still need to talk to Jerhyn and Meshif to move to Act 3).

---

## Supporting Function: buyAct2Belt()

This function attempts to purchase a 12-slot belt from Fara in Lut Gholein. Here's the complete logic:

### Inventory Check Phase

1. **Check Equipped Belt**:
   - Loops through all equipped items.
   - For each item, checks if it's one of these belt types:
     - "Belt" (basic 12-slot belt)
     - "HeavyBelt" (12-slot belt with better stats)
     - "PlatedBelt" (12-slot belt with even better stats)
   - If any of these are found:
     - Logs "Already have a 12+ slot belt equipped, skipping."
     - Returns nil (no purchase needed).

2. **Check Inventory Belt**:
   - Loops through all items in the inventory.
   - Checks for the same belt types as above.
   - If any are found:
     - Logs "Already have a 12+ slot belt in inventory, skipping."
     - Returns nil.

### Gold Check

3. **Verify Gold Amount**:
   - Checks if the player has at least 1000 gold.
   - If less than 1000 gold:
     - Logs "Not enough gold to buy a belt, skipping."
     - Returns nil (can't afford a belt).

### Vendor Interaction Phase

4. **Visit Fara**:
   - Logs "No 12 slot belt found, trying to buy one from Fara."
   - Interacts with Fara (the armor and weapon vendor in Act 2).
   - If interaction fails, returns the error.
   - Sets up a deferred function to close all menus when done (ensures cleanup happens even if there's an error).

5. **Open Shop Menu**:
   - Simulates keyboard input:
     - HOME (goes to top of dialogue menu)
     - DOWN (selects "trade" option)
     - RETURN (opens the shop)
   - Waits 1000 milliseconds (1 second) for the shop to open.

6. **Switch to Armor Tab**:
   - Calls a function to switch to tab 1 (armor tab).
   - Belts are on the armor tab, not the weapon tab.

7. **Refresh Game Data**:
   - Refreshes the game data to see the updated vendor inventory.
   - Waits 500 milliseconds for the data to update.

### Belt Purchase Phase

8. **Search Vendor Inventory**:
   - Loops through all items in the vendor's inventory (location type: Vendor).
   - For each item:
     - Checks if the item name is "Belt" (the basic 12-slot belt).
     - If it's a belt:
       - Gets the item's strength requirement.
       - Logs "Vendor item found" with the name and strength requirement.
       - **Strength Check**: Checks if the strength requirement is 25 or less.
       - If the requirement is low enough:
         - Logs "Found a suitable belt, buying it."
         - **Buy Belt**: Calls a function to purchase 1 of that belt.
         - Logs "Belt purchased, running AutoEquip."
         - **Auto-Equip**: Calls the auto-equip function to automatically equip the belt if it's better than the current one.
         - If auto-equip fails, logs an error.
         - Returns nil (purchase successful).

9. **No Belt Found**:
   - If the loop completes without finding a suitable belt:
     - Logs "No suitable belt found at Fara."
     - Returns nil.

---

## Supporting Function: RockyWaste()

This function provides a gold farming option by clearing the Rocky Waste area. Here's what it does:

### Navigation Phase

1. **Logging**:
   - Logs "Entering Rocky Waste for gold farming..."

2. **Move to Rocky Waste**:
   - Uses the area movement function to navigate to Rocky Waste.
   - This automatically handles pathing from wherever the player is to the Rocky Waste.
   - If movement fails:
     - Logs an error with the failure details.
     - Returns the error.

3. **Success Log**:
   - Logs "Successfully reached Rocky Waste."

### Clearing Phase

4. **Clear the Area**:
   - Calls the clear current level function.
   - Parameters:
     - `false` means don't open chests/containers
     - Uses a monster filter that accepts any monster type
   - If clearing fails:
     - Logs an error with the failure details.
     - Returns the error.

5. **Completion Log**:
   - Logs "Successfully cleared Rocky Waste area."

6. **Success**:
   - Returns nil (farming complete).

---

## Supporting Function: FarOasis()

This function provides another farming option by clearing the Far Oasis area. Here's what it does:

### Navigation Phase

1. **Waypoint to Far Oasis**:
   - Uses the waypoint system to travel directly to Far Oasis.
   - Note: This function doesn't check if the waypoint call succeeded (potential bug).

### Clearing Phase

2. **Clear the Area**:
   - Calls the clear current level function.
   - Parameters:
     - `false` means don't open chests/containers
     - Uses a monster filter that only targets elite monsters (champions, uniques, super uniques)
   - This is more efficient than clearing all monsters—you only kill elite packs for better experience and loot.
   - If clearing fails:
     - Logs an error with the failure details.
     - Returns the error.

3. **Success**:
   - Returns nil (farming complete).

---

## Supporting Function: DurielFilter()

This function creates a custom monster filter specifically for finding Duriel. Here's how it works:

### Function Structure

1. **Return Type**:
   - The function returns a "MonsterFilter" type.
   - A MonsterFilter is itself a function that takes a list of monsters and returns a filtered list.

2. **Filter Implementation**:
   - The returned function takes a parameter called `a` (a collection of all monsters).
   - Creates an empty list called `filteredMonsters`.
   - Loops through all monsters in the input collection:
     - For each monster, checks if its name equals "Duriel".
     - If yes, adds that monster to the filtered list.
   - Returns the filtered list.

### Purpose

This filter is used after the Duriel fight to specifically target only Duriel if he's still somehow alive, ignoring all other monsters in the area.

---

## Summary

The `leveling_act2.go` file implements a sophisticated priority-based system for Act 2 progression:

### Quest Orchestration
- **Smart Sequencing**: Handles complex interdependent quests (Horadric Cube → Staff components → Combine staff → Summoner → Tombs → Duriel)
- **Component Tracking**: Individually tracks the Staff of Kings, Amulet of the Viper, and completed Horadric Staff
- **Quest State Management**: Uses multiple quest statuses to determine exact progression state

### Character Readiness System
- **Level Requirements**: Most classes must reach level 24 before fighting Duriel
- **Necromancer Special Case**: Necromancers must have Bone Prison skill instead of level 24 (critical for Duriel fight)
- **Farming Loops**: Automatically farms Tal Rasha's Tombs until readiness conditions are met

### Equipment Management
- **Belt Upgrades**: Purchases 12-slot belts with appropriate strength requirements
- **Necromancer Support**: Buys 2-socket Bone Wands for the White runeword
- **Mercenary Optimization**: Special logic to hire Holy Freeze mercenary in Nightmare difficulty

### Gold Farming
- **Difficulty-Specific Strategies**:
  - Normal: Radament farming
  - Nightmare: No specific farming (relies on natural gold acquisition)
  - Hell: Mausoleum farming (higher returns)

### Complex Interactions
- **Frozen Aura Mercenary**: Multi-step process involving graphics mode switching, mercenary list reading, and precise hiring
- **Portal Mechanics**: Handles the Summoner's portal to Canyon of the Magi
- **Tyrael Dialogue**: Post-Duriel interaction to complete the act

### Act Transition
- **Multiple Exit Points**: Handles both immediate completion (if returning to Act 2) and post-Duriel completion
- **Jerhyn Quest Turn-in**: Properly completes the quest dialogue chain
- **Meshif Travel**: Automated dialogue navigation to sail to Act 3

The system is designed to handle all possible Act 2 states and progression paths, ensuring smooth leveling regardless of where the character is in the act's quest sequence.
