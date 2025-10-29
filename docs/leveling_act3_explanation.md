# Detailed Explanation of leveling_act3.go

This document provides a comprehensive explanation of everything that happens in the `internal/run/leveling_act3.go` file, written in plain English.

## Overview

This file contains the logic for automatically leveling a character through Act 3 of Diablo 2. Act 3 involves collecting the four pieces of Khalim's relics (Eye, Brain, Heart, and Flail), combining them into Khalim's Will, using it to open the Durance of Hate, and ultimately defeating Mephisto. The file also handles the Golden Bird side quest and gold farming strategies.

---

## Main Function: act3()

The `act3()` function is the central orchestrator for all Act 3 leveling activities. It uses a priority-based system similar to Acts 1 and 2. Here's the complete flow:

### Initial Setup and Validation

1. **Running Flag Check**:
   - Creates a local variable called `running` and sets it to false.
   - This variable prevents the function from running multiple times simultaneously.

2. **Location and Running Check**:
   - Checks two conditions:
     - If the `running` flag is already true (function is already executing)
     - OR if the player is not in Kurast Docks (the Act 3 town)
   - If either condition is true, the function exits immediately and does nothing.

3. **Quest Log Update**:
   - Updates the quest log to get the current state of all Act 3 quests.
   - The parameter `false` means it's not a verbose update.

### Hratli Special Interaction

4. **Hratli Pier Check**:
   - When you first arrive in Act 3, Hratli (an NPC) is at the pier instead of his normal location.
   - Searches for Hratli in the monster/NPC list.
   - If Hratli is found:
     - Interacts with him (talks to him).
     - This causes Hratli to move to his normal position in town later.
   - This is important because you need access to Hratli's shop and quest dialogues.

5. **Set Running Flag**:
   - Sets the `running` flag to true, indicating that the function is now executing.

### Equipment and Quest Preparation

6. **Vendor Refill**:
   - Refills all potions at a vendor.
   - Parameters: `false` means don't force a refill, `true` means do buy potions.

7. **Golden Bird Quest Check**:
   - Searches the inventory for a "PotionOfLife" item.
   - Gets the status of the Golden Bird quest.
   - Checks two conditions (connected by OR):
     - The quest is completed AND a Potion of Life was found in inventory
     - OR the quest is in progress (not "QuestNotStarted") AND not yet completed
   - If either condition is true:
     - Calls the `jadefigurine()` function to handle the Golden Bird quest chain.

### Gold Farming Logic

8. **Low Gold Detection**:
   - Calls a function to check if the player has low gold.
   - If gold is low:
     - Logs "Low on gold. Initiating Lower Kurast Chests gold farm."
     - Runs the Lower Kurast Chest farming route (opens super chests for gold and items).
     - If the farming run returns an error:
       - Logs the error.
       - Returns the error (exits the function).
     - Logs "Lower Kurast Chests gold farming completed. Quitting current run to re-evaluate in next game."
     - Returns nil (exits to start a new game with more gold).

9. **Hell Difficulty Additional Farming**:
   - Checks if the difficulty is Hell.
   - If yes:
     - Runs the Mausoleum farming route (in Act 1, for additional gold/experience).
     - After farming, uses the waypoint to return to Kurast Docks.
     - If the waypoint fails, logs an error message.

### Priority 0: Act Completion Check (Highest Priority)

This checks if Act 3 is already complete and moves to Act 4 if so.

10. **Guardian Quest Completed Check**:
    - Checks if The Guardian quest (killing Mephisto) is completed.
    - If completed:
      - Logs "Attempting to reach Act 4 via The Pandemonium Fortress waypoint."
      - Attempts to use the waypoint to travel directly to The Pandemonium Fortress (Act 4).
      - If the waypoint succeeds:
        - Logs "Successfully reached Act 4 via waypoint. Ending Act 3 script."
        - Returns nil (Act 3 is done).
      - If the waypoint fails:
        - Logs "Could not use waypoint to The Pandemonium Fortress. Falling back to manual portal entry."
        - **Manual Portal Route**:
          - Uses waypoint to Durance of Hate Level 2.
          - If waypoint fails, returns the error.
          - Moves to Durance of Hate Level 3.
          - If movement fails, returns the error.
          - Logs "Moving to bridge"
          - Moves to coordinates (17588, 8068) - this is the bridge that rises after Mephisto dies.
          - Waits 1000 milliseconds (1 second) for the bridge to finish rising.
          - Logs "Moving to red portal"
          - Finds the HellGate object (the red portal to Act 4).
          - Moves to the portal's position.
          - Interacts with the portal, waiting until the player's area changes to The Pandemonium Fortress.
          - Waits 500 milliseconds.
          - Holds the SPACE key for 3000 milliseconds (3 seconds) to skip cinematics.
          - Waits another 500 milliseconds.
          - Returns nil (successfully moved to Act 4).

