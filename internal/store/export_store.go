package store

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
)

// exportColumn defines one column available for export from a dataset.
// sqlExpr is the SQL expression (with table alias) used in SELECT and ORDER BY.
// label is the CSV header written for this column.
// dataType is "string", "int", or "float" — used by the frontend catalog to
// format cell values (the TS catalog mirrors this; keys must stay in sync).
type exportColumn struct {
	sqlExpr  string
	label    string
	dataType string
}

// datasetDef describes one named export dataset: its display label, the SQL
// FROM clause (including all JOINs and any fixed WHERE conditions), and the
// full map of exportable columns keyed by the column key used in ExportOptionsDTO.
type datasetDef struct {
	id      string
	label   string
	fromSQL string            // everything after SELECT ... (FROM + JOINs + fixed WHERE conditions)
	cols    map[string]exportColumn
}

// validFilterOps is the set of filter operators accepted in FilterRow.Op.
var validFilterOps = map[string]string{
	"eq":       "=",
	"neq":      "!=",
	"lt":       "<",
	"lte":      "<=",
	"gt":       ">",
	"gte":      ">=",
	"contains": "LIKE",
}

// ── Dataset definitions ───────────────────────────────────────────────────────

var battingSeasonDataset = datasetDef{
	id:    "batting_season",
	label: "Player Season Batting",
	fromSQL: `FROM player_season_batting_stats bs
JOIN player_seasons ps ON ps.id = bs.player_season_id
JOIN players p         ON p.id  = ps.player_id
JOIN seasons s         ON s.id  = ps.season_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE bs.is_regular_season = 1`,
	cols: map[string]exportColumn{
		"player_name":     {`p.first_name || ' ' || p.last_name`, "Player", "string"},
		"first_name":      {`p.first_name`, "First Name", "string"},
		"last_name":       {`p.last_name`, "Last Name", "string"},
		"season_num":      {`s.season_num`, "Season", "int"},
		"team_name":       {`COALESCE(tsh.team_name, '')`, "Team", "string"},
		"age":             {`ps.age`, "Age", "int"},
		"primary_position": {`ps.primary_position`, "Position", "string"},
		"bat_hand":        {`ps.bat_hand`, "Bat Hand", "string"},
		"throw_hand":      {`ps.throw_hand`, "Throw Hand", "string"},
		"chemistry_type":  {`ps.chemistry_type`, "Chemistry", "string"},
		"salary":          {`ps.salary`, "Salary", "int"},
		"games_played":    {`bs.games_played`, "G", "int"},
		"games_batting":   {`bs.games_batting`, "G Bat", "int"},
		"at_bats":         {`bs.at_bats`, "AB", "int"},
		"runs":            {`bs.runs`, "R", "int"},
		"hits":            {`bs.hits`, "H", "int"},
		"doubles":         {`bs.doubles`, "2B", "int"},
		"triples":         {`bs.triples`, "3B", "int"},
		"home_runs":       {`bs.home_runs`, "HR", "int"},
		"rbi":             {`bs.rbi`, "RBI", "int"},
		"stolen_bases":    {`bs.stolen_bases`, "SB", "int"},
		"caught_stealing": {`bs.caught_stealing`, "CS", "int"},
		"walks":           {`bs.walks`, "BB", "int"},
		"strikeouts":      {`bs.strikeouts`, "K", "int"},
		"hit_by_pitch":    {`bs.hit_by_pitch`, "HBP", "int"},
		"sac_hits":        {`bs.sac_hits`, "SH", "int"},
		"sac_flies":       {`bs.sac_flies`, "SF", "int"},
		"errors":          {`bs.errors`, "E", "int"},
		"passed_balls":    {`bs.passed_balls`, "PB", "int"},
		"ba":              {`bs.ba`, "BA", "float"},
		"obp":             {`bs.obp`, "OBP", "float"},
		"slg":             {`bs.slg`, "SLG", "float"},
		"ops":             {`bs.ops`, "OPS", "float"},
		"iso":             {`bs.iso`, "ISO", "float"},
		"babip":           {`bs.babip`, "BABIP", "float"},
		"k_pct":           {`bs.k_pct`, "K%", "float"},
		"bb_pct":          {`bs.bb_pct`, "BB%", "float"},
		"ab_per_hr":       {`bs.ab_per_hr`, "AB/HR", "float"},
		"ops_plus":        {`bs.ops_plus`, "OPS+", "float"},
		"smb_war":         {`bs.smb_war`, "smbWAR", "float"},
	},
}

