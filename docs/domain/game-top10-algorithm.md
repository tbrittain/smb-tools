# SMB4 Top 10 Ranking Algorithm

## Status and scope

The ranking algorithm on Super Mega Baseball 4's **League Leaders > Top 10**
page has been independently reproduced and verified against observed in-game
leaderboards and their corresponding raw save-game statistics.

The same scoring rules apply to regular-season, playoff, single-season, and
career scopes. Only the source statistics and selected scope change.

## Player classification

The game uses exactly one scoring formula for each player. Players with a
pitcher role use the pitching formula; all other players use the batting
formula. The two scores are never combined, so a pitcher's batting statistics
do not contribute to their Top 10 ranking.

There is no minimum plate-appearance or innings-pitched qualification in the
Top 10 calculation. It calculates a score for every player in the selected scope,
sorts descending, and returns ten rows.

## Batting score

```text
battingScore =
    0.9 * singles
  + 1.2 * doubles
  + 1.6 * triples
  + 2.0 * homeRuns
  + 0.7 * walks
  + 0.7 * hitByPitch
  + 0.5 * RBI
  + 0.2 * stolenBases

singles = hits - doubles - triples - homeRuns
```

Each event has the following effective value:

| Event | Score |
|---|---:|
| Single | `0.9` |
| Double | `1.2` |
| Triple | `1.6` |
| Home run | `2.0` |
| Walk | `0.7` |
| Hit by pitch | `0.7` |
| RBI | `0.5` |
| Stolen base | `0.2` |

Runs, strikeouts, caught stealing, sacrifice hits, sacrifice flies, errors,
passed balls, games, at-bats, and all batting rate statistics are not part of
the score. In particular, AVG, OBP, SLG, and OPS are display values only.

## Pitching score

```text
pitchingScore = 1.75 * (
    0.3 * outsPitched
  + 0.3 * strikeouts
  + shutouts
  + saves
  - earnedRuns
  - 0.3 * homeRunsAllowed
  - 0.3 * walksAllowed
  - 0.3 * battersHitByPitch
)
```

After applying the `1.75` pitcher-to-batter scale factor, the effective values
are:

| Event | Score |
|---|---:|
| Out pitched | `+0.525` |
| Strikeout | `+0.525` |
| Shutout | `+1.75` |
| Save | `+1.75` |
| Earned run | `-1.75` |
| Home run allowed | `-0.525` |
| Walk allowed | `-0.525` |
| Batter hit by pitch | `-0.525` |

Wins, losses, hits allowed, runs allowed, games, games started, complete games,
games finished, wild pitches, batters faced, total pitches, and all pitching
rate statistics are not part of the score. ERA, WHIP, opponent AVG, and the
other rates shown by the UI are display values only.

The `1.75` multiplier does not affect the order among pitchers, but it is
essential when pitchers and position players are merged into one list.

No secondary tie-breaker has been identified. The relative order of players
with exactly equal scores should therefore be treated as unspecified.

## Verification

The documented expression was evaluated against raw save-game statistics from
both an in-progress season and a completed season. In both cases it reproduced
all ten players in their exact in-game order. It also placed the first player
outside the current-season list immediately below the observed cutoff.