### Priority 1: Khalim's Will Completion Check

This section handles the case where all relics are collected and combined.

11. **Search for Khalim's Will**:
    - Searches for "KhalimsWill" in inventory, stash, or equipped slots.
    - Stores the result in a variable `willFound`.

12. **Khalim's Will Quest Completed Check**:
    - Checks if the Khalim's Will quest is marked as completed.
    - If completed:
      - Logs "Khalims will completed. Starting Mephisto."
      - **Mephisto Configuration**:
        - Sets Mephisto configuration to NOT open chests.
        - Sets Mephisto configuration to NOT kill council members.
        - Sets Mephisto configuration to exit to Act 4 after killing him.
      - **Run Mephisto**:
        - Calls the Mephisto run function.
        - If the run returns an error:
          - Logs "Mephisto run failed, ending Act 3 script."
          - Returns the error.
      - **Post-Mephisto Check**:
        - Checks if The Guardian quest is now completed.
        - If completed:
          - Searches for the HellGate object (portal to Act 4).
          - If not found:
            - Logs "Gate to Pandemonium Fortress not found after killing Mephisto. Ending script."
            - Returns nil.
          - If found:
            - Interacts with the Hell Gate.
            - The interaction callback:
              - Waits 500 milliseconds.
              - Checks if the player's area is now The Pandemonium Fortress.
              - Returns true when the player has transitioned.
            - If interaction fails:
              - Logs an error.
              - Returns the error.
            - Logs "Successfully interacted with Hell Gate. Attempting to skip cinematic."
            - Waits 500 milliseconds.
            - Holds SPACE key for 3000 milliseconds to skip cinematic.
            - Waits 500 milliseconds.
            - Logs "Successfully attempted to enter Act 4. Ending Act 3 script."
            - Returns nil.
        - If The Guardian quest is NOT completed:
          - Logs "Mephisto run completed, but 'The Guardian' quest is not marked as complete. Ending Act 3 script."
          - Returns nil.

### Priority 2: Khalim's Will Already Found (Skip Collection)

This section handles the case where Khalim's Will is already in inventory/stash but the stairs might not be opened yet.

13. **Will Found But Quest Not Marked Complete**:
    - This block only runs if the previous check (quest completed) didn't trigger.
    - If `willFound` is true:
      - Logs "Khalim's Will found (or quest completed), skipping Khalims Will quests"
      - Calls `openMephistoStairs()` to use Khalim's Will on the Compelling Orb.
      - If opening stairs returns an error, returns that error.
      - **Post-Stairs Check**:
        - Checks if The Guardian quest is completed.
        - If completed:
          - Searches for the HellGate object.
          - If not found:
            - Logs "Gate to Pandemonium Fortress not found after using Khalim's Will. Ending script."
            - Returns nil.
          - If found:
            - Interacts with the Hell Gate.
            - The interaction callback:
              - Waits 500 milliseconds.
              - Waits another 1000 milliseconds.
              - Holds SPACE key for 3000 milliseconds.
              - Waits 1000 milliseconds.
              - Checks if area is The Pandemonium Fortress.
            - If interaction fails:
              - Waits 1000 milliseconds.
              - Holds SPACE key for 3000 milliseconds (retry).
              - Waits 1000 milliseconds.
              - Returns the error.
            - If successful:
              - Waits 1000 milliseconds.
              - Holds SPACE key for 3000 milliseconds (final cinematic skip).
              - Waits 1000 milliseconds.
              - Returns nil.
      - If The Guardian quest is NOT completed:
        - Logs "Khalim's Will found and used, but Mephisto not yet dead. Ending Act 3 script for this run."
        - Returns nil.