var pitchingSeasonDataset = datasetDef{
	id:    "pitching_season",
	label: "Player Season Pitching",
	fromSQL: `FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN players p         ON p.id  = ps.player_id
JOIN seasons s         ON s.id  = ps.season_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE pit.is_regular_season = 1`,
	cols: map[string]exportColumn{
		"player_name":     {`p.first_name || ' ' || p.last_name`, "Player", "string"},
		"first_name":      {`p.first_name`, "First Name", "string"},
		"last_name":       {`p.last_name`, "Last Name", "string"},
		"season_num":      {`s.season_num`, "Season", "int"},
		"team_name":       {`COALESCE(tsh.team_name, '')`, "Team", "string"},
		"age":             {`ps.age`, "Age", "int"},
		"pitcher_role":    {`ps.pitcher_role`, "Role", "string"},
		"throw_hand":      {`ps.throw_hand`, "Throw Hand", "string"},
		"chemistry_type":  {`ps.chemistry_type`, "Chemistry", "string"},
		"salary":          {`ps.salary`, "Salary", "int"},
		"wins":            {`pit.wins`, "W", "int"},
		"losses":          {`pit.losses`, "L", "int"},
		"games":           {`pit.games`, "G", "int"},
		"games_started":   {`pit.games_started`, "GS", "int"},
		"complete_games":  {`pit.complete_games`, "CG", "int"},
		"shutouts":        {`pit.shutouts`, "SHO", "int"},
		"saves":           {`pit.saves`, "SV", "int"},
		"outs_pitched":    {`pit.outs_pitched`, "Outs", "int"},
		"hits_allowed":    {`pit.hits_allowed`, "H", "int"},
		"earned_runs":     {`pit.earned_runs`, "ER", "int"},
		"home_runs_allowed": {`pit.home_runs_allowed`, "HR", "int"},
		"walks":           {`pit.walks`, "BB", "int"},
		"strikeouts":      {`pit.strikeouts`, "K", "int"},
		"hit_batters":     {`pit.hit_batters`, "HBP", "int"},
		"batters_faced":   {`pit.batters_faced`, "BF", "int"},
		"games_finished":  {`pit.games_finished`, "GF", "int"},
		"runs_allowed":    {`pit.runs_allowed`, "R", "int"},
		"wild_pitches":    {`pit.wild_pitches`, "WP", "int"},
		"total_pitches":   {`pit.total_pitches`, "Pitches", "int"},
		"era":             {`pit.era`, "ERA", "float"},
		"whip":            {`pit.whip`, "WHIP", "float"},
		"k_per_9":         {`pit.k_per_9`, "K/9", "float"},
		"bb_per_9":        {`pit.bb_per_9`, "BB/9", "float"},
		"h_per_9":         {`pit.h_per_9`, "H/9", "float"},
		"hr_per_9":        {`pit.hr_per_9`, "HR/9", "float"},
		"k_per_bb":        {`pit.k_per_bb`, "K/BB", "float"},
		"k_pct":           {`pit.k_pct`, "K%", "float"},
		"win_pct":         {`pit.win_pct`, "W%", "float"},
		"p_per_ip":        {`pit.p_per_ip`, "P/IP", "float"},
		"era_plus":        {`pit.era_plus`, "ERA+", "float"},
		"fip":             {`pit.fip`, "FIP", "float"},
		"fip_minus":       {`pit.fip_minus`, "FIP-", "float"},
		"smb_war":         {`pit.smb_war`, "smbWAR", "float"},
	},
}

