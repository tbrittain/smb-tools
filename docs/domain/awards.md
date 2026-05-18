# Awards

## Overview

SmbExplorerCompanion tracks 28 built-in awards plus supports custom user-created awards. Awards are assigned at the season level (per `PlayerSeason`).

## Award Categories

Each award has boolean flags for its category:
- `IsBattingAward` — given for offensive performance
- `IsPitchingAward` — given for pitching performance
- `IsFieldingAward` — given for defensive performance
- `IsPlayoffAward` — given for postseason performance

## Assignment Type

- `IsUserAssignable = true` — user manually assigns in the Awards Delegation screen
- `IsUserAssignable = false` — automatically computed from statistics

## Display Priority

`Importance` controls the display order and visual weight of awards:

| Importance | Awards |
|-----------|--------|
| 0 | MVP, Triple Crown (Batting), Triple Crown (Pitching) |
| 1 | Cy Young, Silver Slugger, ROY |
| 2 | Gold Glove, Playoff MVP, Championship MVP |
| 3 | Batting Title, Home Run Title, RBI Title, ERA Title, Wins Title, Strikeouts Title |
| 4 | All-Star |
| 5 | Runner-up versions (MVP-2 through MVP-5, Cy Young-2 through Cy Young-5, ROY-2 through ROY-5) |

`OmitFromGroupings = true` for all runner-up (importance 5) awards — they are not included in award group summaries on the home screen.

## Complete Award List

| Award | Importance | Batting | Pitching | Fielding | Playoff | User Assignable | Auto Computed |
|-------|-----------|---------|----------|----------|---------|-----------------|---------------|
| MVP | 0 | ✓ | ✓ | | | ✓ | |
| Triple Crown (Batting) | 0 | ✓ | | | | | ✓ |
| Triple Crown (Pitching) | 0 | | ✓ | | | | ✓ |
| Cy Young | 1 | | ✓ | | | ✓ | |
| Silver Slugger | 1 | ✓ | | | | ✓ | |
| ROY | 1 | ✓ | ✓ | | | ✓ | |
| Gold Glove | 2 | | | ✓ | | ✓ | |
| Playoff MVP | 2 | ✓ | ✓ | | ✓ | ✓ | |
| Championship MVP | 2 | ✓ | ✓ | | ✓ | ✓ | |
| Batting Title | 3 | ✓ | | | | | ✓ |
| Home Run Title | 3 | ✓ | | | | | ✓ |
| RBI Title | 3 | ✓ | | | | | ✓ |
| ERA Title | 3 | | ✓ | | | | ✓ |
| Wins Title | 3 | | ✓ | | | | ✓ |
| Strikeouts Title | 3 | | ✓ | | | | ✓ |
| All-Star | 4 | ✓ | ✓ | | | ✓ | |
| MVP-2 | 5 | ✓ | ✓ | | | ✓ | |
| MVP-3 | 5 | ✓ | ✓ | | | ✓ | |
| MVP-4 | 5 | ✓ | ✓ | | | ✓ | |
| MVP-5 | 5 | ✓ | ✓ | | | ✓ | |
| Cy Young-2 | 5 | | ✓ | | | ✓ | |
| Cy Young-3 | 5 | | ✓ | | | ✓ | |
| Cy Young-4 | 5 | | ✓ | | | ✓ | |
| Cy Young-5 | 5 | | ✓ | | | ✓ | |
| ROY-2 | 5 | ✓ | ✓ | | | ✓ | |
| ROY-3 | 5 | ✓ | ✓ | | | ✓ | |
| ROY-4 | 5 | ✓ | ✓ | | | ✓ | |
| ROY-5 | 5 | ✓ | ✓ | | | ✓ | |

## Automatic Award Criteria

| Award | Criterion |
|-------|-----------|
| Batting Title | Highest batting average (BA) in the season |
| Home Run Title | Most home runs (HR) |
| RBI Title | Most runs batted in (RBI) |
| ERA Title | Lowest earned run average (ERA) |
| Wins Title | Most pitcher wins (W) |
| Strikeouts Title | Most strikeouts (K) by a pitcher |
| Triple Crown (Batting) | Franchise leader in BA, HR, and RBI simultaneously |
| Triple Crown (Pitching) | Franchise leader in W, ERA, and K simultaneously |

## Hall of Fame

The `IsHallOfFamer` flag is set on the `Player` entity (not `PlayerSeason`). Hall of Fame status is:
- Evaluated when a player retires (they become an eligible candidate)
- Manually assigned by the user in the Hall of Famers screen
- Permanent once assigned — cannot be revoked through the UI

Hall of Famers are highlighted on the home screen and can be filtered in the leaderboard views.

## Custom Awards

Users can create custom awards beyond the built-in list. Custom awards have `IsBuiltIn = false`. They are user-assignable and appear in the Awards Delegation screen. This enables tracking of custom franchise awards (e.g., "Iron Man Award", "Best Defensive Team").