### Relic Collection Phase

The following sections collect each of the four relics needed to create Khalim's Will.

#### Khalim's Eye Collection

14. **Search for Khalim's Eye**:
    - Searches for "KhalimsEye" in inventory and stash.
    - If found:
      - Logs "KhalimsEye found, skipping quest"
    - If NOT found AND Khalim's Will is NOT already found AND the Khalim's Will quest is NOT completed:
      - Logs "KhalimsEye not found, starting quest"
      - Calls `findKhalimsEye()` to go to Spider Cavern and retrieve the eye.

#### Khalim's Brain Collection

15. **Search for Khalim's Brain**:
    - Searches for "KhalimsBrain" in inventory and stash.
    - If found:
      - Logs "KhalimsBrain found, skipping quest"
    - If NOT found AND Khalim's Will is NOT already found AND the Khalim's Will quest is NOT completed:
      - Logs "KhalimsBrain not found, starting quest"
      - Runs the Endugu (Flayer Jungle) encounter.
      - **Pickup Retry**: The comment notes "Sometimes it doesn't pick up the brain"
        - Waits 500 milliseconds.
        - Calls item pickup with range 10 to ensure the brain is picked up.

#### Khalim's Heart Collection

16. **Search for Khalim's Heart**:
    - Searches for "KhalimsHeart" in inventory and stash.
    - If found:
      - Logs "KhalimsHeart found, skipping quest"
    - If NOT found AND Khalim's Will is NOT already found AND the Khalim's Will quest is NOT completed:
      - Logs "KhalimsHeart not found, starting quest"
      - Calls `findKhalimsHeart()` to go to the Sewers and retrieve the heart.

#### Khalim's Flail Collection

17. **Search for Khalim's Flail**:
    - Searches for "KhalimsFlail" in inventory and stash.
    - If found:
      - **Flail Found Logic**:
        - The comment notes "Trav" (Travincal, where the flail quest takes place).
        - The comment explains: "If flail is found, it means all parts are likely collected, so we try to open Mephisto stairs."
        - Calls `openMephistoStairs()`.
        - **Post-Stairs Check**:
          - Checks if The Guardian quest is completed.
          - If completed:
            - Searches for the HellGate object.
            - If not found:
              - Logs "Gate to Pandemonium Fortress not found after using Khalim's Will. Ending script."
              - Returns nil.
            - If found:
              - Interacts with the Hell Gate.
              - The interaction callback:
                - Waits 500 milliseconds.
                - Waits 1000 milliseconds.
                - Holds SPACE for 2000 milliseconds (note: different timing than previous portal interactions).
                - Waits 1000 milliseconds.
                - Checks if area is The Pandemonium Fortress.
              - If interaction fails:
                - Waits 1000 milliseconds.
                - Holds SPACE for 2000 milliseconds (retry).
                - Waits 1000 milliseconds.
                - Returns the error.
              - If successful:
                - Waits 1000 milliseconds.
                - Holds SPACE for 2000 milliseconds (final skip).
                - Waits 1000 milliseconds.
                - Returns nil.
        - If The Guardian quest is NOT completed:
          - Logs "Khalim's Flail found and Mephisto stairs opened, but Mephisto not yet dead. Ending Act 3 script for this run."
          - Returns nil.
    - If NOT found AND Khalim's Will is NOT already found AND the Khalim's Will quest is NOT completed:
      - Logs "KhalimsFlail not found, starting quest"
      - Runs the Travincal encounter to kill the Council and get the flail.
      - If Travincal run returns an error, returns that error.

18. **Function End**:
    - Returns nil (all checks complete).

---

## Supporting Function: findKhalimsEye()

This function navigates to the Spider Cavern and retrieves Khalim's Eye. Here's the complete sequence:

### Navigation Phase

1. **Waypoint to Spider Forest**:
   - Uses the waypoint system to travel to Spider Forest.
   - If this fails, returns the error.

2. **Buff Before Entering Dungeon**:
   - Calls the buff function to apply all active buffs (armor, Energy Shield, etc.).