var standingsDataset = datasetDef{
	id:    "standings",
	label: "Team Season Standings",
	fromSQL: `FROM team_season_history tsh
JOIN seasons s ON s.id = tsh.season_id`,
	cols: map[string]exportColumn{
		"team_name":           {`tsh.team_name`, "Team", "string"},
		"season_num":          {`s.season_num`, "Season", "int"},
		"conference_name":     {`tsh.conference_name`, "Conference", "string"},
		"division_name":       {`tsh.division_name`, "Division", "string"},
		"wins":                {`tsh.wins`, "W", "int"},
		"losses":              {`tsh.losses`, "L", "int"},
		"win_pct":             {`CAST(tsh.wins AS REAL) / NULLIF(tsh.wins + tsh.losses, 0)`, "W%", "float"},
		"games_back":          {`tsh.games_back`, "GB", "float"},
		"runs_for":            {`tsh.runs_for`, "RF", "int"},
		"runs_against":        {`tsh.runs_against`, "RA", "int"},
		"run_diff":            {`tsh.runs_for - tsh.runs_against`, "RD", "int"},
		"playoff_seed":        {`tsh.playoff_seed`, "Playoff Seed", "int"},
		"playoff_wins":        {`tsh.playoff_wins`, "PO W", "int"},
		"playoff_losses":      {`tsh.playoff_losses`, "PO L", "int"},
		"playoff_runs_for":    {`tsh.playoff_runs_for`, "PO RF", "int"},
		"playoff_runs_against": {`tsh.playoff_runs_against`, "PO RA", "int"},
		"budget":              {`tsh.budget`, "Budget", "int"},
		"payroll":             {`tsh.payroll`, "Payroll", "int"},
	},
}

var careerBattingDataset = datasetDef{
	id:    "career_batting",
	label: "Career Batting Stats",
	fromSQL: `FROM player_career_batting_stats cbs
JOIN players p ON p.id = cbs.player_id`,
	cols: map[string]exportColumn{
		"player_name":     {`p.first_name || ' ' || p.last_name`, "Player", "string"},
		"first_name":      {`p.first_name`, "First Name", "string"},
		"last_name":       {`p.last_name`, "Last Name", "string"},
		"seasons_played":  {`cbs.seasons_played`, "Seasons", "int"},
		"games_played":    {`cbs.games_played`, "G", "int"},
		"games_batting":   {`cbs.games_batting`, "G Bat", "int"},
		"at_bats":         {`cbs.at_bats`, "AB", "int"},
		"runs":            {`cbs.runs`, "R", "int"},
		"hits":            {`cbs.hits`, "H", "int"},
		"doubles":         {`cbs.doubles`, "2B", "int"},
		"triples":         {`cbs.triples`, "3B", "int"},
		"home_runs":       {`cbs.home_runs`, "HR", "int"},
		"rbi":             {`cbs.rbi`, "RBI", "int"},
		"stolen_bases":    {`cbs.stolen_bases`, "SB", "int"},
		"caught_stealing": {`cbs.caught_stealing`, "CS", "int"},
		"walks":           {`cbs.walks`, "BB", "int"},
		"strikeouts":      {`cbs.strikeouts`, "K", "int"},
		"hit_by_pitch":    {`cbs.hit_by_pitch`, "HBP", "int"},
		"sac_hits":        {`cbs.sac_hits`, "SH", "int"},
		"sac_flies":       {`cbs.sac_flies`, "SF", "int"},
		"errors":          {`cbs.errors`, "E", "int"},
		"passed_balls":    {`cbs.passed_balls`, "PB", "int"},
		"ba":              {`cbs.ba`, "BA", "float"},
		"obp":             {`cbs.obp`, "OBP", "float"},
		"slg":             {`cbs.slg`, "SLG", "float"},
		"ops":             {`cbs.ops`, "OPS", "float"},
		"iso":             {`cbs.iso`, "ISO", "float"},
		"babip":           {`cbs.babip`, "BABIP", "float"},
		"k_pct":           {`cbs.k_pct`, "K%", "float"},
		"bb_pct":          {`cbs.bb_pct`, "BB%", "float"},
		"ab_per_hr":       {`cbs.ab_per_hr`, "AB/HR", "float"},
		"ops_plus":        {`cbs.ops_plus`, "OPS+", "float"},
		"smb_war":         {`cbs.smb_war`, "smbWAR", "float"},
	},
}

