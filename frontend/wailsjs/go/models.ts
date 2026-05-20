export namespace main {
	
	export class BattingLeaderRowDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    isHallOfFamer: boolean;
	    seasonsPlayed: number;
	    seasonNum: number;
	    teamName: string;
	    age: number;
	    primaryPosition: string;
	    batHand: string;
	    chemistryType: string;
	    gamesPlayed: number;
	    gamesBatting: number;
	    atBats: number;
	    runs: number;
	    hits: number;
	    doubles: number;
	    triples: number;
	    homeRuns: number;
	    rbi: number;
	    stolenBases: number;
	    caughtStealing: number;
	    walks: number;
	    strikeouts: number;
	    hitByPitch: number;
	    sacHits: number;
	    sacFlies: number;
	    errors: number;
	    passedBalls: number;
	    ba?: number;
	    obp?: number;
	    slg?: number;
	    ops?: number;
	    iso?: number;
	    babip?: number;
	    kPct?: number;
	    bbPct?: number;
	    abPerHr?: number;
	
	    static createFrom(source: any = {}) {
	        return new BattingLeaderRowDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.isHallOfFamer = source["isHallOfFamer"];
	        this.seasonsPlayed = source["seasonsPlayed"];
	        this.seasonNum = source["seasonNum"];
	        this.teamName = source["teamName"];
	        this.age = source["age"];
	        this.primaryPosition = source["primaryPosition"];
	        this.batHand = source["batHand"];
	        this.chemistryType = source["chemistryType"];
	        this.gamesPlayed = source["gamesPlayed"];
	        this.gamesBatting = source["gamesBatting"];
	        this.atBats = source["atBats"];
	        this.runs = source["runs"];
	        this.hits = source["hits"];
	        this.doubles = source["doubles"];
	        this.triples = source["triples"];
	        this.homeRuns = source["homeRuns"];
	        this.rbi = source["rbi"];
	        this.stolenBases = source["stolenBases"];
	        this.caughtStealing = source["caughtStealing"];
	        this.walks = source["walks"];
	        this.strikeouts = source["strikeouts"];
	        this.hitByPitch = source["hitByPitch"];
	        this.sacHits = source["sacHits"];
	        this.sacFlies = source["sacFlies"];
	        this.errors = source["errors"];
	        this.passedBalls = source["passedBalls"];
	        this.ba = source["ba"];
	        this.obp = source["obp"];
	        this.slg = source["slg"];
	        this.ops = source["ops"];
	        this.iso = source["iso"];
	        this.babip = source["babip"];
	        this.kPct = source["kPct"];
	        this.bbPct = source["bbPct"];
	        this.abPerHr = source["abPerHr"];
	    }
	}
	export class CareerBattingStatsDTO {
	    gamesPlayed: number;
	    gamesBatting: number;
	    atBats: number;
	    runs: number;
	    hits: number;
	    doubles: number;
	    triples: number;
	    homeRuns: number;
	    rbi: number;
	    stolenBases: number;
	    caughtStealing: number;
	    walks: number;
	    strikeouts: number;
	    hitByPitch: number;
	    sacHits: number;
	    sacFlies: number;
	    errors: number;
	    passedBalls: number;
	    ba?: number;
	    obp?: number;
	    slg?: number;
	    ops?: number;
	    iso?: number;
	    babip?: number;
	    kPct?: number;
	    bbPct?: number;
	    abPerHr?: number;
	
	    static createFrom(source: any = {}) {
	        return new CareerBattingStatsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gamesPlayed = source["gamesPlayed"];
	        this.gamesBatting = source["gamesBatting"];
	        this.atBats = source["atBats"];
	        this.runs = source["runs"];
	        this.hits = source["hits"];
	        this.doubles = source["doubles"];
	        this.triples = source["triples"];
	        this.homeRuns = source["homeRuns"];
	        this.rbi = source["rbi"];
	        this.stolenBases = source["stolenBases"];
	        this.caughtStealing = source["caughtStealing"];
	        this.walks = source["walks"];
	        this.strikeouts = source["strikeouts"];
	        this.hitByPitch = source["hitByPitch"];
	        this.sacHits = source["sacHits"];
	        this.sacFlies = source["sacFlies"];
	        this.errors = source["errors"];
	        this.passedBalls = source["passedBalls"];
	        this.ba = source["ba"];
	        this.obp = source["obp"];
	        this.slg = source["slg"];
	        this.ops = source["ops"];
	        this.iso = source["iso"];
	        this.babip = source["babip"];
	        this.kPct = source["kPct"];
	        this.bbPct = source["bbPct"];
	        this.abPerHr = source["abPerHr"];
	    }
	}
	export class CareerLeaderDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    statValue: number;
	    seasonsPlayed: number;
	
	    static createFrom(source: any = {}) {
	        return new CareerLeaderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.statValue = source["statValue"];
	        this.seasonsPlayed = source["seasonsPlayed"];
	    }
	}
	export class CareerLeadersDTO {
	    hr: CareerLeaderDTO[];
	    hits: CareerLeaderDTO[];
	    rbi: CareerLeaderDTO[];
	    wins: CareerLeaderDTO[];
	    strikeouts: CareerLeaderDTO[];
	    saves: CareerLeaderDTO[];
	
	    static createFrom(source: any = {}) {
	        return new CareerLeadersDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hr = this.convertValues(source["hr"], CareerLeaderDTO);
	        this.hits = this.convertValues(source["hits"], CareerLeaderDTO);
	        this.rbi = this.convertValues(source["rbi"], CareerLeaderDTO);
	        this.wins = this.convertValues(source["wins"], CareerLeaderDTO);
	        this.strikeouts = this.convertValues(source["strikeouts"], CareerLeaderDTO);
	        this.saves = this.convertValues(source["saves"], CareerLeaderDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CareerPitchingStatsDTO {
	    wins: number;
	    losses: number;
	    games: number;
	    gamesStarted: number;
	    completeGames: number;
	    shutouts: number;
	    saves: number;
	    outsPitched: number;
	    hitsAllowed: number;
	    earnedRuns: number;
	    homeRunsAllowed: number;
	    walks: number;
	    strikeouts: number;
	    hitBatters: number;
	    battersFaced: number;
	    gamesFinished: number;
	    runsAllowed: number;
	    wildPitches: number;
	    totalPitches: number;
	    era?: number;
	    whip?: number;
	    k9?: number;
	    bb9?: number;
	    h9?: number;
	    hr9?: number;
	    kPerBb?: number;
	    kPct?: number;
	    winPct?: number;
	    pPerIp?: number;
	
	    static createFrom(source: any = {}) {
	        return new CareerPitchingStatsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.games = source["games"];
	        this.gamesStarted = source["gamesStarted"];
	        this.completeGames = source["completeGames"];
	        this.shutouts = source["shutouts"];
	        this.saves = source["saves"];
	        this.outsPitched = source["outsPitched"];
	        this.hitsAllowed = source["hitsAllowed"];
	        this.earnedRuns = source["earnedRuns"];
	        this.homeRunsAllowed = source["homeRunsAllowed"];
	        this.walks = source["walks"];
	        this.strikeouts = source["strikeouts"];
	        this.hitBatters = source["hitBatters"];
	        this.battersFaced = source["battersFaced"];
	        this.gamesFinished = source["gamesFinished"];
	        this.runsAllowed = source["runsAllowed"];
	        this.wildPitches = source["wildPitches"];
	        this.totalPitches = source["totalPitches"];
	        this.era = source["era"];
	        this.whip = source["whip"];
	        this.k9 = source["k9"];
	        this.bb9 = source["bb9"];
	        this.h9 = source["h9"];
	        this.hr9 = source["hr9"];
	        this.kPerBb = source["kPerBb"];
	        this.kPct = source["kPct"];
	        this.winPct = source["winPct"];
	        this.pPerIp = source["pPerIp"];
	    }
	}
	export class FranchiseDTO {
	    id: string;
	    name: string;
	    gameVersion: string;
	    saveFilePath: string;
	    lastSynced: string;
	    lastSeason: number;
	
	    static createFrom(source: any = {}) {
	        return new FranchiseDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.gameVersion = source["gameVersion"];
	        this.saveFilePath = source["saveFilePath"];
	        this.lastSynced = source["lastSynced"];
	        this.lastSeason = source["lastSeason"];
	    }
	}
	export class LeaderboardFiltersDTO {
	    isPlayoffs: boolean;
	    onlyHallOfFamers: boolean;
	    position: string;
	    batHand: string;
	    throwHand: string;
	    chemistryType: string;
	    seasonStart: number;
	    seasonEnd: number;
	
	    static createFrom(source: any = {}) {
	        return new LeaderboardFiltersDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isPlayoffs = source["isPlayoffs"];
	        this.onlyHallOfFamers = source["onlyHallOfFamers"];
	        this.position = source["position"];
	        this.batHand = source["batHand"];
	        this.throwHand = source["throwHand"];
	        this.chemistryType = source["chemistryType"];
	        this.seasonStart = source["seasonStart"];
	        this.seasonEnd = source["seasonEnd"];
	    }
	}
	export class PitchingLeaderRowDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    isHallOfFamer: boolean;
	    seasonsPlayed: number;
	    seasonNum: number;
	    teamName: string;
	    age: number;
	    pitcherRole: string;
	    throwHand: string;
	    chemistryType: string;
	    wins: number;
	    losses: number;
	    games: number;
	    gamesStarted: number;
	    completeGames: number;
	    shutouts: number;
	    saves: number;
	    outsPitched: number;
	    hitsAllowed: number;
	    earnedRuns: number;
	    homeRunsAllowed: number;
	    walks: number;
	    strikeouts: number;
	    hitBatters: number;
	    battersFaced: number;
	    gamesFinished: number;
	    runsAllowed: number;
	    wildPitches: number;
	    totalPitches: number;
	    era?: number;
	    whip?: number;
	    k9?: number;
	    bb9?: number;
	    h9?: number;
	    hr9?: number;
	    kPerBb?: number;
	    kPct?: number;
	    winPct?: number;
	    pPerIp?: number;
	
	    static createFrom(source: any = {}) {
	        return new PitchingLeaderRowDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.isHallOfFamer = source["isHallOfFamer"];
	        this.seasonsPlayed = source["seasonsPlayed"];
	        this.seasonNum = source["seasonNum"];
	        this.teamName = source["teamName"];
	        this.age = source["age"];
	        this.pitcherRole = source["pitcherRole"];
	        this.throwHand = source["throwHand"];
	        this.chemistryType = source["chemistryType"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.games = source["games"];
	        this.gamesStarted = source["gamesStarted"];
	        this.completeGames = source["completeGames"];
	        this.shutouts = source["shutouts"];
	        this.saves = source["saves"];
	        this.outsPitched = source["outsPitched"];
	        this.hitsAllowed = source["hitsAllowed"];
	        this.earnedRuns = source["earnedRuns"];
	        this.homeRunsAllowed = source["homeRunsAllowed"];
	        this.walks = source["walks"];
	        this.strikeouts = source["strikeouts"];
	        this.hitBatters = source["hitBatters"];
	        this.battersFaced = source["battersFaced"];
	        this.gamesFinished = source["gamesFinished"];
	        this.runsAllowed = source["runsAllowed"];
	        this.wildPitches = source["wildPitches"];
	        this.totalPitches = source["totalPitches"];
	        this.era = source["era"];
	        this.whip = source["whip"];
	        this.k9 = source["k9"];
	        this.bb9 = source["bb9"];
	        this.h9 = source["h9"];
	        this.hr9 = source["hr9"];
	        this.kPerBb = source["kPerBb"];
	        this.kPct = source["kPct"];
	        this.winPct = source["winPct"];
	        this.pPerIp = source["pPerIp"];
	    }
	}
	export class PlayerCareerDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    isHallOfFamer: boolean;
	    batting?: CareerBattingStatsDTO;
	    pitching?: CareerPitchingStatsDTO;
	
	    static createFrom(source: any = {}) {
	        return new PlayerCareerDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.isHallOfFamer = source["isHallOfFamer"];
	        this.batting = this.convertValues(source["batting"], CareerBattingStatsDTO);
	        this.pitching = this.convertValues(source["pitching"], CareerPitchingStatsDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlayerSearchResultDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    isHallOfFamer: boolean;
	    seasonsPlayed: number;
	    firstSeason: number;
	    lastSeason: number;
	
	    static createFrom(source: any = {}) {
	        return new PlayerSearchResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.isHallOfFamer = source["isHallOfFamer"];
	        this.seasonsPlayed = source["seasonsPlayed"];
	        this.firstSeason = source["firstSeason"];
	        this.lastSeason = source["lastSeason"];
	    }
	}
	export class PlayerSeasonLogDTO {
	    seasonNum: number;
	    seasonId: number;
	    teamName: string;
	    age: number;
	    salary: number;
	    primaryPosition: string;
	    secondaryPosition: string;
	    pitcherRole: string;
	    batHand: string;
	    throwHand: string;
	    chemistryType: string;
	    traitsJson: string;
	    pitchesJson: string;
	    power: number;
	    contact: number;
	    speed: number;
	    fielding: number;
	    arm: number;
	    velocity: number;
	    junk: number;
	    accuracy: number;
	    batting?: CareerBattingStatsDTO;
	    pitching?: CareerPitchingStatsDTO;
	    playoffBatting?: CareerBattingStatsDTO;
	    playoffPitching?: CareerPitchingStatsDTO;
	
	    static createFrom(source: any = {}) {
	        return new PlayerSeasonLogDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seasonNum = source["seasonNum"];
	        this.seasonId = source["seasonId"];
	        this.teamName = source["teamName"];
	        this.age = source["age"];
	        this.salary = source["salary"];
	        this.primaryPosition = source["primaryPosition"];
	        this.secondaryPosition = source["secondaryPosition"];
	        this.pitcherRole = source["pitcherRole"];
	        this.batHand = source["batHand"];
	        this.throwHand = source["throwHand"];
	        this.chemistryType = source["chemistryType"];
	        this.traitsJson = source["traitsJson"];
	        this.pitchesJson = source["pitchesJson"];
	        this.power = source["power"];
	        this.contact = source["contact"];
	        this.speed = source["speed"];
	        this.fielding = source["fielding"];
	        this.arm = source["arm"];
	        this.velocity = source["velocity"];
	        this.junk = source["junk"];
	        this.accuracy = source["accuracy"];
	        this.batting = this.convertValues(source["batting"], CareerBattingStatsDTO);
	        this.pitching = this.convertValues(source["pitching"], CareerPitchingStatsDTO);
	        this.playoffBatting = this.convertValues(source["playoffBatting"], CareerBattingStatsDTO);
	        this.playoffPitching = this.convertValues(source["playoffPitching"], CareerPitchingStatsDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlayoffGameDTO {
	    seriesNumber: number;
	    gameNumber: number;
	    homeTeamHistoryId: number;
	    homeTeamName: string;
	    awayTeamHistoryId: number;
	    awayTeamName: string;
	    homeScore?: number;
	    awayScore?: number;
	    homePitcherName: string;
	    awayPitcherName: string;
	
	    static createFrom(source: any = {}) {
	        return new PlayoffGameDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seriesNumber = source["seriesNumber"];
	        this.gameNumber = source["gameNumber"];
	        this.homeTeamHistoryId = source["homeTeamHistoryId"];
	        this.homeTeamName = source["homeTeamName"];
	        this.awayTeamHistoryId = source["awayTeamHistoryId"];
	        this.awayTeamName = source["awayTeamName"];
	        this.homeScore = source["homeScore"];
	        this.awayScore = source["awayScore"];
	        this.homePitcherName = source["homePitcherName"];
	        this.awayPitcherName = source["awayPitcherName"];
	    }
	}
	export class RosterPlayerDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    isHallOfFamer: boolean;
	    age: number;
	    salary: number;
	    primaryPosition: string;
	    secondaryPosition: string;
	    pitcherRole: string;
	    batHand: string;
	    throwHand: string;
	    chemistryType: string;
	    traitsJson: string;
	    pitchesJson: string;
	    power: number;
	    contact: number;
	    speed: number;
	    fielding: number;
	    arm: number;
	    velocity: number;
	    junk: number;
	    accuracy: number;
	    batting?: CareerBattingStatsDTO;
	    pitching?: CareerPitchingStatsDTO;
	
	    static createFrom(source: any = {}) {
	        return new RosterPlayerDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.isHallOfFamer = source["isHallOfFamer"];
	        this.age = source["age"];
	        this.salary = source["salary"];
	        this.primaryPosition = source["primaryPosition"];
	        this.secondaryPosition = source["secondaryPosition"];
	        this.pitcherRole = source["pitcherRole"];
	        this.batHand = source["batHand"];
	        this.throwHand = source["throwHand"];
	        this.chemistryType = source["chemistryType"];
	        this.traitsJson = source["traitsJson"];
	        this.pitchesJson = source["pitchesJson"];
	        this.power = source["power"];
	        this.contact = source["contact"];
	        this.speed = source["speed"];
	        this.fielding = source["fielding"];
	        this.arm = source["arm"];
	        this.velocity = source["velocity"];
	        this.junk = source["junk"];
	        this.accuracy = source["accuracy"];
	        this.batting = this.convertValues(source["batting"], CareerBattingStatsDTO);
	        this.pitching = this.convertValues(source["pitching"], CareerPitchingStatsDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveFileCandidateDTO {
	    path: string;
	    gameVersion: string;
	    leagueName: string;
	    numSeasons: number;
	    mode: string;
	    isFranchise: boolean;
	    playerTeamName: string;
	    leagueGUID: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveFileCandidateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.gameVersion = source["gameVersion"];
	        this.leagueName = source["leagueName"];
	        this.numSeasons = source["numSeasons"];
	        this.mode = source["mode"];
	        this.isFranchise = source["isFranchise"];
	        this.playerTeamName = source["playerTeamName"];
	        this.leagueGUID = source["leagueGUID"];
	    }
	}
	export class ScheduleGameDTO {
	    gameNumber: number;
	    day: number;
	    homeTeamHistoryId: number;
	    homeTeamName: string;
	    awayTeamHistoryId: number;
	    awayTeamName: string;
	    homeScore?: number;
	    awayScore?: number;
	    homePitcherName: string;
	    awayPitcherName: string;
	
	    static createFrom(source: any = {}) {
	        return new ScheduleGameDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gameNumber = source["gameNumber"];
	        this.day = source["day"];
	        this.homeTeamHistoryId = source["homeTeamHistoryId"];
	        this.homeTeamName = source["homeTeamName"];
	        this.awayTeamHistoryId = source["awayTeamHistoryId"];
	        this.awayTeamName = source["awayTeamName"];
	        this.homeScore = source["homeScore"];
	        this.awayScore = source["awayScore"];
	        this.homePitcherName = source["homePitcherName"];
	        this.awayPitcherName = source["awayPitcherName"];
	    }
	}
	export class SeasonSummaryDTO {
	    id: number;
	    seasonNum: number;
	    numGames: number;
	    importedAt: string;
	    championTeamName: string;
	    championHistoryId?: number;
	
	    static createFrom(source: any = {}) {
	        return new SeasonSummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.seasonNum = source["seasonNum"];
	        this.numGames = source["numGames"];
	        this.importedAt = source["importedAt"];
	        this.championTeamName = source["championTeamName"];
	        this.championHistoryId = source["championHistoryId"];
	    }
	}
	export class StatLeaderDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    teamName: string;
	    statValue: number;
	
	    static createFrom(source: any = {}) {
	        return new StatLeaderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.teamName = source["teamName"];
	        this.statValue = source["statValue"];
	    }
	}
	export class StatLeadersDTO {
	    seasonNum: number;
	    ba?: StatLeaderDTO;
	    hr?: StatLeaderDTO;
	    rbi?: StatLeaderDTO;
	    era?: StatLeaderDTO;
	    wins?: StatLeaderDTO;
	    strikeouts?: StatLeaderDTO;
	
	    static createFrom(source: any = {}) {
	        return new StatLeadersDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seasonNum = source["seasonNum"];
	        this.ba = this.convertValues(source["ba"], StatLeaderDTO);
	        this.hr = this.convertValues(source["hr"], StatLeaderDTO);
	        this.rbi = this.convertValues(source["rbi"], StatLeaderDTO);
	        this.era = this.convertValues(source["era"], StatLeaderDTO);
	        this.wins = this.convertValues(source["wins"], StatLeaderDTO);
	        this.strikeouts = this.convertValues(source["strikeouts"], StatLeaderDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SyncSeasonResult {
	    seasonId: number;
	    seasonNum: number;
	    players: number;
	    teams: number;
	    games: number;
	    playoffGames: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncSeasonResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seasonId = source["seasonId"];
	        this.seasonNum = source["seasonNum"];
	        this.players = source["players"];
	        this.teams = source["teams"];
	        this.games = source["games"];
	        this.playoffGames = source["playoffGames"];
	    }
	}
	export class TeamSeasonSummaryDTO {
	    historyId: number;
	    seasonId: number;
	    seasonNum: number;
	    teamName: string;
	    divisionName: string;
	    conferenceName: string;
	    wins: number;
	    losses: number;
	    winPct: number;
	    gamesBack: number;
	    runsFor: number;
	    runsAgainst: number;
	    budget: number;
	    payroll: number;
	    playoffSeed?: number;
	    playoffWins?: number;
	    playoffLosses?: number;
	    playoffRunsFor?: number;
	    playoffRunsAgainst?: number;
	    totalPower: number;
	    totalContact: number;
	    totalSpeed: number;
	    totalFielding: number;
	    totalArm: number;
	    totalVelocity: number;
	    totalJunk: number;
	    totalAccuracy: number;
	    isChampion: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TeamSeasonSummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.historyId = source["historyId"];
	        this.seasonId = source["seasonId"];
	        this.seasonNum = source["seasonNum"];
	        this.teamName = source["teamName"];
	        this.divisionName = source["divisionName"];
	        this.conferenceName = source["conferenceName"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.winPct = source["winPct"];
	        this.gamesBack = source["gamesBack"];
	        this.runsFor = source["runsFor"];
	        this.runsAgainst = source["runsAgainst"];
	        this.budget = source["budget"];
	        this.payroll = source["payroll"];
	        this.playoffSeed = source["playoffSeed"];
	        this.playoffWins = source["playoffWins"];
	        this.playoffLosses = source["playoffLosses"];
	        this.playoffRunsFor = source["playoffRunsFor"];
	        this.playoffRunsAgainst = source["playoffRunsAgainst"];
	        this.totalPower = source["totalPower"];
	        this.totalContact = source["totalContact"];
	        this.totalSpeed = source["totalSpeed"];
	        this.totalFielding = source["totalFielding"];
	        this.totalArm = source["totalArm"];
	        this.totalVelocity = source["totalVelocity"];
	        this.totalJunk = source["totalJunk"];
	        this.totalAccuracy = source["totalAccuracy"];
	        this.isChampion = source["isChampion"];
	    }
	}
	export class TeamHistoryDTO {
	    teamId: number;
	    gameGuid: string;
	    seasons: TeamSeasonSummaryDTO[];
	
	    static createFrom(source: any = {}) {
	        return new TeamHistoryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.teamId = source["teamId"];
	        this.gameGuid = source["gameGuid"];
	        this.seasons = this.convertValues(source["seasons"], TeamSeasonSummaryDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TeamSearchResultDTO {
	    teamId: number;
	    teamName: string;
	    seasons: number;
	    firstSeason: number;
	    lastSeason: number;
	
	    static createFrom(source: any = {}) {
	        return new TeamSearchResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.teamId = source["teamId"];
	        this.teamName = source["teamName"];
	        this.seasons = source["seasons"];
	        this.firstSeason = source["firstSeason"];
	        this.lastSeason = source["lastSeason"];
	    }
	}
	export class TeamSeasonDetailDTO {
	    team: TeamSeasonSummaryDTO;
	    roster: RosterPlayerDTO[];
	    schedule: ScheduleGameDTO[];
	    playoffs: PlayoffGameDTO[];
	
	    static createFrom(source: any = {}) {
	        return new TeamSeasonDetailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.team = this.convertValues(source["team"], TeamSeasonSummaryDTO);
	        this.roster = this.convertValues(source["roster"], RosterPlayerDTO);
	        this.schedule = this.convertValues(source["schedule"], ScheduleGameDTO);
	        this.playoffs = this.convertValues(source["playoffs"], PlayoffGameDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TeamSeasonListDTO {
	    seasonNum: number;
	    historyId: number;
	    teamId: number;
	    teamName: string;
	    conferenceName: string;
	    divisionName: string;
	    wins: number;
	    losses: number;
	    winPct: number;
	    runsFor: number;
	    runsAgainst: number;
	    playoffSeed?: number;
	    playoffWins?: number;
	    playoffLosses?: number;
	    isChampion: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TeamSeasonListDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seasonNum = source["seasonNum"];
	        this.historyId = source["historyId"];
	        this.teamId = source["teamId"];
	        this.teamName = source["teamName"];
	        this.conferenceName = source["conferenceName"];
	        this.divisionName = source["divisionName"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.winPct = source["winPct"];
	        this.runsFor = source["runsFor"];
	        this.runsAgainst = source["runsAgainst"];
	        this.playoffSeed = source["playoffSeed"];
	        this.playoffWins = source["playoffWins"];
	        this.playoffLosses = source["playoffLosses"];
	        this.isChampion = source["isChampion"];
	    }
	}
	
	export class TeamStandingDTO {
	    historyId: number;
	    teamId: number;
	    teamName: string;
	    divisionName: string;
	    conferenceName: string;
	    wins: number;
	    losses: number;
	    winPct: number;
	    gamesBack: number;
	    runsFor: number;
	    runsAgainst: number;
	    runDiff: number;
	    playoffSeed?: number;
	
	    static createFrom(source: any = {}) {
	        return new TeamStandingDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.historyId = source["historyId"];
	        this.teamId = source["teamId"];
	        this.teamName = source["teamName"];
	        this.divisionName = source["divisionName"];
	        this.conferenceName = source["conferenceName"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.winPct = source["winPct"];
	        this.gamesBack = source["gamesBack"];
	        this.runsFor = source["runsFor"];
	        this.runsAgainst = source["runsAgainst"];
	        this.runDiff = source["runDiff"];
	        this.playoffSeed = source["playoffSeed"];
	    }
	}

}