3. **Enter Spider Cavern**:
   - Moves from Spider Forest to the Spider Cavern entrance.
   - If this fails, returns the error.

4. **Buff Again**:
   - Calls the buff function again (buffs may have expired during travel).

### Chest Location Phase - Part 1: Find Tree

5. **Move to Infifuss Tree**:
   - Calls a movement function with a position-finding callback.
   - The callback function:
     - Loops through all objects in the current area.
     - For each object, checks if its name is "InifussTree".
     - If the tree is found, returns its position and true.
     - If not found, returns an empty position and false.
   - The movement function navigates to the tree's position.
   - If movement fails, returns the error.
   - Note: The Infifuss Tree is a landmark near Khalim's Chest, making navigation easier.

### Chest Location Phase - Part 2: Find Chest

6. **Move to Khalim's Chest**:
   - Calls a movement function with another position-finding callback.
   - The callback function:
     - Searches for the "KhalimChest3" object (the chest containing the eye).
     - If found:
       - Logs "Khalm Chest found, moving to that room"
       - Returns the chest's position and true.
     - If not found, returns an empty position and false.
   - Navigates to the chest's position.
   - If movement fails, returns the error.

### Combat Phase

7. **Clear Area**:
   - Clears all monsters within 15 units of the player.
   - Uses a filter that accepts any monster type.
   - This ensures safe chest opening.

### Chest Interaction Phase

8. **Relocate Chest**:
   - Searches for the KhalimChest3 object again (to get updated state).
   - If not found, logs "Khalim Chest not found" (debug message).

9. **Open Chest**:
   - Interacts with the chest object.
   - The interaction waits until the chest becomes non-selectable (opened).
   - The callback function:
     - Repeatedly finds the chest object.
     - Checks if the chest is NOT selectable anymore.
     - Returns true when non-selectable (successfully opened).
   - If interaction fails, returns the error.

10. **Success**:
    - Returns nil (the eye is now picked up automatically).

---

## Supporting Function: findKhalimsHeart()

This function navigates to the Kurast Sewers and retrieves Khalim's Heart. The sewers have a special mechanic where you need to find hidden stairs from Level 1 to Level 2. Here's the complete sequence:

### Navigation Phase - Level 1

1. **Waypoint to Kurast Bazaar**:
   - Uses the waypoint system to travel to Kurast Bazaar.
   - If this fails, returns the error.

2. **Buff Before Sewers**:
   - Calls the buff function to apply all active buffs.

3. **Enter Sewers Level 1**:
   - Moves from Kurast Bazaar to Sewers Level 1.
   - If this fails, returns the error.

4. **Buff Again**:
   - Calls the buff function again.

### Finding Hidden Stairs

5. **Locate Stairs to Level 2**:
   - Calls a movement function with a position-finding callback.
   - The callback function:
     - Loops through all adjacent levels (areas you can travel to from current position).
     - For each adjacent level, checks if its area ID is "SewersLevel2Act3".
     - If Sewers Level 2 is found as an adjacent area:
       - Returns that adjacency's position and true.
     - If not found, returns an empty position and false.
   - Navigates to the position where Level 2 can be accessed.
   - If movement fails, returns the error.

6. **Clear Area Around Stairs**:
   - Clears all monsters within 10 units of the player.
   - Uses a filter that accepts any monster type.
   - This ensures safe interaction with the stairs.

### Stairs Interaction

7. **Find Stairs Object**:
   - Searches for the "Act3SewerStairsToLevel3" object.
   - If not found, logs "Khalim Chest not found" (note: this is a typo in the log message—should say "Stairs not found").

8. **Interact with Stairs**:
   - Interacts with the stairs object.
   - The interaction waits until the stairs become non-selectable (used).
   - The callback function:
     - Repeatedly finds the stairs object.
     - Checks if the stairs are NOT selectable anymore.
     - Returns true when non-selectable.
   - If interaction fails, returns the error.

9. **Wait for Level Transition**:
   - Waits 4000 milliseconds (4 seconds).
   - This gives time for the game to load the new level.

### Navigation Phase - Level 2

10. **Move to Sewers Level 2**:
    - Moves to Sewers Level 2 (now that the stairs have been used).
    - If this fails, returns the error.

