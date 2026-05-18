# Player Traits

## Overview

Traits are passive modifiers that affect a player's performance in specific situations. Each player can have **up to 2 traits** at any time.

Traits differ significantly between SMB3 and SMB4:
- **SMB3**: 20 traits, no chemistry system
- **SMB4**: 80+ traits, each tied to a chemistry type; marked as positive or negative

## Save Game Storage

Traits are stored in `t_baseball_player_traits` as a JSON array of objects with `traitId` and `subtypeId` fields. The combination of `(traitId, subtypeId)` maps to a specific trait via separate lookup tables for SMB3 and SMB4.

SMB3 and SMB4 use **different traitId/subtypeId mappings** — the same pair means different things in each game.

## SMB3 Traits (20 total)

SMB3 traits have no chemistry type. The `(traitId, subtypeId)` → trait name mapping:

| traitId | subtypeId | Trait Name | Positive? |
|---------|-----------|-----------|-----------|
| 0 | 0 | POW vs RHP | Yes |
| 0 | 1 | POW vs LHP | Yes |
| 1 | 0 | CON vs RHP | Yes |
| 1 | 1 | CON vs LHP | Yes |
| 2 | 6 | RBI Man | Yes |
| 2 | 7 | RBI Dud | No |
| 3 | 2 | High Pitch | Yes |
| 3 | 3 | Low Pitch | Yes |
| 3 | 4 | Inside Pitch | Yes |
| 3 | 5 | Outside Pitch | Yes |
| 4 | 6 | Tough Out | Yes |
| 4 | 7 | Whiffer | No |
| 5 | null | Specialist | Yes |
| 6 | 6 | Composed | Yes |
| 6 | 7 | BB Prone | No |
| 7 | 6 | K Man | Yes |
| 7 | 7 | K Dud | No |
| 8 | 6 | Stealer | Yes |
| 8 | 7 | Bad Jumps | No |
| 9 | null | Utility | Yes |

## SMB4 Traits (80 total)

SMB4 traits each belong to a chemistry type and are explicitly marked positive or negative.

| Trait Name | Chemistry | Positive? | Category |
|-----------|-----------|-----------|----------|
| Ace Exterminator | Scholarly | Yes | Batting |
| Bad Ball Hitter | Crafty | Yes | Batting |
| Bad Jumps | Crafty | No | Baserunning |
| Base Jogger | Disciplined | No | Baserunning |
| Base Rounder | Disciplined | Yes | Baserunning |
| BB Prone | Disciplined | No | Pitching |
| Big Hack | Scholarly | Yes | Batting |
| Bunter | Scholarly | Yes | Batting |
| Butter Fingers | Disciplined | No | Fielding |
| Cannon Arm | Competitive | Yes | Fielding |
| Choker | Spirited | No | General |
| Clutch | Spirited | Yes | General |
| Composed | Disciplined | Yes | Pitching |
| CON vs LHP | Spirited | Yes | Batting |
| CON vs RHP | Spirited | Yes | Batting |
| Consistent | Disciplined | Yes | General |
| Crossed Up | Scholarly | No | Pitching |
| Distractor | Crafty | Yes | Batting |
| Dive Wizard | Spirited | Yes | Fielding |
| Durable | Competitive | Yes | General |
| Easy Jumps | Crafty | No | Baserunning |
| Easy Target | Crafty | No | Batting |
| Elite 2F | Scholarly | Yes | Pitching |
| Elite 4F | Scholarly | Yes | Pitching |
| Elite CB | Scholarly | Yes | Pitching |
| Elite CF | Scholarly | Yes | Pitching |
| Elite CH | Scholarly | Yes | Pitching |
| Elite FK | Scholarly | Yes | Pitching |
| Elite SB | Scholarly | Yes | Pitching |
| Elite SL | Scholarly | Yes | Pitching |
| Falls Behind | Scholarly | No | Pitching |
| Fastball Hitter | Disciplined | Yes | Batting |
| First Pitch Prayer | Competitive | No | Batting |
| First Pitch Slayer | Competitive | Yes | Batting |
| Gets Ahead | Scholarly | Yes | Pitching |
| High Pitch | Disciplined | Yes | Batting |
| Injury Prone | Competitive | No | General |
| Inside Pitch | Disciplined | Yes | Batting |
| K Collector | Competitive | Yes | Pitching |
| K Neglector | Competitive | No | Pitching |
| Little Hack | Scholarly | Yes | Batting |
| Low Pitch | Disciplined | Yes | Batting |
| Magic Hands | Disciplined | Yes | Fielding |
| Meltdown | Spirited | No | Pitching |
| Metal Head | Disciplined | Yes | Pitching |
| Mind Gamer | Crafty | Yes | Batting |
| Noodle Arm | Competitive | No | Fielding |
| Off-speed Hitter | Disciplined | Yes | Batting |
| Outside Pitch | Disciplined | Yes | Batting |
| Pick Officer | Crafty | Yes | Pitching |
| Pinch Perfect | Disciplined | Yes | Batting |
| POW vs LHP | Spirited | Yes | Batting |
| POW vs RHP | Spirited | Yes | Batting |
| Rally Starter | Spirited | Yes | Batting |
| Rally Stopper | Spirited | Yes | Pitching |
| RBI Hero | Spirited | Yes | Batting |
| RBI Zero | Spirited | No | Batting |
| Reverse Splits | Crafty | Yes | Batting |
| Sign Stealer | Crafty | Yes | Batting |
| Slow Poke | Competitive | No | Baserunning |
| Specialist | Crafty | Yes | Pitching |
| Sprinter | Competitive | Yes | Baserunning |
| Stealer | Crafty | Yes | Baserunning |
| Stimulated | Crafty | Yes | General |
| Surrounded | Spirited | No | Pitching |
| Tough Out | Competitive | Yes | Batting |
| Two Way (C) | Spirited | Yes | General |
| Two Way (IF) | Spirited | Yes | General |
| Two Way (OF) | Spirited | Yes | General |
| Utility | Scholarly | Yes | General |
| Volatile | Disciplined | Yes | General |
| Whiffer | Competitive | No | Batting |
| Wild Thing | Spirited | No | Pitching |
| Wild Thrower | Crafty | No | Fielding |
| Workhorse | Competitive | Yes | Pitching |