var careerPitchingDataset = datasetDef{
	id:    "career_pitching",
	label: "Career Pitching Stats",
	fromSQL: `FROM player_career_pitching_stats cps
JOIN players p ON p.id = cps.player_id`,
	cols: map[string]exportColumn{
		"player_name":       {`p.first_name || ' ' || p.last_name`, "Player", "string"},
		"first_name":        {`p.first_name`, "First Name", "string"},
		"last_name":         {`p.last_name`, "Last Name", "string"},
		"seasons_played":    {`cps.seasons_played`, "Seasons", "int"},
		"wins":              {`cps.wins`, "W", "int"},
		"losses":            {`cps.losses`, "L", "int"},
		"games":             {`cps.games`, "G", "int"},
		"games_started":     {`cps.games_started`, "GS", "int"},
		"complete_games":    {`cps.complete_games`, "CG", "int"},
		"shutouts":          {`cps.shutouts`, "SHO", "int"},
		"saves":             {`cps.saves`, "SV", "int"},
		"outs_pitched":      {`cps.outs_pitched`, "Outs", "int"},
		"hits_allowed":      {`cps.hits_allowed`, "H", "int"},
		"earned_runs":       {`cps.earned_runs`, "ER", "int"},
		"home_runs_allowed": {`cps.home_runs_allowed`, "HR", "int"},
		"walks":             {`cps.walks`, "BB", "int"},
		"strikeouts":        {`cps.strikeouts`, "K", "int"},
		"hit_batters":       {`cps.hit_batters`, "HBP", "int"},
		"batters_faced":     {`cps.batters_faced`, "BF", "int"},
		"games_finished":    {`cps.games_finished`, "GF", "int"},
		"runs_allowed":      {`cps.runs_allowed`, "R", "int"},
		"wild_pitches":      {`cps.wild_pitches`, "WP", "int"},
		"total_pitches":     {`cps.total_pitches`, "Pitches", "int"},
		"era":               {`cps.era`, "ERA", "float"},
		"whip":              {`cps.whip`, "WHIP", "float"},
		"k_per_9":           {`cps.k_per_9`, "K/9", "float"},
		"bb_per_9":          {`cps.bb_per_9`, "BB/9", "float"},
		"h_per_9":           {`cps.h_per_9`, "H/9", "float"},
		"hr_per_9":          {`cps.hr_per_9`, "HR/9", "float"},
		"k_per_bb":          {`cps.k_per_bb`, "K/BB", "float"},
		"k_pct":             {`cps.k_pct`, "K%", "float"},
		"win_pct":           {`cps.win_pct`, "W%", "float"},
		"p_per_ip":          {`cps.p_per_ip`, "P/IP", "float"},
		"era_plus":          {`cps.era_plus`, "ERA+", "float"},
		"fip":               {`cps.fip`, "FIP", "float"},
		"fip_minus":         {`cps.fip_minus`, "FIP-", "float"},
		"smb_war":           {`cps.smb_war`, "smbWAR", "float"},
	},
}

// allDatasets is the ordered list of datasets returned by GetExportDatasets.
// The TypeScript constant in frontend/src/lib/exportDatasets.ts must match
// these dataset IDs and column keys.
var allDatasets = []datasetDef{
	battingSeasonDataset,
	pitchingSeasonDataset,
	standingsDataset,
	careerBattingDataset,
	careerPitchingDataset,
}

// datasetByID maps dataset ID strings to their definitions for fast lookup.
var datasetByID = func() map[string]datasetDef {
	m := make(map[string]datasetDef, len(allDatasets))
	for _, d := range allDatasets {
		m[d.id] = d
	}
	return m
}()