11. **Buff in Level 2**:
    - Calls the buff function again.

### Chest Location Phase

12. **Move to Khalim's Chest**:
    - Calls a movement function with a position-finding callback.
    - Logs "Khalm Chest found, moving to that room" (note: this log happens BEFORE checking if the chest is found—likely a misplacement).
    - The callback function:
      - Searches for the "KhalimChest1" object (the chest containing the heart).
      - Returns the chest's position and whether it was found.
    - Navigates to the chest's position.
    - If movement fails, returns the error.

### Combat Phase

13. **Clear Area**:
    - Clears all monsters within 15 units of the player.
    - Uses a filter that accepts any monster type.

### Chest Interaction Phase

14. **Relocate Chest**:
    - Searches for the KhalimChest1 object again.
    - If not found, logs "Khalim Chest not found".

15. **Open Chest**:
    - Interacts with the chest object.
    - The interaction waits until the chest becomes non-selectable (opened).
    - The callback function:
      - Repeatedly finds the chest object.
      - Checks if it's NOT selectable anymore.
      - Returns true when non-selectable.
    - If interaction fails, returns the error.

16. **Success**:
    - Returns nil (the heart is now picked up automatically).

---

## Supporting Function: prepareWill()

This function combines the four relics in the Horadric Cube to create Khalim's Will. Here's the detailed logic:

### Check for Complete Will

1. **Search for Khalim's Will**:
   - Searches for "KhalimsWill" in inventory, stash, and equipped slots.
   - If found:
     - The will is already complete, no need to combine parts.
     - Skips all the following steps.

### Component Verification

If Khalim's Will is NOT found, the function checks for each component:

