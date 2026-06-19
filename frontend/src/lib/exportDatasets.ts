// Compile-time column catalog for the Data Export page.
// Column keys MUST match the keys in internal/store/export_store.go datasetDef.cols maps.
// When a column is added or renamed, update both files.

export type ExportDataType = 'string' | 'int' | 'float' | 'enum'

// Describes how to turn this column's cell into an AppLink in the preview table.
// idKeys name the row keys (set server-side, see export_store.go's linkCols) that
// hold the raw IDs needed to build the route — never rendered as their own column
// and never present in the CSV export query.
export interface ExportColumnLinkDef {
  type: 'player' | 'teamSeason'
  idKeys: { playerId?: string; teamId?: string; teamHistoryId?: string }
}

export interface ExportColumnDef {
  key: string
  label: string
  dataType: ExportDataType
  // Defined for static enum columns (position, hand, chemistry, role).
  // Absent for dynamic enums (team_name, conference_name, division_name) whose options
  // are populated at runtime via the columnOptions map in useExportConfig.
  options?: readonly string[]
  link?: ExportColumnLinkDef
}

export interface ExportDatasetDef {
  id: string
  label: string
  columns: ExportColumnDef[]
  // 'none': no stat-type toggle. 'season': Reg Season / Playoffs toggle.
  // 'career': Reg Season / Playoffs / Total toggle.
  statTypeOptions: 'none' | 'season' | 'career'
  // Whether the "Qualified Players Only" toggle applies to this dataset
  // (batting/pitching season and career datasets only — see export_store.go).
  supportsQualifiedFilter: boolean
}

const battingSeasonColumns: ExportColumnDef[] = [
  {
    key: 'player_name',
    label: 'Player',
    dataType: 'string',
    link: { type: 'player', idKeys: { playerId: '_player_id' } },
  },
  { key: 'first_name', label: 'First Name', dataType: 'string' },
  { key: 'last_name', label: 'Last Name', dataType: 'string' },
  { key: 'season_num', label: 'Season', dataType: 'int' },
  {
    key: 'team_name',
    label: 'Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_team_id', teamHistoryId: '_team_history_id' } },
  },
  { key: 'age', label: 'Age', dataType: 'int' },
  {
    key: 'primary_position',
    label: 'Position',
    dataType: 'enum',
    options: ['P', 'C', '1B', '2B', '3B', 'SS', 'LF', 'CF', 'RF'],
  },
  { key: 'bat_hand', label: 'Bat Hand', dataType: 'enum', options: ['L', 'R', 'S'] },
  { key: 'throw_hand', label: 'Throw Hand', dataType: 'enum', options: ['L', 'R'] },
  {
    key: 'chemistry_type',
    label: 'Chemistry',
    dataType: 'enum',
    options: ['Competitive', 'Spirited', 'Disciplined', 'Scholarly', 'Crafty'],
  },
  { key: 'salary', label: 'Salary', dataType: 'int' },
  { key: 'games_played', label: 'G', dataType: 'int' },
  { key: 'games_batting', label: 'G Bat', dataType: 'int' },
  { key: 'at_bats', label: 'AB', dataType: 'int' },
  { key: 'runs', label: 'R', dataType: 'int' },
  { key: 'hits', label: 'H', dataType: 'int' },
  { key: 'doubles', label: '2B', dataType: 'int' },
  { key: 'triples', label: '3B', dataType: 'int' },
  { key: 'home_runs', label: 'HR', dataType: 'int' },
  { key: 'rbi', label: 'RBI', dataType: 'int' },
  { key: 'stolen_bases', label: 'SB', dataType: 'int' },
  { key: 'caught_stealing', label: 'CS', dataType: 'int' },
  { key: 'walks', label: 'BB', dataType: 'int' },
  { key: 'strikeouts', label: 'K', dataType: 'int' },
  { key: 'hit_by_pitch', label: 'HBP', dataType: 'int' },
  { key: 'sac_hits', label: 'SH', dataType: 'int' },
  { key: 'sac_flies', label: 'SF', dataType: 'int' },
  { key: 'errors', label: 'E', dataType: 'int' },
  { key: 'passed_balls', label: 'PB', dataType: 'int' },
  { key: 'ba', label: 'BA', dataType: 'float' },
  { key: 'obp', label: 'OBP', dataType: 'float' },
  { key: 'slg', label: 'SLG', dataType: 'float' },
  { key: 'ops', label: 'OPS', dataType: 'float' },
  { key: 'iso', label: 'ISO', dataType: 'float' },
  { key: 'babip', label: 'BABIP', dataType: 'float' },
  { key: 'k_pct', label: 'K%', dataType: 'float' },
  { key: 'bb_pct', label: 'BB%', dataType: 'float' },
  { key: 'ab_per_hr', label: 'AB/HR', dataType: 'float' },
  { key: 'ops_plus', label: 'OPS+', dataType: 'float' },
  { key: 'smb_war', label: 'smbWAR', dataType: 'float' },
]

