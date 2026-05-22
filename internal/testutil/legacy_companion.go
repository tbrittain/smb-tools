package testutil

import (
	"database/sql"
	"testing"
)

// NewLegacyCompanionDB creates an in-memory SQLite database matching the exact
// SmbExplorerCompanion EF Core schema (confirmed from 20230731165716_Initial.cs
// and subsequent migrations). Seeded with two franchises:
//
//   - Franchise A (Id=1): 2 seasons, 3 players (batter, SP, RP), regular + playoff
//     stats, traits, pitch types, one built-in award (MVP), one user-defined award,
//     a championship winner, and a full schedule.
//   - Franchise B (Id=2): 1 season, 2 players, regular season only — used by
//     TestMigrateLegacy_TwoFranchises to assert no cross-contamination.
func NewLegacyCompanionDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("testutil.NewLegacyCompanionDB: open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := createLegacySchema(db); err != nil {
		t.Fatalf("testutil.NewLegacyCompanionDB: schema: %v", err)
	}
	if err := seedLegacyData(db); err != nil {
		t.Fatalf("testutil.NewLegacyCompanionDB: seed: %v", err)
	}
	return db
}

func createLegacySchema(db *sql.DB) error {
	_, err := db.Exec(`
		-- Lookup tables
		CREATE TABLE BatHandedness (
			Id   INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL
		);
		CREATE TABLE ThrowHandedness (
			Id   INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL
		);
		CREATE TABLE PitcherRoles (
			Id   INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL
		);
		CREATE TABLE PitchTypes (
			Id   INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL
		);
		CREATE TABLE Positions (
			Id               INTEGER PRIMARY KEY AUTOINCREMENT,
			Name             TEXT NOT NULL,
			IsPrimaryPosition INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE Chemistry (
			Id   INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL
		);
		CREATE TABLE Traits (
			Id          INTEGER PRIMARY KEY AUTOINCREMENT,
			Name        TEXT NOT NULL,
			IsSmb3      INTEGER NOT NULL DEFAULT 0,
			IsPositive  INTEGER NOT NULL DEFAULT 0,
			ChemistryId INTEGER REFERENCES Chemistry(Id)
		);
		CREATE TABLE PlayerAwards (
			Id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			Name               TEXT NOT NULL,
			OriginalName       TEXT NOT NULL,
			IsBuiltIn          INTEGER NOT NULL DEFAULT 1,
			Importance         INTEGER NOT NULL DEFAULT 0,
			OmitFromGroupings  INTEGER NOT NULL DEFAULT 0,
			IsBattingAward     INTEGER NOT NULL DEFAULT 0,
			IsPitchingAward    INTEGER NOT NULL DEFAULT 0,
			IsFieldingAward    INTEGER NOT NULL DEFAULT 0,
			IsPlayoffAward     INTEGER NOT NULL DEFAULT 0,
			IsUserAssignable   INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE LookupSeeds (
			Id       TEXT PRIMARY KEY,
			SeededAt TEXT NOT NULL
		);

		-- Core tables
		CREATE TABLE Franchises (
			Id     INTEGER PRIMARY KEY AUTOINCREMENT,
			Name   TEXT NOT NULL,
			IsSmb3 INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE Conferences (
			Id                  INTEGER PRIMARY KEY AUTOINCREMENT,
			Name                TEXT NOT NULL,
			FranchiseId         INTEGER NOT NULL REFERENCES Franchises(Id),
			IsDesignatedHitter  INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE Divisions (
			Id           INTEGER PRIMARY KEY AUTOINCREMENT,
			Name         TEXT NOT NULL,
			ConferenceId INTEGER NOT NULL REFERENCES Conferences(Id)
		);
		CREATE TABLE Seasons (
			Id                    INTEGER PRIMARY KEY,
			Number                INTEGER NOT NULL,
			NumGamesRegularSeason INTEGER NOT NULL,
			FranchiseId           INTEGER NOT NULL REFERENCES Franchises(Id),
			ChampionshipWinnerId  INTEGER
		);
		CREATE TABLE Teams (
			Id          INTEGER PRIMARY KEY AUTOINCREMENT,
			FranchiseId INTEGER NOT NULL REFERENCES Franchises(Id)
		);
		CREATE TABLE TeamLogoHistory (
			Id          INTEGER PRIMARY KEY AUTOINCREMENT,
			LogoFullSize BLOB NOT NULL,
			LogoIconSize BLOB NOT NULL,
			"Order"     INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE TeamNameHistory (
			Id               INTEGER PRIMARY KEY AUTOINCREMENT,
			Name             TEXT NOT NULL,
			TeamLogoHistoryId INTEGER REFERENCES TeamLogoHistory(Id)
		);
		CREATE TABLE TeamGameIdHistory (
			Id     INTEGER PRIMARY KEY AUTOINCREMENT,
			TeamId INTEGER NOT NULL REFERENCES Teams(Id),
			GameId TEXT NOT NULL
		);
		CREATE TABLE SeasonTeamHistory (
			Id                     INTEGER PRIMARY KEY AUTOINCREMENT,
			SeasonId               INTEGER NOT NULL REFERENCES Seasons(Id),
			TeamId                 INTEGER NOT NULL REFERENCES Teams(Id),
			DivisionId             INTEGER NOT NULL REFERENCES Divisions(Id),
			TeamNameHistoryId      INTEGER NOT NULL REFERENCES TeamNameHistory(Id),
			Budget                 INTEGER NOT NULL DEFAULT 0,
			Payroll                INTEGER NOT NULL DEFAULT 0,
			Surplus                INTEGER NOT NULL DEFAULT 0,
			SurplusPerGame         REAL    NOT NULL DEFAULT 0,
			Wins                   INTEGER NOT NULL DEFAULT 0,
			Losses                 INTEGER NOT NULL DEFAULT 0,
			GamesBehind            REAL    NOT NULL DEFAULT 0,
			WinPercentage          REAL    NOT NULL DEFAULT 0,
			PythagoreanWinPercentage REAL  NOT NULL DEFAULT 0,
			ExpectedWins           INTEGER NOT NULL DEFAULT 0,
			ExpectedLosses         INTEGER NOT NULL DEFAULT 0,
			RunsScored             INTEGER NOT NULL DEFAULT 0,
			RunsAllowed            INTEGER NOT NULL DEFAULT 0,
			TotalPower             INTEGER NOT NULL DEFAULT 0,
			TotalContact           INTEGER NOT NULL DEFAULT 0,
			TotalSpeed             INTEGER NOT NULL DEFAULT 0,
			TotalFielding          INTEGER NOT NULL DEFAULT 0,
			TotalArm               INTEGER NOT NULL DEFAULT 0,
			TotalVelocity          INTEGER NOT NULL DEFAULT 0,
			TotalJunk              INTEGER NOT NULL DEFAULT 0,
			TotalAccuracy          INTEGER NOT NULL DEFAULT 0,
			PlayoffSeed            INTEGER,
			PlayoffWins            INTEGER,
			PlayoffLosses          INTEGER,
			PlayoffRunsScored      INTEGER,
			PlayoffRunsAllowed     INTEGER
		);
		CREATE TABLE ChampionshipWinners (
			Id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			SeasonTeamHistoryId INTEGER NOT NULL UNIQUE REFERENCES SeasonTeamHistory(Id),
			SeasonId           INTEGER NOT NULL UNIQUE REFERENCES Seasons(Id)
		);
		CREATE TABLE Players (
			Id               INTEGER PRIMARY KEY AUTOINCREMENT,
			FirstName        TEXT NOT NULL,
			LastName         TEXT NOT NULL,
			IsHallOfFamer    INTEGER NOT NULL DEFAULT 0,
			BatHandednessId  INTEGER NOT NULL REFERENCES BatHandedness(Id),
			ThrowHandednessId INTEGER NOT NULL REFERENCES ThrowHandedness(Id),
			PrimaryPositionId INTEGER NOT NULL REFERENCES Positions(Id),
			PitcherRoleId    INTEGER REFERENCES PitcherRoles(Id),
			ChemistryId      INTEGER REFERENCES Chemistry(Id),
			FranchiseId      INTEGER NOT NULL REFERENCES Franchises(Id)
		);
		CREATE TABLE PlayerGameIdHistory (
			Id       INTEGER PRIMARY KEY AUTOINCREMENT,
			PlayerId INTEGER NOT NULL REFERENCES Players(Id),
			GameId   TEXT NOT NULL
		);
		CREATE TABLE PlayerSeasons (
			Id                  INTEGER PRIMARY KEY AUTOINCREMENT,
			PlayerId            INTEGER NOT NULL REFERENCES Players(Id),
			SeasonId            INTEGER NOT NULL REFERENCES Seasons(Id),
			Age                 INTEGER NOT NULL DEFAULT 0,
			Salary              INTEGER NOT NULL DEFAULT 0,
			SecondaryPositionId INTEGER REFERENCES Positions(Id),
			ChampionshipWinnerId INTEGER REFERENCES ChampionshipWinners(Id)
		);
		CREATE TABLE PlayerSeasonGameStats (
			Id             INTEGER PRIMARY KEY AUTOINCREMENT,
			PlayerSeasonId INTEGER NOT NULL UNIQUE REFERENCES PlayerSeasons(Id),
			Power          INTEGER NOT NULL DEFAULT 0,
			Contact        INTEGER NOT NULL DEFAULT 0,
			Speed          INTEGER NOT NULL DEFAULT 0,
			Fielding       INTEGER NOT NULL DEFAULT 0,
			Arm            INTEGER,
			Velocity       INTEGER,
			Junk           INTEGER,
			Accuracy       INTEGER
		);
		CREATE TABLE PlayerSeasonBattingStats (
			Id                    INTEGER PRIMARY KEY AUTOINCREMENT,
			PlayerSeasonId        INTEGER NOT NULL REFERENCES PlayerSeasons(Id),
			GamesPlayed           INTEGER NOT NULL DEFAULT 0,
			GamesBatting          INTEGER NOT NULL DEFAULT 0,
			AtBats                INTEGER NOT NULL DEFAULT 0,
			PlateAppearances      INTEGER NOT NULL DEFAULT 0,
			Runs                  INTEGER NOT NULL DEFAULT 0,
			Hits                  INTEGER NOT NULL DEFAULT 0,
			Singles               INTEGER NOT NULL DEFAULT 0,
			Doubles               INTEGER NOT NULL DEFAULT 0,
			Triples               INTEGER NOT NULL DEFAULT 0,
			HomeRuns              INTEGER NOT NULL DEFAULT 0,
			RunsBattedIn          INTEGER NOT NULL DEFAULT 0,
			ExtraBaseHits         INTEGER NOT NULL DEFAULT 0,
			TotalBases            INTEGER NOT NULL DEFAULT 0,
			StolenBases           INTEGER NOT NULL DEFAULT 0,
			CaughtStealing        INTEGER NOT NULL DEFAULT 0,
			Walks                 INTEGER NOT NULL DEFAULT 0,
			Strikeouts            INTEGER NOT NULL DEFAULT 0,
			HitByPitch            INTEGER NOT NULL DEFAULT 0,
			SacrificeHits         INTEGER NOT NULL DEFAULT 0,
			SacrificeFlies        INTEGER NOT NULL DEFAULT 0,
			Errors                INTEGER NOT NULL DEFAULT 0,
			PassedBalls           INTEGER NOT NULL DEFAULT 0,
			Obp                   REAL,
			Slg                   REAL,
			Ops                   REAL,
			Woba                  REAL,
			Iso                   REAL,
			Babip                 REAL,
			BattingAverage        REAL,
			PaPerGame             REAL,
			AbPerHomeRun          REAL,
			StrikeoutPercentage   REAL,
			WalkPercentage        REAL,
			ExtraBaseHitPercentage REAL,
			OpsPlus               REAL,
			IsRegularSeason       INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE PlayerSeasonPitchingStats (
			Id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			PlayerSeasonId       INTEGER NOT NULL REFERENCES PlayerSeasons(Id),
			Wins                 INTEGER NOT NULL DEFAULT 0,
			Losses               INTEGER NOT NULL DEFAULT 0,
			CompleteGames        INTEGER NOT NULL DEFAULT 0,
			Shutouts             INTEGER NOT NULL DEFAULT 0,
			Hits                 INTEGER NOT NULL DEFAULT 0,
			EarnedRuns           INTEGER NOT NULL DEFAULT 0,
			HomeRuns             INTEGER NOT NULL DEFAULT 0,
			Walks                INTEGER NOT NULL DEFAULT 0,
			Strikeouts           INTEGER NOT NULL DEFAULT 0,
			InningsPitched       REAL,
			EarnedRunAverage     REAL,
			TotalPitches         INTEGER NOT NULL DEFAULT 0,
			Saves                INTEGER NOT NULL DEFAULT 0,
			HitByPitch           INTEGER NOT NULL DEFAULT 0,
			BattersFaced         INTEGER NOT NULL DEFAULT 0,
			GamesPlayed          INTEGER NOT NULL DEFAULT 0,
			GamesStarted         INTEGER NOT NULL DEFAULT 0,
			GamesFinished        INTEGER NOT NULL DEFAULT 0,
			RunsAllowed          INTEGER NOT NULL DEFAULT 0,
			WildPitches          INTEGER NOT NULL DEFAULT 0,
			BattingAverageAgainst REAL,
			Fip                  REAL,
			Whip                 REAL,
			WinPercentage        REAL,
			OpponentObp          REAL,
			StrikeoutsPerWalk    REAL,
			StrikeoutsPerNine    REAL,
			WalksPerNine         REAL,
			HitsPerNine          REAL,
			HomeRunsPerNine      REAL,
			PitchesPerInning     REAL,
			PitchesPerGame       REAL,
			EraMinus             REAL,
			FipMinus             REAL,
			IsRegularSeason      INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE PlayerTeamHistory (
			Id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			PlayerSeasonId     INTEGER NOT NULL REFERENCES PlayerSeasons(Id),
			SeasonTeamHistoryId INTEGER REFERENCES SeasonTeamHistory(Id),
			"Order"            INTEGER NOT NULL DEFAULT 1
		);
		-- Junction tables (EF auto-generated names confirmed from migration)
		CREATE TABLE PitchTypePlayerSeason (
			PitchTypesId   INTEGER NOT NULL REFERENCES PitchTypes(Id),
			PlayerSeasonsId INTEGER NOT NULL REFERENCES PlayerSeasons(Id),
			PRIMARY KEY (PitchTypesId, PlayerSeasonsId)
		);
		CREATE TABLE PlayerAwardPlayerSeason (
			AwardsId        INTEGER NOT NULL REFERENCES PlayerAwards(Id),
			PlayerSeasonsId INTEGER NOT NULL REFERENCES PlayerSeasons(Id),
			PRIMARY KEY (AwardsId, PlayerSeasonsId)
		);
		CREATE TABLE PlayerSeasonTrait (
			PlayerSeasonsId INTEGER NOT NULL REFERENCES PlayerSeasons(Id),
			TraitsId        INTEGER NOT NULL REFERENCES Traits(Id),
			PRIMARY KEY (PlayerSeasonsId, TraitsId)
		);
		CREATE TABLE TeamSeasonSchedules (
			Id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			HomeTeamHistoryId  INTEGER NOT NULL REFERENCES SeasonTeamHistory(Id),
			AwayTeamHistoryId  INTEGER NOT NULL REFERENCES SeasonTeamHistory(Id),
			HomePitcherSeasonId INTEGER REFERENCES PlayerSeasons(Id),
			AwayPitcherSeasonId INTEGER REFERENCES PlayerSeasons(Id),
			Day                INTEGER NOT NULL DEFAULT 0,
			GlobalGameNumber   INTEGER NOT NULL,
			HomeScore          INTEGER,
			AwayScore          INTEGER
		);
		CREATE TABLE TeamPlayoffSchedules (
			Id                 INTEGER PRIMARY KEY AUTOINCREMENT,
			HomeTeamHistoryId  INTEGER NOT NULL REFERENCES SeasonTeamHistory(Id),
			AwayTeamHistoryId  INTEGER NOT NULL REFERENCES SeasonTeamHistory(Id),
			HomePitcherSeasonId INTEGER REFERENCES PlayerSeasons(Id),
			AwayPitcherSeasonId INTEGER REFERENCES PlayerSeasons(Id),
			SeriesNumber       INTEGER NOT NULL,
			GlobalGameNumber   INTEGER NOT NULL,
			HomeScore          INTEGER,
			AwayScore          INTEGER
		);
	`)
	return err
}