2. **Search for Khalim's Eye**:
   - Searches for "KhalimsEye" in inventory, stash, and equipped slots.
   - If NOT found:
     - Logs "Khalim's Eye not found, skipping"
     - Returns nil (can't combine without this piece).

3. **Search for Khalim's Brain**:
   - Searches for "KhalimsBrain" in inventory, stash, and equipped slots.
   - If NOT found:
     - Logs "Khalim's Brain not found, skipping"
     - Returns nil (can't combine without this piece).

4. **Search for Khalim's Heart**:
   - Searches for "KhalimsHeart" in inventory, stash, and equipped slots.
   - If NOT found:
     - Logs "Khalim's Heart not found, skipping"
     - Returns nil (can't combine without this piece).

5. **Search for Khalim's Flail**:
   - Searches for "KhalimsFlail" in inventory, stash, and equipped slots.
   - If NOT found:
     - Logs "Khalim's Flail not found, skipping"
     - Returns nil (can't combine without this piece).

### Cube Recipe Execution

If all four components are found:

6. **Add Items to Cube**:
   - Calls a function to place all four items into the Horadric Cube.
   - Parameters: eye, brain, heart, flail.
   - If this fails, returns the error.

7. **Transmute**:
   - Calls a function to perform the cube transmutation.
   - This combines all four relics into Khalim's Will.
   - If this fails, returns the error.

8. **Success**:
   - Returns nil (Khalim's Will is now created).

---

## Supporting Function: openMephistoStairs()

This function uses Khalim's Will on the Compelling Orb in Travincal to open the stairs to the Durance of Hate. This is one of the most complex functions due to weapon swapping mechanics. Here's the complete sequence:

### Initial Preparation

1. **Search for Khalim's Will**:
   - Searches for "KhalimsWill" in inventory, stash, and equipped slots.
   - Stores the item and a found boolean.

2. **Combine Relics if Needed**:
   - If Khalim's Will is NOT found:
     - Calls `prepareWill()` to combine the four relics.
     - This ensures the will exists before trying to use it.

### Khalim's Will Equipping Logic

3. **Check Quest Status and Will Location**:
   - Checks if The Guardian quest is NOT completed AND Khalim's Will was found.
   - If both conditions are true:
     - Logs "Khalim's Will found!"
     - **Equipment Check**: Checks if Khalim's Will is NOT equipped.
     - If not equipped:

### Weapon Swap Preparation

4. **Swap to Secondary Weapon Slot**:
   - Checks if the active weapon slot is 0 (primary weapon slot).
   - If yes:
     - Waits 500 milliseconds.
     - Presses the weapon swap keybinding.
     - Waits 500 milliseconds.
   - This switches to the secondary weapon slot where we'll equip Khalim's Will.

### Retrieve from Stash (If Needed)

5. **Stash Retrieval**:
   - Checks if Khalim's Will is in the stash.
   - If yes:
     - Logs "It's in the stash, let's pick it up"
     - Searches for the Bank object (stash).
     - If not found, logs "bank object not found".
     - Waits 300 milliseconds.
     - Interacts with the bank, waiting until the stash menu opens.
     - If interaction fails, returns the error.

### Open Inventory (If Needed)

6. **Inventory Opening**:
   - Checks if Khalim's Will is in the inventory AND the inventory is NOT already open.
   - If yes:
     - Presses the inventory keybinding to open the inventory.

### Equip Khalim's Will

7. **Click to Equip**:
   - Gets the screen coordinates for Khalim's Will.
   - Performs a Shift+Left-Click on the item.
   - Shift-clicking from inventory/stash to the character screen equips the item.
   - Waits 300 milliseconds for the action to complete.

### Weapon Swap Back

8. **Return to Primary Weapon Slot**:
   - Checks if the active weapon slot is now 1 (secondary slot).
   - If yes:
     - Waits 500 milliseconds.
     - Presses the weapon swap keybinding.
     - Waits 500 milliseconds.
   - This returns to the primary weapon slot for normal combat.

9. **Close Menus**:
   - Closes all open menus (inventory, stash).

### Travel to Travincal

10. **Waypoint to Travincal**:
    - Uses the waypoint system to travel to Travincal.
    - If this fails, returns the error.

### Compelling Orb Interaction

11. **Locate Compelling Orb**:
    - Searches for the "CompellingOrb" object.
    - If not found, logs "Compelling Orb not found" (debug message).

12. **Move to Orb**:
    - Moves to the Compelling Orb's position.

### Swap to Khalim's Will

13. **Equip Khalim's Will for Interaction**:
    - Checks if the active weapon slot is 0 (primary).
    - If yes:
      - Waits 500 milliseconds.
      - Presses weapon swap keybinding.
      - Waits 500 milliseconds.
    - Now Khalim's Will is equipped and ready to use on the orb.

### Interact with Orb

14. **Use Khalim's Will on Orb**:
    - Interacts with the Compelling Orb object.
    - The interaction waits until the orb becomes non-selectable (used).
    - The callback function:
      - Repeatedly finds the orb object.
      - Checks if it's NOT selectable anymore.
      - Returns true when non-selectable.
    - If interaction fails, returns the error.

15. **Wait After Interaction**:
    - Waits 300 milliseconds.

### Swap Back to Primary Weapon

16. **Return to Combat Weapon**:
    - Checks if the active weapon slot is 1 (secondary with Khalim's Will).
    - If yes:
      - Waits 500 milliseconds.
      - Presses weapon swap keybinding.
      - Waits 500 milliseconds.
    - Returns to the primary weapon for combat.

### Wait for Stairs Animation

17. **Animation Wait**:
    - Waits 12000 milliseconds (12 seconds).
    - This gives time for the stairs animation to complete.
    - The stairs rise up after using Khalim's Will on the orb.

### Enter Durance of Hate

18. **Blackened Temple Quest Check**:
    - Checks if The Blackened Temple quest is completed.
    - This quest is automatically completed when you use Khalim's Will on the orb.
    - If completed:

19. **Find Stairs**:
    - Searches for the "StairSR" object (stairs to Durance of Hate Level 1).
    - If not found, logs "Stairs to Durance not found".

20. **Interact with Stairs**:
    - Interacts with the stairs object.
    - The interaction waits until the player's area changes to Durance of Hate Level 1.
    - If interaction fails, returns the error.

### Waypoint Discovery

21. **Navigate to Level 2**:
    - Moves from Durance of Hate Level 1 to Level 2.

22. **Discover Waypoint**:
    - Calls a function to discover and activate the waypoint in Durance of Hate Level 2.
    - This is important because you'll need this waypoint for the Mephisto fight.
    - If discovery fails, returns the error.

23. **Success**:
    - Returns nil (stairs are open, waypoint is discovered, ready for Mephisto).

---

## Supporting Function: jadefigurine()

This function handles the Golden Bird quest chain, which rewards a permanent +20 life bonus. Here's the complete sequence:

### Quest Chain Overview

The Golden Bird quest has multiple stages:
1. Find Jade Figurine (drops from a unique monster)
2. Trade Jade Figurine to Meshif for Golden Bird
3. Trade Golden Bird to Alkor for Potion of Life
4. Drink the Potion of Life for permanent +20 life

### Jade Figurine Stage

1. **Search for Jade Figurine**:
   - Searches the inventory for "AJadeFigurine".
   - If found:
     - Interacts with NPC Meshif2 (Meshif at the docks).
     - This trades the Jade Figurine for the Golden Bird.

### Golden Bird Stage

2. **Search for Golden Bird**:
   - Searches the inventory for "TheGoldenBird".
   - If found:
     - **Alkor Interaction**: Interacts with Alkor (the potion seller).
     - **Ormus Interaction**: Interacts with Ormus (the mage/shop keeper).
     - **Alkor Again**: Interacts with Alkor a second time.
     - This sequence of NPC interactions is necessary to complete the quest dialogue chain.
     - After these interactions, Alkor gives you the Potion of Life.

### Potion of Life Stage

3. **Wait**:
   - Waits 500 milliseconds for the quest transactions to complete.

4. **Search for Potion of Life**:
   - Searches the inventory for "PotionOfLife".
   - If found:
     - **Open Inventory**: Presses the inventory keybinding.
     - **Get Potion Position**: Calculates the screen coordinates for the potion.
     - **Right-Click Potion**: Right-clicks the potion to drink it.
     - This permanently increases maximum life by 20 points.
     - **Close Menus**: Closes all open menus.

5. **Success**:
   - Returns nil (quest chain completed).

---

## Summary

The `leveling_act3.go` file implements a comprehensive system for Act 3 progression:

### Quest Orchestration
- **Relic Collection**: Systematically collects all four pieces of Khalim's relics (Eye, Brain, Heart, Flail)
- **Component Tracking**: Individually tracks each relic to avoid duplicate collection
- **Cube Combination**: Automatically combines the four relics into Khalim's Will
- **Compelling Orb Interaction**: Complex weapon swapping logic to use Khalim's Will on the orb

### Navigation Complexity
- **Spider Cavern**: Uses the Infifuss Tree as a landmark to find Khalim's Eye
- **Kurast Sewers**: Handles the hidden stairs mechanic from Level 1 to Level 2
- **Travincal**: Manages Council fight and Compelling Orb interaction
- **Durance of Hate**: Three-level dungeon navigation to reach Mephisto

### Special Mechanics
- **Weapon Swapping**: Sophisticated weapon slot management to equip Khalim's Will temporarily
- **Stash Management**: Retrieves Khalim's Will from stash if needed
- **Animation Timing**: Multiple wait periods for stairs animation, bridge raising, etc.

### Side Quest Management
- **Golden Bird Quest**: Complete three-stage quest chain for permanent life bonus
- **Hratli Movement**: Triggers Hratli to move from pier to normal position
- **Quest Dialogue**: Handles complex NPC interaction sequences

### Gold Farming
- **Lower Kurast Chests**: Efficient super chest farming when gold is low
- **Hell Difficulty**: Additional Mausoleum farming for high-level characters

### Act Transition
- **Multiple Exit Strategies**:
  - Direct waypoint to Pandemonium Fortress (if available)
  - Manual portal route through Durance of Hate Level 3
- **Cinematic Skipping**: Multiple SPACE key holds with varying durations
- **Bridge Mechanic**: Waits for bridge to rise after Mephisto dies

### Error Handling
- **Quest State Validation**: Checks quest completion status at multiple points
- **Component Verification**: Ensures all relics are present before cube combination
- **Portal Existence**: Verifies Hell Gate exists before trying to use it
- **Graceful Exits**: Returns nil with informative logging when conditions aren't met

The system is designed to handle Act 3's complex, non-linear quest structure where multiple relics can be collected in any order, with robust logic to skip already-completed steps and handle various edge cases.