const pitchingSeasonColumns: ExportColumnDef[] = [
  {
    key: 'player_name',
    label: 'Player',
    dataType: 'string',
    link: { type: 'player', idKeys: { playerId: '_player_id' } },
  },
  { key: 'first_name', label: 'First Name', dataType: 'string' },
  { key: 'last_name', label: 'Last Name', dataType: 'string' },
  { key: 'season_num', label: 'Season', dataType: 'int' },
  {
    key: 'team_name',
    label: 'Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_team_id', teamHistoryId: '_team_history_id' } },
  },
  { key: 'age', label: 'Age', dataType: 'int' },
  { key: 'pitcher_role', label: 'Role', dataType: 'enum', options: ['SP', 'RP', 'CL'] },
  { key: 'throw_hand', label: 'Throw Hand', dataType: 'enum', options: ['L', 'R'] },
  {
    key: 'chemistry_type',
    label: 'Chemistry',
    dataType: 'enum',
    options: ['Competitive', 'Spirited', 'Disciplined', 'Scholarly', 'Crafty'],
  },
  { key: 'salary', label: 'Salary', dataType: 'int' },
  { key: 'wins', label: 'W', dataType: 'int' },
  { key: 'losses', label: 'L', dataType: 'int' },
  { key: 'games', label: 'G', dataType: 'int' },
  { key: 'games_started', label: 'GS', dataType: 'int' },
  { key: 'complete_games', label: 'CG', dataType: 'int' },
  { key: 'shutouts', label: 'SHO', dataType: 'int' },
  { key: 'saves', label: 'SV', dataType: 'int' },
  { key: 'outs_pitched', label: 'Outs', dataType: 'int' },
  { key: 'hits_allowed', label: 'H', dataType: 'int' },
  { key: 'earned_runs', label: 'ER', dataType: 'int' },
  { key: 'home_runs_allowed', label: 'HR', dataType: 'int' },
  { key: 'walks', label: 'BB', dataType: 'int' },
  { key: 'strikeouts', label: 'K', dataType: 'int' },
  { key: 'hit_batters', label: 'HBP', dataType: 'int' },
  { key: 'batters_faced', label: 'BF', dataType: 'int' },
  { key: 'games_finished', label: 'GF', dataType: 'int' },
  { key: 'runs_allowed', label: 'R', dataType: 'int' },
  { key: 'wild_pitches', label: 'WP', dataType: 'int' },
  { key: 'total_pitches', label: 'Pitches', dataType: 'int' },
  { key: 'era', label: 'ERA', dataType: 'float' },
  { key: 'whip', label: 'WHIP', dataType: 'float' },
  { key: 'k_per_9', label: 'K/9', dataType: 'float' },
  { key: 'bb_per_9', label: 'BB/9', dataType: 'float' },
  { key: 'h_per_9', label: 'H/9', dataType: 'float' },
  { key: 'hr_per_9', label: 'HR/9', dataType: 'float' },
  { key: 'k_per_bb', label: 'K/BB', dataType: 'float' },
  { key: 'k_pct', label: 'K%', dataType: 'float' },
  { key: 'win_pct', label: 'W%', dataType: 'float' },
  { key: 'p_per_ip', label: 'P/IP', dataType: 'float' },
  { key: 'era_plus', label: 'ERA+', dataType: 'float' },
  { key: 'fip', label: 'FIP', dataType: 'float' },
  { key: 'fip_minus', label: 'FIP-', dataType: 'float' },
  { key: 'smb_war', label: 'smbWAR', dataType: 'float' },
]