// ── ExportStore ───────────────────────────────────────────────────────────────

// ExportStore executes parameterized export queries against the companion DB.
type ExportStore struct {
	db DBTX
}

func NewExportStore(db DBTX) *ExportStore {
	return &ExportStore{db: db}
}

// ExportOptions is the domain-layer equivalent of ExportOptionsDTO.
type ExportOptions struct {
	DatasetID      string
	Columns        []string
	Filters        []FilterRow
	SortCol        string
	SortDir        string
	CareerStatType string
}

// FilterRow is one filter condition from the frontend.
type FilterRow struct {
	Column string
	Op     string
	Value  string
	Value2 string
}

// ExportPreview is the result of a preview query (≤500 rows + total count).
type ExportPreview struct {
	Rows       []map[string]any
	TotalCount int
}

// buildExportQuery assembles the full SQL query for the given dataset and options.
// Column keys and the sort field are looked up in the dataset definition — an
// unknown key returns an error so the caller surfaces bad state clearly rather
// than silently producing wrong SQL.
// limit=0 means no LIMIT clause (used by ExportToCSV).
//
//nolint:gocognit // large switch-style allowlist validation — splitting would obscure the mapping
func buildExportQuery(def datasetDef, opts ExportOptions, limit int) (string, []any, error) {
	if len(opts.Columns) == 0 {
		return "", nil, fmt.Errorf("no columns selected for export")
	}

	// Validate and collect SELECT expressions.
	selects := make([]string, 0, len(opts.Columns))
	for _, key := range opts.Columns {
		col, ok := def.cols[key]
		if !ok {
			return "", nil, fmt.Errorf("unknown column key %q for dataset %q", key, def.id)
		}
		selects = append(selects, col.sqlExpr+" AS "+key)
	}

	// Build extra WHERE conditions from the filter rows.
	var extraConds []string
	var args []any

	// career_batting / career_pitching: apply stat_type filter.
	if def.id == "career_batting" || def.id == "career_pitching" {
		st := opts.CareerStatType
		if st == "" {
			st = "regular_season"
		}
		if st != "regular_season" && st != "playoffs" && st != "total_career" {
			return "", nil, fmt.Errorf("invalid career stat type %q", st)
		}
		extraConds = append(extraConds, "cbs.stat_type = ?")
		if def.id == "career_pitching" {
			extraConds[len(extraConds)-1] = "cps.stat_type = ?"
		}
		args = append(args, st)
	}

	for _, f := range opts.Filters {
		col, ok := def.cols[f.Column]
		if !ok {
			// Silently skip filter rows for columns not in this dataset — allows
			// the frontend to send a partial config without causing errors.
			continue
		}
		sqlOp, ok := validFilterOps[f.Op]
		if !ok {
			return "", nil, fmt.Errorf("unknown filter op %q", f.Op)
		}
		if f.Op == "contains" {
			extraConds = append(extraConds, col.sqlExpr+" LIKE ?")
			args = append(args, "%"+f.Value+"%")
		} else {
			extraConds = append(extraConds, col.sqlExpr+" "+sqlOp+" ?")
			args = append(args, f.Value)
		}
	}

	// Determine ORDER BY.
	orderBy := ""
	if opts.SortCol != "" {
		col, ok := def.cols[opts.SortCol]
		if !ok {
			return "", nil, fmt.Errorf("unknown sort column %q for dataset %q", opts.SortCol, def.id)
		}
		dir := "ASC"
		if strings.EqualFold(opts.SortDir, "desc") {
			dir = "DESC"
		}
		orderBy = "\nORDER BY " + col.sqlExpr + " " + dir
	}

	// Assemble the FROM clause: the datasetDef already includes a fixed WHERE
	// (e.g. is_regular_season = 1 for season datasets). We append extra conditions
	// with AND.
	fromClause := def.fromSQL
	if len(extraConds) > 0 {
		fromClause += "\n  AND " + strings.Join(extraConds, "\n  AND ")
	}

	limitClause := ""
	if limit > 0 {
		limitClause = "\nLIMIT ?"
		args = append(args, limit)
	}

	q := "SELECT " + strings.Join(selects, ", ") + "\n" + fromClause + orderBy + limitClause
	return q, args, nil
}

