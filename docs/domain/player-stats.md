# Player Statistics

All statistics are tracked separately for **regular season** and **playoffs**. Career statistics are accumulated across franchise seasons.

## Batting Statistics

### Counting Stats

| Stat | Full Name | Notes |
|------|-----------|-------|
| G | Games Played | Total games the player appeared in |
| GB | Games Batting | Games in which the player actually batted |
| PA | Plate Appearances | |
| AB | At Bats | |
| H | Hits | Total hits |
| 1B | Singles | Derived: H - 2B - 3B - HR |
| 2B | Doubles | |
| 3B | Triples | |
| HR | Home Runs | |
| XBH | Extra Base Hits | 2B + 3B + HR |
| TB | Total Bases | 1B + (2×2B) + (3×3B) + (4×HR) |
| R | Runs Scored | |
| RBI | Runs Batted In | |
| BB | Walks | Base on balls |
| K | Strikeouts | |
| HBP | Hit By Pitch | |
| SB | Stolen Bases | |
| CS | Caught Stealing | |
| SacH | Sacrifice Hits (bunts) | |
| SacF | Sacrifice Flies | |
| E | Errors | Defensive errors committed |
| PB | Passed Balls | Catcher-specific |

### Rate and Advanced Stats

| Stat | Full Name | Formula / Notes |
|------|-----------|-----------------|
| BA | Batting Average | H / AB |
| OBP | On-Base Percentage | (H + BB + HBP) / (AB + BB + HBP + SacF) |
| SLG | Slugging Percentage | TB / AB |
| OPS | On-Base Plus Slugging | OBP + SLG |
| OPS+ | Adjusted OPS | Park/league adjusted OPS (100 = average) |
| wOBA | Weighted On-Base Average | Linear weights formula accounting for value of each offensive event |
| ISO | Isolated Power | SLG - BA (measures raw power) |
| BABIP | Batting Average on Balls in Play | (H - HR) / (AB - K - HR + SacF) |
| K% | Strikeout Rate | K / PA |
| BB% | Walk Rate | BB / PA |
| XBH% | Extra Base Hit Rate | XBH / H |
| PA/G | Plate Appearances per Game | PA / G |
| AB/HR | At Bats per Home Run | AB / HR |

## Pitching Statistics

### Counting Stats

| Stat | Full Name | Notes |
|------|-----------|-------|
| G | Games Pitched | |
| GS | Games Started | |
| GF | Games Finished | |
| CG | Complete Games | |
| CGSO | Complete Game Shutouts | |
| SV | Saves | |
| IP | Innings Pitched | |
| H | Hits Allowed | |
| R | Runs Allowed | |
| ER | Earned Runs | |
| HR | Home Runs Allowed | |
| BB | Walks Issued | |
| K | Strikeouts | |
| HBP | Hit Batters | |
| WP | Wild Pitches | |
| W | Wins | |
| L | Losses | |
| TotalPitches | Total Pitches Thrown | |
| BF | Batters Faced | |

### Rate and Advanced Stats

| Stat | Full Name | Formula / Notes |
|------|-----------|-----------------|
| ERA | Earned Run Average | (ER × 9) / IP |
| ERA- | Adjusted ERA | Park/league adjusted ERA (100 = average, lower = better) |
| WHIP | Walks + Hits per Inning Pitched | (BB + H) / IP |
| FIP | Fielding Independent Pitching | Measures only outcomes the pitcher controls (HR, BB, K); league-constant adjusted |
| FIP- | Adjusted FIP | Park/league adjusted FIP (100 = average, lower = better) |
| K/9 | Strikeouts per 9 Innings | (K × 9) / IP |
| BB/9 | Walks per 9 Innings | (BB × 9) / IP |
| H/9 | Hits per 9 Innings | (H × 9) / IP |
| HR/9 | Home Runs per 9 Innings | (HR × 9) / IP |
| K/BB | Strikeout-to-Walk Ratio | K / BB |
| K% | Strikeout Rate | K / BF |
| Win% | Win Percentage | W / (W + L) |
| OpponentOBP | Opponent On-Base Percentage | |
| P/IP | Pitches per Inning | TotalPitches / IP |
| P/G | Pitches per Game | TotalPitches / GS |

## Custom Metric: smbWAR

`smbWAR` is a custom Wins Above Replacement metric developed for SmbExplorerCompanion. It combines:
- Batting component (weighted offensive production)
- Baserunning component (stolen base value)
- Pitching component (innings-based effectiveness)

Each component uses custom scaling factors calibrated to the SMB run environment. smbWAR is computed in the companion app and is not stored in the SMB save file.

## Playoff vs. Regular Season

All statistics are tracked separately for regular season and playoffs. This is represented in:
- SMB3Explorer: separate CSV export files for each
- SmbExplorerCompanion: `IsRegularSeason` boolean flag on `PlayerSeasonBattingStats` and `PlayerSeasonPitchingStats` entities

## Career vs. Season

- **Season stats**: stats for a single franchise season
- **Career stats**: accumulated totals across all seasons in a franchise

SMB3Explorer exports both. The companion app persists season-level data and computes career aggregates on read.

---

## Source Files

**SMB3Explorer** (`C:\Users\Trey\source\SMB3Explorer`):
- `SMB3Explorer/Resources/Sql/CareerStatsBatting.sql` — all batting stat columns from `t_stats_batting`
- `SMB3Explorer/Resources/Sql/CareerStatsPitching.sql` — all pitching stat columns from `t_stats_pitching`
- `SMB3Explorer/Models/Exports/BattingStatistic.cs` — batting stat export model with derived stat formulas
- `SMB3Explorer/Models/Exports/PitchingStatistic.cs` — pitching stat export model with derived stat formulas

**SmbExplorerCompanion** (`C:\Users\Trey\source\SmbExplorerCompanion`):
- `SmbExplorerCompanion.Database/Entities/PlayerSeasonBattingStat.cs` — all batting stat properties (counting + rate)
- `SmbExplorerCompanion.Database/Entities/PlayerSeasonPitchingStat.cs` — all pitching stat properties (counting + rate)
- `SmbExplorerCompanion.Csv/Models/SeasonStatBatting.cs` — CSV import model for batting stats
- `SmbExplorerCompanion.Csv/Models/SeasonStatPitching.cs` — CSV import model for pitching stats