const standingsColumns: ExportColumnDef[] = [
  {
    key: 'team_name',
    label: 'Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_team_id', teamHistoryId: '_team_history_id' } },
  },
  { key: 'season_num', label: 'Season', dataType: 'int' },
  { key: 'conference_name', label: 'Conference', dataType: 'enum' },
  { key: 'division_name', label: 'Division', dataType: 'enum' },
  { key: 'wins', label: 'W', dataType: 'int' },
  { key: 'losses', label: 'L', dataType: 'int' },
  { key: 'win_pct', label: 'W%', dataType: 'float' },
  { key: 'games_back', label: 'GB', dataType: 'float' },
  { key: 'runs_for', label: 'RF', dataType: 'int' },
  { key: 'runs_against', label: 'RA', dataType: 'int' },
  { key: 'run_diff', label: 'RD', dataType: 'int' },
  { key: 'playoff_seed', label: 'Playoff Seed', dataType: 'int' },
  { key: 'playoff_wins', label: 'PO W', dataType: 'int' },
  { key: 'playoff_losses', label: 'PO L', dataType: 'int' },
  { key: 'playoff_runs_for', label: 'PO RF', dataType: 'int' },
  { key: 'playoff_runs_against', label: 'PO RA', dataType: 'int' },
  { key: 'budget', label: 'Budget', dataType: 'int' },
  { key: 'payroll', label: 'Payroll', dataType: 'int' },
]

const careerBattingColumns: ExportColumnDef[] = [
  {
    key: 'player_name',
    label: 'Player',
    dataType: 'string',
    link: { type: 'player', idKeys: { playerId: '_player_id' } },
  },
  { key: 'first_name', label: 'First Name', dataType: 'string' },
  { key: 'last_name', label: 'Last Name', dataType: 'string' },
  { key: 'seasons_played', label: 'Seasons', dataType: 'int' },
  { key: 'games_played', label: 'G', dataType: 'int' },
  { key: 'games_batting', label: 'G Bat', dataType: 'int' },
  { key: 'at_bats', label: 'AB', dataType: 'int' },
  { key: 'runs', label: 'R', dataType: 'int' },
  { key: 'hits', label: 'H', dataType: 'int' },
  { key: 'doubles', label: '2B', dataType: 'int' },
  { key: 'triples', label: '3B', dataType: 'int' },
  { key: 'home_runs', label: 'HR', dataType: 'int' },
  { key: 'rbi', label: 'RBI', dataType: 'int' },
  { key: 'stolen_bases', label: 'SB', dataType: 'int' },
  { key: 'caught_stealing', label: 'CS', dataType: 'int' },
  { key: 'walks', label: 'BB', dataType: 'int' },
  { key: 'strikeouts', label: 'K', dataType: 'int' },
  { key: 'hit_by_pitch', label: 'HBP', dataType: 'int' },
  { key: 'sac_hits', label: 'SH', dataType: 'int' },
  { key: 'sac_flies', label: 'SF', dataType: 'int' },
  { key: 'errors', label: 'E', dataType: 'int' },
  { key: 'passed_balls', label: 'PB', dataType: 'int' },
  { key: 'ba', label: 'BA', dataType: 'float' },
  { key: 'obp', label: 'OBP', dataType: 'float' },
  { key: 'slg', label: 'SLG', dataType: 'float' },
  { key: 'ops', label: 'OPS', dataType: 'float' },
  { key: 'iso', label: 'ISO', dataType: 'float' },
  { key: 'babip', label: 'BABIP', dataType: 'float' },
  { key: 'k_pct', label: 'K%', dataType: 'float' },
  { key: 'bb_pct', label: 'BB%', dataType: 'float' },
  { key: 'ab_per_hr', label: 'AB/HR', dataType: 'float' },
  { key: 'ops_plus', label: 'OPS+', dataType: 'float' },
  { key: 'smb_war', label: 'smbWAR', dataType: 'float' },
]