## SMB4 traitId/subtypeId Mappings

The `(traitId, subtypeId)` → enum mapping for SMB4 (for reference when reading the save game):

| traitId | subtypeId | Trait |
|---------|-----------|-------|
| 0 | 0 | POW vs RHP |
| 0 | 1 | POW vs LHP |
| 1 | 0 | CON vs RHP |
| 1 | 1 | CON vs LHP |
| 2 | 6 | RBI Hero |
| 2 | 7 | RBI Zero |
| 3 | 2 | High Pitch |
| 3 | 3 | Low Pitch |
| 3 | 4 | Inside Pitch |
| 3 | 5 | Outside Pitch |
| 4 | 6 | Tough Out |
| 4 | 7 | Whiffer |
| 5 | 12 | Specialist |
| 5 | 13 | Reverse Splits |
| 6 | 6 | Composed |
| 6 | 7 | BB Prone |
| 7 | 6 | K Collector |
| 7 | 7 | K Neglector |
| 8 | 6 | Stealer |
| 8 | 7 | Bad Jumps |
| 9 | 6 | Utility |
| 10 | 8 | Fastball Hitter |
| 10 | 9 | Off-speed Hitter |
| 11 | 6 | Bad Ball Hitter |
| 12 | 10 | Big Hack |
| 12 | 11 | Little Hack |
| 13 | 6 | Rally Starter |
| 14 | 6 | First Pitch Slayer |
| 14 | 7 | First Pitch Prayer |
| 15 | 6 | Pinch Perfect |
| 16 | 6 | Ace Exterminator |
| 17 | 6 | Mind Gamer |
| 17 | 7 | Easy Target |
| 18 | 6 | Pick Officer |
| 18 | 7 | Easy Jumps |
| 19 | 6 | Gets Ahead |
| 19 | 7 | Falls Behind |
| 20 | 6 | Rally Stopper |
| 20 | 7 | Surrounded |
| 21 | 7 | Crossed Up |
| 22 | 14 | Elite 4F |
| 22 | 15 | Elite 2F |
| 22 | 16 | Elite CF |
| 22 | 17 | Elite CB |
| 22 | 18 | Elite SL |
| 22 | 19 | Elite CH |
| 22 | 20 | Elite SB |
| 22 | 21 | Elite FK |
| 23 | 6 | Workhorse |
| 24 | 22 | Two Way (OF) |
| 24 | 23 | Two Way (IF) |
| 24 | 24 | Two Way (C) |
| 25 | 6 | Metal Head |
| 26 | 6 | Sprinter |
| 26 | 7 | Slow Poke |
| 27 | 6 | Base Rounder |
| 27 | 7 | Base Jogger |
| 28 | 6 | Distractor |
| 29 | 6 | Magic Hands |
| 29 | 7 | Butter Fingers |
| 30 | 7 | Wild Thrower |
| 31 | 7 | Wild Thing |
| 32 | 6 | Clutch |
| 32 | 7 | Choker |
| 33 | 25 | Consistent |
| 33 | 26 | Volatile |
| 34 | 6 | Durable |
| 34 | 7 | Injury Prone |
| 35 | 6 | Stimulated |
| 36 | 6 | Cannon Arm |
| 36 | 7 | Noodle Arm |
| 37 | 6 | Dive Wizard |
| 38 | 6 | Sign Stealer |
| 39 | 7 | Meltdown |
| 40 | 6 | Bunter |