// PreviewExportData executes the export query with a 500-row limit and also
// runs a COUNT(*) to return the total matching row count.
func (s *ExportStore) PreviewExportData(ctx context.Context, opts ExportOptions) (ExportPreview, error) {
	def, ok := datasetByID[opts.DatasetID]
	if !ok {
		return ExportPreview{}, fmt.Errorf("unknown dataset %q", opts.DatasetID)
	}

	// Build the data query (limit 500) and a count query sharing the same WHERE.
	dataQ, dataArgs, err := buildExportQuery(def, opts, 500)
	if err != nil {
		return ExportPreview{}, err
	}

	// For the count we re-run buildExportQuery with a single dummy column and no limit,
	// then wrap it to avoid duplicating the WHERE building logic.
	countOpts := opts
	countOpts.Columns = []string{opts.Columns[0]}
	countOpts.SortCol = ""
	baseQ, baseArgs, err := buildExportQuery(def, countOpts, 0)
	if err != nil {
		return ExportPreview{}, err
	}
	countQ := "SELECT COUNT(*) FROM (" + baseQ + ")"

	var total int
	if err := s.db.QueryRowContext(ctx, countQ, baseArgs...).Scan(&total); err != nil {
		return ExportPreview{}, fmt.Errorf("export preview count (dataset=%s): %w", opts.DatasetID, err)
	}

	rows, err := s.db.QueryContext(ctx, dataQ, dataArgs...)
	if err != nil {
		return ExportPreview{}, fmt.Errorf("export preview query (dataset=%s): %w", opts.DatasetID, err)
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return ExportPreview{}, fmt.Errorf("export preview columns: %w", err)
	}

	var result []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return ExportPreview{}, fmt.Errorf("export preview scan: %w", err)
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			row[c] = vals[i]
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return ExportPreview{}, fmt.Errorf("export preview rows: %w", err)
	}

	return ExportPreview{Rows: result, TotalCount: total}, nil
}

// csvHeaders resolves each column key to its display label using the dataset definition.
// Falls back to the key itself if not found (should not happen for validated queries).
func csvHeaders(def datasetDef, keys []string) []string {
	headers := make([]string, len(keys))
	for i, key := range keys {
		if col, ok := def.cols[key]; ok {
			headers[i] = col.label
		} else {
			headers[i] = key
		}
	}
	return headers
}

// anyToCSVRecord converts a slice of scanned SQLite values to CSV strings.
// nil (NULL) becomes an empty string; everything else uses %v formatting.
func anyToCSVRecord(vals []any) []string {
	record := make([]string, len(vals))
	for i, v := range vals {
		if v == nil {
			record[i] = ""
		} else {
			record[i] = fmt.Sprintf("%v", v)
		}
	}
	return record
}

// ExportToCSV runs the full (unlimited) export query and writes the results to
// a CSV byte buffer. The first row is the column label headers.
func (s *ExportStore) ExportToCSV(ctx context.Context, opts ExportOptions) ([]byte, error) {
	def, ok := datasetByID[opts.DatasetID]
	if !ok {
		return nil, fmt.Errorf("unknown dataset %q", opts.DatasetID)
	}

	q, args, err := buildExportQuery(def, opts, 0)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("export CSV query (dataset=%s): %w", opts.DatasetID, err)
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("export CSV columns: %w", err)
	}

	var buf strings.Builder
	w := csv.NewWriter(&buf)
	if err := w.Write(csvHeaders(def, cols)); err != nil {
		return nil, fmt.Errorf("export CSV write header: %w", err)
	}

	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("export CSV scan: %w", err)
		}
		if err := w.Write(anyToCSVRecord(vals)); err != nil {
			return nil, fmt.Errorf("export CSV write row: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("export CSV rows: %w", err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("export CSV flush: %w", err)
	}

	return []byte(buf.String()), nil
}