const careerPitchingColumns: ExportColumnDef[] = [
  {
    key: 'player_name',
    label: 'Player',
    dataType: 'string',
    link: { type: 'player', idKeys: { playerId: '_player_id' } },
  },
  { key: 'first_name', label: 'First Name', dataType: 'string' },
  { key: 'last_name', label: 'Last Name', dataType: 'string' },
  { key: 'seasons_played', label: 'Seasons', dataType: 'int' },
  { key: 'wins', label: 'W', dataType: 'int' },
  { key: 'losses', label: 'L', dataType: 'int' },
  { key: 'games', label: 'G', dataType: 'int' },
  { key: 'games_started', label: 'GS', dataType: 'int' },
  { key: 'complete_games', label: 'CG', dataType: 'int' },
  { key: 'shutouts', label: 'SHO', dataType: 'int' },
  { key: 'saves', label: 'SV', dataType: 'int' },
  { key: 'outs_pitched', label: 'Outs', dataType: 'int' },
  { key: 'hits_allowed', label: 'H', dataType: 'int' },
  { key: 'earned_runs', label: 'ER', dataType: 'int' },
  { key: 'home_runs_allowed', label: 'HR', dataType: 'int' },
  { key: 'walks', label: 'BB', dataType: 'int' },
  { key: 'strikeouts', label: 'K', dataType: 'int' },
  { key: 'hit_batters', label: 'HBP', dataType: 'int' },
  { key: 'batters_faced', label: 'BF', dataType: 'int' },
  { key: 'games_finished', label: 'GF', dataType: 'int' },
  { key: 'runs_allowed', label: 'R', dataType: 'int' },
  { key: 'wild_pitches', label: 'WP', dataType: 'int' },
  { key: 'total_pitches', label: 'Pitches', dataType: 'int' },
  { key: 'era', label: 'ERA', dataType: 'float' },
  { key: 'whip', label: 'WHIP', dataType: 'float' },
  { key: 'k_per_9', label: 'K/9', dataType: 'float' },
  { key: 'bb_per_9', label: 'BB/9', dataType: 'float' },
  { key: 'h_per_9', label: 'H/9', dataType: 'float' },
  { key: 'hr_per_9', label: 'HR/9', dataType: 'float' },
  { key: 'k_per_bb', label: 'K/BB', dataType: 'float' },
  { key: 'k_pct', label: 'K%', dataType: 'float' },
  { key: 'win_pct', label: 'W%', dataType: 'float' },
  { key: 'p_per_ip', label: 'P/IP', dataType: 'float' },
  { key: 'era_plus', label: 'ERA+', dataType: 'float' },
  { key: 'fip', label: 'FIP', dataType: 'float' },
  { key: 'fip_minus', label: 'FIP-', dataType: 'float' },
  { key: 'smb_war', label: 'smbWAR', dataType: 'float' },
]

const playerSeasonAttributesColumns: ExportColumnDef[] = [
  {
    key: 'player_name',
    label: 'Player',
    dataType: 'string',
    link: { type: 'player', idKeys: { playerId: '_player_id' } },
  },
  { key: 'first_name', label: 'First Name', dataType: 'string' },
  { key: 'last_name', label: 'Last Name', dataType: 'string' },
  { key: 'season_num', label: 'Season', dataType: 'int' },
  {
    key: 'team_name',
    label: 'Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_team_id', teamHistoryId: '_team_history_id' } },
  },
  { key: 'age', label: 'Age', dataType: 'int' },
  {
    key: 'primary_position',
    label: 'Position',
    dataType: 'enum',
    options: ['P', 'C', '1B', '2B', '3B', 'SS', 'LF', 'CF', 'RF'],
  },
  { key: 'pitcher_role', label: 'Role', dataType: 'enum', options: ['SP', 'RP', 'CL'] },
  { key: 'bat_hand', label: 'Bat Hand', dataType: 'enum', options: ['L', 'R', 'S'] },
  { key: 'throw_hand', label: 'Throw Hand', dataType: 'enum', options: ['L', 'R'] },
  {
    key: 'chemistry_type',
    label: 'Chemistry',
    dataType: 'enum',
    options: ['Competitive', 'Spirited', 'Disciplined', 'Scholarly', 'Crafty'],
  },
  { key: 'salary', label: 'Salary', dataType: 'int' },
  { key: 'power', label: 'Power', dataType: 'int' },
  { key: 'contact', label: 'Contact', dataType: 'int' },
  { key: 'speed', label: 'Speed', dataType: 'int' },
  { key: 'fielding', label: 'Fielding', dataType: 'int' },
  { key: 'arm', label: 'Arm', dataType: 'int' },
  { key: 'velocity', label: 'Velocity', dataType: 'int' },
  { key: 'junk', label: 'Junk', dataType: 'int' },
  { key: 'accuracy', label: 'Accuracy', dataType: 'int' },
]