func seedLegacyData(db *sql.DB) error {
	_, err := db.Exec(`
		-- ── Lookups ───────────────────────────────────────────────────────────────
		INSERT INTO BatHandedness (Id, Name) VALUES (1,'R'),(2,'L'),(3,'S');
		INSERT INTO ThrowHandedness (Id, Name) VALUES (1,'R'),(2,'L');
		INSERT INTO PitcherRoles (Id, Name) VALUES (1,'SP'),(2,'RP'),(3,'SP/RP'),(4,'CL');
		INSERT INTO PitchTypes (Id, Name) VALUES (1,'4F'),(2,'2F'),(3,'SB'),(4,'CH'),(5,'FK'),(6,'CB'),(7,'SL'),(8,'CF');
		INSERT INTO Positions (Id, Name, IsPrimaryPosition) VALUES
			(1,'P',1),(2,'C',1),(3,'1B',1),(4,'2B',1),(5,'3B',1),
			(6,'SS',1),(7,'LF',1),(8,'CF',1),(9,'RF',1),
			(10,'IF',0),(11,'OF',0),(12,'1B/OF',0),(13,'IF/OF',0);
		INSERT INTO Chemistry (Id, Name) VALUES
			(1,'Competitive'),(2,'Spirited'),(3,'Disciplined'),(4,'Scholarly'),(5,'Crafty');
		INSERT INTO Traits (Id, Name, IsSmb3, IsPositive, ChemistryId) VALUES
			(1,'Clutch',0,1,2),
			(2,'Tough Out',0,1,1),
			(3,'Whiffer',0,0,1);
		INSERT INTO PlayerAwards (Id, Name, OriginalName, IsBuiltIn, Importance, OmitFromGroupings,
		                          IsBattingAward, IsPitchingAward, IsFieldingAward, IsPlayoffAward, IsUserAssignable)
		VALUES
			(1,'MVP','MVP',1,0,0, 1,1,0,0,1),
			(2,'Cy Young','Cy Young',1,1,0, 0,1,0,0,1),
			(100,'Greatest Slugger','Greatest Slugger',0,3,0, 1,0,0,0,1);

		-- ── Franchises ────────────────────────────────────────────────────────────
		INSERT INTO Franchises (Id, Name, IsSmb3) VALUES (1,'Test Franchise A',0);
		INSERT INTO Franchises (Id, Name, IsSmb3) VALUES (2,'Test Franchise B',0);

		-- ── Franchise A: conferences, divisions ───────────────────────────────────
		INSERT INTO Conferences (Id, Name, FranchiseId, IsDesignatedHitter) VALUES (1,'East',1,0);
		INSERT INTO Conferences (Id, Name, FranchiseId, IsDesignatedHitter) VALUES (2,'West',1,0);
		INSERT INTO Divisions (Id, Name, ConferenceId) VALUES (1,'East Division',1);
		INSERT INTO Divisions (Id, Name, ConferenceId) VALUES (2,'West Division',2);

		-- ── Franchise A: seasons ──────────────────────────────────────────────────
		-- Season.Id is the save game season ID (NOT autoincrement — set explicitly)
		INSERT INTO Seasons (Id, Number, NumGamesRegularSeason, FranchiseId, ChampionshipWinnerId)
		VALUES (10, 1, 60, 1, NULL);
		INSERT INTO Seasons (Id, Number, NumGamesRegularSeason, FranchiseId, ChampionshipWinnerId)
		VALUES (11, 2, 60, 1, NULL);

		-- ── Franchise A: teams ────────────────────────────────────────────────────
		INSERT INTO Teams (Id, FranchiseId) VALUES (1, 1);  -- Alpha Squad
		INSERT INTO Teams (Id, FranchiseId) VALUES (2, 1);  -- Beta Ballers
		INSERT INTO TeamGameIdHistory (Id, TeamId, GameId) VALUES
			(1, 1, 'aaaaaaaa-0001-0000-0000-000000000000'),
			(2, 1, 'aaaaaaaa-0001-0001-0000-000000000000'),  -- alt GUID (fork)
			(3, 2, 'bbbbbbbb-0002-0000-0000-000000000000');
		INSERT INTO TeamNameHistory (Id, Name, TeamLogoHistoryId) VALUES
			(1, 'Alpha Squad', NULL),
			(2, 'Beta Ballers', NULL);

		-- ── Franchise A: season team history ──────────────────────────────────────
		-- Season 10 (Number=1)
		INSERT INTO SeasonTeamHistory
		    (Id,SeasonId,TeamId,DivisionId,TeamNameHistoryId,
		     Budget,Payroll,Surplus,SurplusPerGame,
		     Wins,Losses,GamesBehind,WinPercentage,PythagoreanWinPercentage,
		     ExpectedWins,ExpectedLosses,RunsScored,RunsAllowed,
		     TotalPower,TotalContact,TotalSpeed,TotalFielding,TotalArm,
		     TotalVelocity,TotalJunk,TotalAccuracy,
		     PlayoffSeed,PlayoffWins,PlayoffLosses,PlayoffRunsScored,PlayoffRunsAllowed)
		VALUES
		    (1,10,1,1,1,  5000,4200,800,13,  40,20,0,0.667,0.650,42,18,200,160,
		     80,75,60,70,65,88,78,82,  1,3,1,15,8),
		    (2,10,2,2,2,  4800,4500,300,5,   20,40,20,0.333,0.320,18,42,160,200,
		     72,68,56,67,61,56,51,51,  2,1,3,8,15);
		-- Season 11 (Number=2) — Alpha Squad is champion
		INSERT INTO SeasonTeamHistory
		    (Id,SeasonId,TeamId,DivisionId,TeamNameHistoryId,
		     Budget,Payroll,Surplus,SurplusPerGame,
		     Wins,Losses,GamesBehind,WinPercentage,PythagoreanWinPercentage,
		     ExpectedWins,ExpectedLosses,RunsScored,RunsAllowed,
		     TotalPower,TotalContact,TotalSpeed,TotalFielding,TotalArm,
		     TotalVelocity,TotalJunk,TotalAccuracy,
		     PlayoffSeed,PlayoffWins,PlayoffLosses,PlayoffRunsScored,PlayoffRunsAllowed)
		VALUES
		    (3,11,1,1,1,  5500,4400,1100,18,  45,15,0,0.750,0.720,46,14,220,155,
		     85,78,65,72,68,90,80,85,  1,5,0,20,8),
		    (4,11,2,2,2,  4900,4600,300,5,    15,45,30,0.250,0.240,12,48,155,220,
		     73,68,56,67,61,57,51,51,  2,1,3,8,16);

		-- Championship winner for season 11 (Alpha Squad, STH Id=3)
		INSERT INTO ChampionshipWinners (Id, SeasonTeamHistoryId, SeasonId) VALUES (1, 3, 11);
		UPDATE Seasons SET ChampionshipWinnerId = 1 WHERE Id = 11;

		-- ── Franchise A: players ──────────────────────────────────────────────────
		-- Player 1: Alex Power — batter, CF, R/R, Competitive, HoF
		INSERT INTO Players (Id,FirstName,LastName,IsHallOfFamer,
		                     BatHandednessId,ThrowHandednessId,PrimaryPositionId,
		                     PitcherRoleId,ChemistryId,FranchiseId)
		VALUES (1,'Alex','Power',1, 1,1,8, NULL,1,1);
		INSERT INTO PlayerGameIdHistory (Id,PlayerId,GameId) VALUES
			(1,1,'aaaaaaaa-aa01-0000-0000-000000000001'),
			(2,1,'aaaaaaaa-aa01-0001-0000-000000000001');  -- alt GUID

		-- Player 2: Sam Strike — SP, P, R/R, Spirited
		INSERT INTO Players (Id,FirstName,LastName,IsHallOfFamer,
		                     BatHandednessId,ThrowHandednessId,PrimaryPositionId,
		                     PitcherRoleId,ChemistryId,FranchiseId)
		VALUES (2,'Sam','Strike',0, 1,1,1, 1,2,1);
		INSERT INTO PlayerGameIdHistory (Id,PlayerId,GameId)
		VALUES (3,2,'bbbbbbbb-bb02-0000-0000-000000000002');

		-- Player 3: Riley Closer — RP, P, L/L, Competitive
		INSERT INTO Players (Id,FirstName,LastName,IsHallOfFamer,
		                     BatHandednessId,ThrowHandednessId,PrimaryPositionId,
		                     PitcherRoleId,ChemistryId,FranchiseId)
		VALUES (3,'Riley','Closer',0, 2,2,1, 2,1,1);
		INSERT INTO PlayerGameIdHistory (Id,PlayerId,GameId)
		VALUES (4,3,'cccccccc-cc03-0000-0000-000000000003');

		-- ── Franchise A: player seasons ───────────────────────────────────────────
		-- Season 10 (SeasonId=10)
		INSERT INTO PlayerSeasons (Id,PlayerId,SeasonId,Age,Salary,SecondaryPositionId)
		VALUES (1,1,10, 27,250,NULL),  -- Alex S10
		       (2,2,10, 30,800,NULL),  -- Sam S10
		       (3,3,10, 25,300,NULL);  -- Riley S10
		-- Season 11 (SeasonId=11)
		INSERT INTO PlayerSeasons (Id,PlayerId,SeasonId,Age,Salary,SecondaryPositionId)
		VALUES (4,1,11, 28,300,NULL),  -- Alex S11
		       (5,2,11, 31,900,NULL),  -- Sam S11
		       (6,3,11, 26,350,NULL);  -- Riley S11

		-- PlayerTeamHistory: all players on Alpha Squad (STH Id=1/3) as current team
		INSERT INTO PlayerTeamHistory (PlayerSeasonId, SeasonTeamHistoryId, "Order") VALUES
			(1,1,1),(2,1,1),(3,1,1),  -- Season 10: Alpha Squad
			(4,3,1),(5,3,1),(6,3,1);  -- Season 11: Alpha Squad

		-- ── Franchise A: game stats ───────────────────────────────────────────────
		-- Alex (batter): no Velocity/Junk/Accuracy (NULL)
		INSERT INTO PlayerSeasonGameStats (PlayerSeasonId,Power,Contact,Speed,Fielding,Arm,Velocity,Junk,Accuracy)
		VALUES (1, 80,75,60,70,65, NULL,NULL,NULL),
		       (4, 82,76,62,71,66, NULL,NULL,NULL);
		-- Sam (SP): all attrs
		INSERT INTO PlayerSeasonGameStats (PlayerSeasonId,Power,Contact,Speed,Fielding,Arm,Velocity,Junk,Accuracy)
		VALUES (2, 40,40,40,50,55,88,78,82),
		       (5, 42,41,41,51,56,90,80,84);
		-- Riley (RP)
		INSERT INTO PlayerSeasonGameStats (PlayerSeasonId,Power,Contact,Speed,Fielding,Arm,Velocity,Junk,Accuracy)
		VALUES (3, 35,35,45,48,52,85,82,80),
		       (6, 36,36,46,49,53,86,83,81);

		-- ── Franchise A: batting stats ────────────────────────────────────────────
		-- Alex season 10, regular
		INSERT INTO PlayerSeasonBattingStats
		    (PlayerSeasonId,GamesPlayed,GamesBatting,AtBats,PlateAppearances,Runs,Hits,Singles,Doubles,Triples,HomeRuns,
		     RunsBattedIn,ExtraBaseHits,TotalBases,StolenBases,CaughtStealing,Walks,Strikeouts,HitByPitch,
		     SacrificeHits,SacrificeFlies,Errors,PassedBalls,IsRegularSeason)
		VALUES (1, 60,60,220,251,35,66, 44,10,2,10, 38,22,100, 15,4,26,44,3, 1,1,2,0,1);
		-- Alex season 10, playoff
		INSERT INTO PlayerSeasonBattingStats
		    (PlayerSeasonId,GamesPlayed,GamesBatting,AtBats,PlateAppearances,Runs,Hits,Singles,Doubles,Triples,HomeRuns,
		     RunsBattedIn,ExtraBaseHits,TotalBases,StolenBases,CaughtStealing,Walks,Strikeouts,HitByPitch,
		     SacrificeHits,SacrificeFlies,Errors,PassedBalls,IsRegularSeason)
		VALUES (1, 8,8,28,31,4,9, 7,1,0,1, 4,2,12, 2,0,3,6,0, 0,0,0,0,0);
		-- Alex season 11, regular
		INSERT INTO PlayerSeasonBattingStats
		    (PlayerSeasonId,GamesPlayed,GamesBatting,AtBats,PlateAppearances,Runs,Hits,Singles,Doubles,Triples,HomeRuns,
		     RunsBattedIn,ExtraBaseHits,TotalBases,StolenBases,CaughtStealing,Walks,Strikeouts,HitByPitch,
		     SacrificeHits,SacrificeFlies,Errors,PassedBalls,IsRegularSeason)
		VALUES (4, 60,60,225,257,40,72, 50,12,2,8, 35,22,108, 18,3,28,40,2, 1,1,1,0,1);

		-- ── Franchise A: pitching stats ───────────────────────────────────────────
		-- Sam season 10, regular — 180.0 IP = 540 outs
		INSERT INTO PlayerSeasonPitchingStats
		    (PlayerSeasonId,Wins,Losses,CompleteGames,Shutouts,Hits,EarnedRuns,HomeRuns,Walks,Strikeouts,
		     InningsPitched,TotalPitches,Saves,HitByPitch,BattersFaced,GamesPlayed,GamesStarted,
		     GamesFinished,RunsAllowed,WildPitches,IsRegularSeason)
		VALUES (2, 12,8,4,1,140,55,15,40,180, 180.0,3200,0,8,740,25,25,5,60,5,1);
		-- Sam season 10, playoff — 18.0 IP = 54 outs
		INSERT INTO PlayerSeasonPitchingStats
		    (PlayerSeasonId,Wins,Losses,CompleteGames,Shutouts,Hits,EarnedRuns,HomeRuns,Walks,Strikeouts,
		     InningsPitched,TotalPitches,Saves,HitByPitch,BattersFaced,GamesPlayed,GamesStarted,
		     GamesFinished,RunsAllowed,WildPitches,IsRegularSeason)
		VALUES (2, 2,0,0,0,10,3,1,5,18, 18.0,350,0,1,62,2,2,0,3,1,0);
		-- Riley season 10, regular — 45.2 IP = 137 outs
		INSERT INTO PlayerSeasonPitchingStats
		    (PlayerSeasonId,Wins,Losses,CompleteGames,Shutouts,Hits,EarnedRuns,HomeRuns,Walks,Strikeouts,
		     InningsPitched,TotalPitches,Saves,HitByPitch,BattersFaced,GamesPlayed,GamesStarted,
		     GamesFinished,RunsAllowed,WildPitches,IsRegularSeason)
		VALUES (3, 4,2,0,0,30,10,3,15,60, 45.2,800,12,2,200,40,0,30,12,2,1);
		-- Sam season 11, regular — 185.1 IP = 556 outs
		INSERT INTO PlayerSeasonPitchingStats
		    (PlayerSeasonId,Wins,Losses,CompleteGames,Shutouts,Hits,EarnedRuns,HomeRuns,Walks,Strikeouts,
		     InningsPitched,TotalPitches,Saves,HitByPitch,BattersFaced,GamesPlayed,GamesStarted,
		     GamesFinished,RunsAllowed,WildPitches,IsRegularSeason)
		VALUES (5, 14,7,5,2,135,50,12,38,192, 185.1,3400,0,7,752,26,26,4,55,4,1);

		-- ── Franchise A: traits ───────────────────────────────────────────────────
		-- Alex S10 has Clutch (1) and Tough Out (2)
		INSERT INTO PlayerSeasonTrait (PlayerSeasonsId, TraitsId) VALUES (1,1),(1,2);
		-- Alex S11 has Clutch (1) only
		INSERT INTO PlayerSeasonTrait (PlayerSeasonsId, TraitsId) VALUES (4,1);

		-- ── Franchise A: pitch types ──────────────────────────────────────────────
		-- Sam S10: 4F (1), SL (7)
		INSERT INTO PitchTypePlayerSeason (PitchTypesId, PlayerSeasonsId) VALUES (1,2),(7,2);
		-- Sam S11: 4F (1), 2F (2), SL (7)
		INSERT INTO PitchTypePlayerSeason (PitchTypesId, PlayerSeasonsId) VALUES (1,5),(2,5),(7,5);

		-- ── Franchise A: awards ───────────────────────────────────────────────────
		-- Alex S10: MVP (built-in)
		INSERT INTO PlayerAwardPlayerSeason (AwardsId, PlayerSeasonsId) VALUES (1, 1);
		-- Alex S11: Greatest Slugger (user-defined, Id=100)
		INSERT INTO PlayerAwardPlayerSeason (AwardsId, PlayerSeasonsId) VALUES (100, 4);

		-- ── Franchise A: schedules ────────────────────────────────────────────────
		-- Season 10 regular: 2 games
		INSERT INTO TeamSeasonSchedules (HomeTeamHistoryId,AwayTeamHistoryId,HomePitcherSeasonId,AwayPitcherSeasonId,Day,GlobalGameNumber,HomeScore,AwayScore)
		VALUES (1,2, 2,NULL, 1,1, 5,3),
		       (2,1, NULL,2, 2,2, 2,7);
		-- Season 10 playoff: 1 game, series 1
		INSERT INTO TeamPlayoffSchedules (HomeTeamHistoryId,AwayTeamHistoryId,HomePitcherSeasonId,AwayPitcherSeasonId,SeriesNumber,GlobalGameNumber,HomeScore,AwayScore)
		VALUES (1,2, 2,NULL, 1,1, 4,1);
		-- Season 11 regular: 1 game
		INSERT INTO TeamSeasonSchedules (HomeTeamHistoryId,AwayTeamHistoryId,HomePitcherSeasonId,AwayPitcherSeasonId,Day,GlobalGameNumber,HomeScore,AwayScore)
		VALUES (3,4, 5,NULL, 1,1, 6,2);

		-- ══ Franchise B ══════════════════════════════════════════════════════════
		INSERT INTO Conferences (Id, Name, FranchiseId, IsDesignatedHitter) VALUES (3,'NL West',2,0);
		INSERT INTO Divisions   (Id, Name, ConferenceId) VALUES (3,'NL West Div',3);
		INSERT INTO Seasons (Id, Number, NumGamesRegularSeason, FranchiseId, ChampionshipWinnerId)
		VALUES (20, 1, 60, 2, NULL);
		INSERT INTO Teams (Id, FranchiseId) VALUES (3, 2), (4, 2);
		INSERT INTO TeamGameIdHistory (Id, TeamId, GameId) VALUES
			(5, 3, 'dddddddd-d003-0000-0000-000000000003'),
			(6, 4, 'eeeeeeee-e004-0000-0000-000000000004');
		INSERT INTO TeamNameHistory (Id, Name, TeamLogoHistoryId) VALUES
			(3,'Gamma Stars',NULL),(4,'Delta Dynamos',NULL);
		INSERT INTO SeasonTeamHistory
		    (Id,SeasonId,TeamId,DivisionId,TeamNameHistoryId,
		     Budget,Payroll,Surplus,SurplusPerGame,
		     Wins,Losses,GamesBehind,WinPercentage,PythagoreanWinPercentage,
		     ExpectedWins,ExpectedLosses,RunsScored,RunsAllowed,
		     TotalPower,TotalContact,TotalSpeed,TotalFielding,TotalArm,
		     TotalVelocity,TotalJunk,TotalAccuracy)
		VALUES
		    (5,20,3,3,3, 4000,3800,200,3, 35,25,0,0.583,0.570,36,24,180,160, 75,70,58,68,62,80,70,75),
		    (6,20,4,3,4, 3800,3700,100,2, 25,35,10,0.417,0.410,24,36,160,180, 70,65,55,65,58,75,65,70);
		-- Players for Franchise B
		INSERT INTO Players (Id,FirstName,LastName,IsHallOfFamer,
		                     BatHandednessId,ThrowHandednessId,PrimaryPositionId,
		                     PitcherRoleId,ChemistryId,FranchiseId)
		VALUES (4,'Chris','Slugger',0, 1,1,8, NULL,1,2),
		       (5,'Pat','Pitcher',0,  1,1,1, 1,2,2);
		INSERT INTO PlayerGameIdHistory (Id,PlayerId,GameId) VALUES
			(5,4,'ffffffff-f004-0000-0000-000000000004'),
			(6,5,'gggggggg-g005-0000-0000-000000000005');
		INSERT INTO PlayerSeasons (Id,PlayerId,SeasonId,Age,Salary) VALUES
			(7,4,20, 24,200),(8,5,20, 28,400);
		INSERT INTO PlayerTeamHistory (PlayerSeasonId, SeasonTeamHistoryId, "Order") VALUES
			(7,5,1),(8,5,1);
		INSERT INTO PlayerSeasonGameStats (PlayerSeasonId,Power,Contact,Speed,Fielding,Arm)
		VALUES (7,70,65,55,65,60),(8,38,38,40,48,50);
		INSERT INTO PlayerSeasonBattingStats
		    (PlayerSeasonId,GamesPlayed,GamesBatting,AtBats,PlateAppearances,Runs,Hits,Singles,Doubles,Triples,HomeRuns,
		     RunsBattedIn,ExtraBaseHits,TotalBases,StolenBases,CaughtStealing,Walks,Strikeouts,HitByPitch,
		     SacrificeHits,SacrificeFlies,Errors,PassedBalls,IsRegularSeason)
		VALUES (7, 55,55,200,225,28,58,44,8,2,4, 20,14,76, 10,2,22,38,2, 1,0,1,0,1);
		INSERT INTO PlayerSeasonPitchingStats
		    (PlayerSeasonId,Wins,Losses,CompleteGames,Shutouts,Hits,EarnedRuns,HomeRuns,Walks,Strikeouts,
		     InningsPitched,TotalPitches,Saves,HitByPitch,BattersFaced,GamesPlayed,GamesStarted,
		     GamesFinished,RunsAllowed,WildPitches,IsRegularSeason)
		VALUES (8, 10,10,3,0,130,60,14,45,150, 165.0,3000,0,6,680,25,25,4,65,6,1);
	`)
	return err
}