const awardWinnersColumns: ExportColumnDef[] = [
  {
    key: 'player_name',
    label: 'Player',
    dataType: 'string',
    link: { type: 'player', idKeys: { playerId: '_player_id' } },
  },
  { key: 'first_name', label: 'First Name', dataType: 'string' },
  { key: 'last_name', label: 'Last Name', dataType: 'string' },
  { key: 'season_num', label: 'Season', dataType: 'int' },
  {
    key: 'team_name',
    label: 'Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_team_id', teamHistoryId: '_team_history_id' } },
  },
  { key: 'award_name', label: 'Award', dataType: 'string' },
  { key: 'award_original_name', label: 'Award (Original)', dataType: 'string' },
  { key: 'award_type', label: 'Type', dataType: 'enum', options: ['Winner', 'Runner-Up'] },
]

const regularSeasonScheduleColumns: ExportColumnDef[] = [
  { key: 'season_num', label: 'Season', dataType: 'int' },
  { key: 'game_number', label: 'Game #', dataType: 'int' },
  { key: 'day', label: 'Day', dataType: 'int' },
  {
    key: 'home_team_name',
    label: 'Home Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_home_team_id', teamHistoryId: '_home_team_history_id' } },
  },
  {
    key: 'away_team_name',
    label: 'Away Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_away_team_id', teamHistoryId: '_away_team_history_id' } },
  },
  { key: 'home_score', label: 'Home Score', dataType: 'int' },
  { key: 'away_score', label: 'Away Score', dataType: 'int' },
]

const playoffScheduleColumns: ExportColumnDef[] = [
  { key: 'season_num', label: 'Season', dataType: 'int' },
  { key: 'series_number', label: 'Series #', dataType: 'int' },
  { key: 'game_number', label: 'Game #', dataType: 'int' },
  {
    key: 'home_team_name',
    label: 'Home Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_home_team_id', teamHistoryId: '_home_team_history_id' } },
  },
  {
    key: 'away_team_name',
    label: 'Away Team',
    dataType: 'enum',
    link: { type: 'teamSeason', idKeys: { teamId: '_away_team_id', teamHistoryId: '_away_team_history_id' } },
  },
  { key: 'home_score', label: 'Home Score', dataType: 'int' },
  { key: 'away_score', label: 'Away Score', dataType: 'int' },
]

export const EXPORT_DATASETS: ExportDatasetDef[] = [
  {
    id: 'batting_season',
    label: 'Player Season Batting',
    columns: battingSeasonColumns,
    statTypeOptions: 'season',
    supportsQualifiedFilter: true,
  },
  {
    id: 'pitching_season',
    label: 'Player Season Pitching',
    columns: pitchingSeasonColumns,
    statTypeOptions: 'season',
    supportsQualifiedFilter: true,
  },
  {
    id: 'standings',
    label: 'Team Season Standings',
    columns: standingsColumns,
    statTypeOptions: 'none',
    supportsQualifiedFilter: false,
  },
  {
    id: 'career_batting',
    label: 'Career Batting Stats',
    columns: careerBattingColumns,
    statTypeOptions: 'career',
    supportsQualifiedFilter: true,
  },
  {
    id: 'career_pitching',
    label: 'Career Pitching Stats',
    columns: careerPitchingColumns,
    statTypeOptions: 'career',
    supportsQualifiedFilter: true,
  },
  {
    id: 'player_season_attributes',
    label: 'Player Season Attributes',
    columns: playerSeasonAttributesColumns,
    statTypeOptions: 'none',
    supportsQualifiedFilter: false,
  },
  {
    id: 'award_winners',
    label: 'Season Award Winners',
    columns: awardWinnersColumns,
    statTypeOptions: 'none',
    supportsQualifiedFilter: false,
  },
  {
    id: 'regular_season_schedule',
    label: 'Regular Season Schedule',
    columns: regularSeasonScheduleColumns,
    statTypeOptions: 'none',
    supportsQualifiedFilter: false,
  },
  {
    id: 'playoff_schedule',
    label: 'Playoff Schedule',
    columns: playoffScheduleColumns,
    statTypeOptions: 'none',
    supportsQualifiedFilter: false,
  },
]

export const EXPORT_DATASET_MAP: Record<string, ExportDatasetDef> = Object.fromEntries(
  EXPORT_DATASETS.map((d) => [d.id, d]),
)
