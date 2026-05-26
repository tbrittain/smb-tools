export namespace main {
	
	export class AwardDTO {
	    id: number;
	    name: string;
	    originalName: string;
	    importance: number;
	    omitFromGroupings: boolean;
	    isBattingAward: boolean;
	    isPitchingAward: boolean;
	    isFieldingAward: boolean;
	    isPlayoffAward: boolean;
	    isUserAssignable: boolean;
	    isBuiltIn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AwardDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.originalName = source["originalName"];
	        this.importance = source["importance"];
	        this.omitFromGroupings = source["omitFromGroupings"];
	        this.isBattingAward = source["isBattingAward"];
	        this.isPitchingAward = source["isPitchingAward"];
	        this.isFieldingAward = source["isFieldingAward"];
	        this.isPlayoffAward = source["isPlayoffAward"];
	        this.isUserAssignable = source["isUserAssignable"];
	        this.isBuiltIn = source["isBuiltIn"];
	    }
	}
	export class BattingCandidateDTO {
	    playerSeasonId: number;
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    teamName: string;
	    primaryPosition: string;
	    pitcherRole: string;
	    atBats: number;
	    hits: number;
	    homeRuns: number;
	    rbi: number;
	    walks: number;
	    runs: number;
	    stolenBases: number;
	    strikeouts: number;
	    doubles: number;
	    triples: number;
	    ba: number;
	    obp: number;
	    slg: number;
	    ops: number;
	    isChampionTeam: boolean;
	    awardIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new BattingCandidateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerSeasonId = source["playerSeasonId"];
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.teamName = source["teamName"];
	        this.primaryPosition = source["primaryPosition"];
	        this.pitcherRole = source["pitcherRole"];
	        this.atBats = source["atBats"];
	        this.hits = source["hits"];
	        this.homeRuns = source["homeRuns"];
	        this.rbi = source["rbi"];
	        this.walks = source["walks"];
	        this.runs = source["runs"];
	        this.stolenBases = source["stolenBases"];
	        this.strikeouts = source["strikeouts"];
	        this.doubles = source["doubles"];
	        this.triples = source["triples"];
	        this.ba = source["ba"];
	        this.obp = source["obp"];
	        this.slg = source["slg"];
	        this.ops = source["ops"];
	        this.isChampionTeam = source["isChampionTeam"];
	        this.awardIds = source["awardIds"];
	    }
	}
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
	    opsPlus?: number;
	    smbWar?: number;
	
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
	        this.opsPlus = source["opsPlus"];
	        this.smbWar = source["smbWar"];
	    }
	}
	export class BattingLeaderPageDTO {
	    rows: BattingLeaderRowDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new BattingLeaderPageDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], BattingLeaderRowDTO);
	        this.total = source["total"];
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
	    opsPlus?: number;
	    smbWar?: number;
	
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
	        this.opsPlus = source["opsPlus"];
	        this.smbWar = source["smbWar"];
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
	    eraPlus?: number;
	    fip?: number;
	    fipMinus?: number;
	    smbWar?: number;
	
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
	        this.eraPlus = source["eraPlus"];
	        this.fip = source["fip"];
	        this.fipMinus = source["fipMinus"];
	        this.smbWar = source["smbWar"];
	    }
	}
	export class FranchiseDTO {
	    id: string;
	    name: string;
	    gameVersion: string;
	    hasActiveSource: boolean;
	    hasLegacySource: boolean;
	    activeSourcePath: string;
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
	        this.hasActiveSource = source["hasActiveSource"];
	        this.hasLegacySource = source["hasLegacySource"];
	        this.activeSourcePath = source["activeSourcePath"];
	        this.lastSynced = source["lastSynced"];
	        this.lastSeason = source["lastSeason"];
	    }
	}
	export class FranchiseSourceDTO {
	    id: number;
	    saveFilePath: string;
	    leagueGUID: string;
	    seasonOffset: number;
	    addedAt: string;
	    isLegacy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FranchiseSourceDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.saveFilePath = source["saveFilePath"];
	        this.leagueGUID = source["leagueGUID"];
	        this.seasonOffset = source["seasonOffset"];
	        this.addedAt = source["addedAt"];
	        this.isLegacy = source["isLegacy"];
	    }
	}
	export class HistoricalTeamDTO {
	    teamId: number;
	    teamName: string;
	    numSeasons: number;
	    firstSeason: number;
	    lastSeason: number;
	    wins: number;
	    losses: number;
	    winPct: number;
	    gamesOver500: number;
	    playoffWins: number;
	    playoffLosses: number;
	    playoffAppearances: number;
	    divisionTitles: number;
	    conferenceTitles: number;
	    championships: number;
	    championshipDrought: number;
	    runsFor: number;
	    runsAgainst: number;
	    totalAB: number;
	    totalHits: number;
	    totalHR: number;
	    numPlayers: number;
	    numHoF: number;
	    ba?: number;
	    era?: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoricalTeamDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.teamId = source["teamId"];
	        this.teamName = source["teamName"];
	        this.numSeasons = source["numSeasons"];
	        this.firstSeason = source["firstSeason"];
	        this.lastSeason = source["lastSeason"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.winPct = source["winPct"];
	        this.gamesOver500 = source["gamesOver500"];
	        this.playoffWins = source["playoffWins"];
	        this.playoffLosses = source["playoffLosses"];
	        this.playoffAppearances = source["playoffAppearances"];
	        this.divisionTitles = source["divisionTitles"];
	        this.conferenceTitles = source["conferenceTitles"];
	        this.championships = source["championships"];
	        this.championshipDrought = source["championshipDrought"];
	        this.runsFor = source["runsFor"];
	        this.runsAgainst = source["runsAgainst"];
	        this.totalAB = source["totalAB"];
	        this.totalHits = source["totalHits"];
	        this.totalHR = source["totalHR"];
	        this.numPlayers = source["numPlayers"];
	        this.numHoF = source["numHoF"];
	        this.ba = source["ba"];
	        this.era = source["era"];
	    }
	}
	export class HoFCandidateDTO {
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    isHallOfFamer: boolean;
	    firstSeason: number;
	    lastSeason: number;
	    seasons: number;
	    hits: number;
	    homeRuns: number;
	    rbi: number;
	    stolenBases: number;
	    atBats: number;
	    walks: number;
	    wins: number;
	    losses: number;
	    outsPitched: number;
	    strikeouts: number;
	    earnedRuns: number;
	
	    static createFrom(source: any = {}) {
	        return new HoFCandidateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.isHallOfFamer = source["isHallOfFamer"];
	        this.firstSeason = source["firstSeason"];
	        this.lastSeason = source["lastSeason"];
	        this.seasons = source["seasons"];
	        this.hits = source["hits"];
	        this.homeRuns = source["homeRuns"];
	        this.rbi = source["rbi"];
	        this.stolenBases = source["stolenBases"];
	        this.atBats = source["atBats"];
	        this.walks = source["walks"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.outsPitched = source["outsPitched"];
	        this.strikeouts = source["strikeouts"];
	        this.earnedRuns = source["earnedRuns"];
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
	    sortField: string;
	    sortDesc: boolean;
	    offset: number;
	    pageSize: number;
	
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
	        this.sortField = source["sortField"];
	        this.sortDesc = source["sortDesc"];
	        this.offset = source["offset"];
	        this.pageSize = source["pageSize"];
	    }
	}
	export class LegacyFranchiseDTO {
	    id: number;
	    name: string;
	    isSmb3: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LegacyFranchiseDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.isSmb3 = source["isSmb3"];
	    }
	}
	export class MigrateLegacyResult {
	    franchiseId: string;
	    franchiseName: string;
	    seasonsMigrated: number;
	    teamsMigrated: number;
	    playersMigrated: number;
	    awardsMigrated: number;
	    logosSkipped: number;
	
	    static createFrom(source: any = {}) {
	        return new MigrateLegacyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.franchiseId = source["franchiseId"];
	        this.franchiseName = source["franchiseName"];
	        this.seasonsMigrated = source["seasonsMigrated"];
	        this.teamsMigrated = source["teamsMigrated"];
	        this.playersMigrated = source["playersMigrated"];
	        this.awardsMigrated = source["awardsMigrated"];
	        this.logosSkipped = source["logosSkipped"];
	    }
	}
	export class PitchingCandidateDTO {
	    playerSeasonId: number;
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    teamName: string;
	    primaryPosition: string;
	    pitcherRole: string;
	    wins: number;
	    losses: number;
	    saves: number;
	    outsPitched: number;
	    hitsAllowed: number;
	    earnedRuns: number;
	    walks: number;
	    strikeouts: number;
	    homeRunsAllowed: number;
	    completeGames: number;
	    shutouts: number;
	    era: number;
	    whip: number;
	    k9: number;
	    bb9: number;
	    h9: number;
	    hr9: number;
	    kPerBb: number;
	    isChampionTeam: boolean;
	    awardIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new PitchingCandidateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerSeasonId = source["playerSeasonId"];
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.teamName = source["teamName"];
	        this.primaryPosition = source["primaryPosition"];
	        this.pitcherRole = source["pitcherRole"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.saves = source["saves"];
	        this.outsPitched = source["outsPitched"];
	        this.hitsAllowed = source["hitsAllowed"];
	        this.earnedRuns = source["earnedRuns"];
	        this.walks = source["walks"];
	        this.strikeouts = source["strikeouts"];
	        this.homeRunsAllowed = source["homeRunsAllowed"];
	        this.completeGames = source["completeGames"];
	        this.shutouts = source["shutouts"];
	        this.era = source["era"];
	        this.whip = source["whip"];
	        this.k9 = source["k9"];
	        this.bb9 = source["bb9"];
	        this.h9 = source["h9"];
	        this.hr9 = source["hr9"];
	        this.kPerBb = source["kPerBb"];
	        this.isChampionTeam = source["isChampionTeam"];
	        this.awardIds = source["awardIds"];
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
	    eraPlus?: number;
	    fip?: number;
	    fipMinus?: number;
	    smbWar?: number;
	
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
	        this.eraPlus = source["eraPlus"];
	        this.fip = source["fip"];
	        this.fipMinus = source["fipMinus"];
	        this.smbWar = source["smbWar"];
	    }
	}
	export class PitchingLeaderPageDTO {
	    rows: PitchingLeaderRowDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new PitchingLeaderPageDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], PitchingLeaderRowDTO);
	        this.total = source["total"];
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
	
	export class PlayerAwardEntryDTO {
	    playerSeasonId: number;
	    awardIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new PlayerAwardEntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerSeasonId = source["playerSeasonId"];
	        this.awardIds = source["awardIds"];
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
	export class TeamRefDTO {
	    teamId: number;
	    teamHistoryId: number;
	    teamName: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new TeamRefDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.teamId = source["teamId"];
	        this.teamHistoryId = source["teamHistoryId"];
	        this.teamName = source["teamName"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class PlayerSeasonLogDTO {
	    seasonNum: number;
	    seasonId: number;
	    teams: TeamRefDTO[];
	    age: number;
	    salary: number;
	    primaryPosition: string;
	    secondaryPosition: string;
	    pitcherRole: string;
	    batHand: string;
	    throwHand: string;
	    chemistryType: string;
	    traits: string[];
	    pitches: string[];
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
	        this.teams = this.convertValues(source["teams"], TeamRefDTO);
	        this.age = source["age"];
	        this.salary = source["salary"];
	        this.primaryPosition = source["primaryPosition"];
	        this.secondaryPosition = source["secondaryPosition"];
	        this.pitcherRole = source["pitcherRole"];
	        this.batHand = source["batHand"];
	        this.throwHand = source["throwHand"];
	        this.chemistryType = source["chemistryType"];
	        this.traits = source["traits"];
	        this.pitches = source["pitches"];
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
	    roundNumber: number;
	    roundLabel: string;
	    gameNumber: number;
	    homeTeamHistoryId: number;
	    homeTeamName: string;
	    homeTeamId: number;
	    awayTeamHistoryId: number;
	    awayTeamName: string;
	    awayTeamId: number;
	    homeScore?: number;
	    awayScore?: number;
	    homePitcherName: string;
	    awayPitcherName: string;
	    homePitcherPlayerId?: number;
	    awayPitcherPlayerId?: number;
	
	    static createFrom(source: any = {}) {
	        return new PlayoffGameDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.roundNumber = source["roundNumber"];
	        this.roundLabel = source["roundLabel"];
	        this.gameNumber = source["gameNumber"];
	        this.homeTeamHistoryId = source["homeTeamHistoryId"];
	        this.homeTeamName = source["homeTeamName"];
	        this.homeTeamId = source["homeTeamId"];
	        this.awayTeamHistoryId = source["awayTeamHistoryId"];
	        this.awayTeamName = source["awayTeamName"];
	        this.awayTeamId = source["awayTeamId"];
	        this.homeScore = source["homeScore"];
	        this.awayScore = source["awayScore"];
	        this.homePitcherName = source["homePitcherName"];
	        this.awayPitcherName = source["awayPitcherName"];
	        this.homePitcherPlayerId = source["homePitcherPlayerId"];
	        this.awayPitcherPlayerId = source["awayPitcherPlayerId"];
	    }
	}
	export class PositionAwardCandidatesDTO {
	    position: string;
	    batters: BattingCandidateDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PositionAwardCandidatesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.position = source["position"];
	        this.batters = this.convertValues(source["batters"], BattingCandidateDTO);
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
	    isOnFinalRoster: boolean;
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
	        this.isOnFinalRoster = source["isOnFinalRoster"];
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
	    teamGameNum: number;
	    gameNumber: number;
	    day: number;
	    homeTeamHistoryId: number;
	    homeTeamName: string;
	    homeTeamId: number;
	    awayTeamHistoryId: number;
	    awayTeamName: string;
	    awayTeamId: number;
	    homeScore?: number;
	    awayScore?: number;
	    homePitcherName: string;
	    awayPitcherName: string;
	    homePitcherPlayerId?: number;
	    awayPitcherPlayerId?: number;
	
	    static createFrom(source: any = {}) {
	        return new ScheduleGameDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.teamGameNum = source["teamGameNum"];
	        this.gameNumber = source["gameNumber"];
	        this.day = source["day"];
	        this.homeTeamHistoryId = source["homeTeamHistoryId"];
	        this.homeTeamName = source["homeTeamName"];
	        this.homeTeamId = source["homeTeamId"];
	        this.awayTeamHistoryId = source["awayTeamHistoryId"];
	        this.awayTeamName = source["awayTeamName"];
	        this.awayTeamId = source["awayTeamId"];
	        this.homeScore = source["homeScore"];
	        this.awayScore = source["awayScore"];
	        this.homePitcherName = source["homePitcherName"];
	        this.awayPitcherName = source["awayPitcherName"];
	        this.homePitcherPlayerId = source["homePitcherPlayerId"];
	        this.awayPitcherPlayerId = source["awayPitcherPlayerId"];
	    }
	}
	export class TeamAwardCandidatesDTO {
	    historyId: number;
	    teamName: string;
	    batters: BattingCandidateDTO[];
	    pitchers: PitchingCandidateDTO[];
	
	    static createFrom(source: any = {}) {
	        return new TeamAwardCandidatesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.historyId = source["historyId"];
	        this.teamName = source["teamName"];
	        this.batters = this.convertValues(source["batters"], BattingCandidateDTO);
	        this.pitchers = this.convertValues(source["pitchers"], PitchingCandidateDTO);
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
	export class SeasonAwardCandidatesDTO {
	    seasonId: number;
	    seasonNum: number;
	    topBatters: BattingCandidateDTO[];
	    topPitchers: PitchingCandidateDTO[];
	    topRookieBatters: BattingCandidateDTO[];
	    topRookiePitchers: PitchingCandidateDTO[];
	    byTeam: TeamAwardCandidatesDTO[];
	    byPosition: PositionAwardCandidatesDTO[];
	    playoffBatters: BattingCandidateDTO[];
	    playoffPitchers: PitchingCandidateDTO[];
	    championBatters: BattingCandidateDTO[];
	    championPitchers: PitchingCandidateDTO[];
	    autoSuggested: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SeasonAwardCandidatesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seasonId = source["seasonId"];
	        this.seasonNum = source["seasonNum"];
	        this.topBatters = this.convertValues(source["topBatters"], BattingCandidateDTO);
	        this.topPitchers = this.convertValues(source["topPitchers"], PitchingCandidateDTO);
	        this.topRookieBatters = this.convertValues(source["topRookieBatters"], BattingCandidateDTO);
	        this.topRookiePitchers = this.convertValues(source["topRookiePitchers"], PitchingCandidateDTO);
	        this.byTeam = this.convertValues(source["byTeam"], TeamAwardCandidatesDTO);
	        this.byPosition = this.convertValues(source["byPosition"], PositionAwardCandidatesDTO);
	        this.playoffBatters = this.convertValues(source["playoffBatters"], BattingCandidateDTO);
	        this.playoffPitchers = this.convertValues(source["playoffPitchers"], PitchingCandidateDTO);
	        this.championBatters = this.convertValues(source["championBatters"], BattingCandidateDTO);
	        this.championPitchers = this.convertValues(source["championPitchers"], PitchingCandidateDTO);
	        this.autoSuggested = source["autoSuggested"];
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
	export class SeasonPlayerAwardRowDTO {
	    playerSeasonId: number;
	    playerId: number;
	    firstName: string;
	    lastName: string;
	    teamName: string;
	    primaryPosition: string;
	    pitcherRole: string;
	    awards: AwardDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SeasonPlayerAwardRowDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerSeasonId = source["playerSeasonId"];
	        this.playerId = source["playerId"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.teamName = source["teamName"];
	        this.primaryPosition = source["primaryPosition"];
	        this.pitcherRole = source["pitcherRole"];
	        this.awards = this.convertValues(source["awards"], AwardDTO);
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
	export class SetPlayerAwardsRequestDTO {
	    playerSeasonId: number;
	    awardIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new SetPlayerAwardsRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.playerSeasonId = source["playerSeasonId"];
	        this.awardIds = source["awardIds"];
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
	export class SubmitSeasonAwardsDTO {
	    seasonId: number;
	    playerAwards: PlayerAwardEntryDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SubmitSeasonAwardsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seasonId = source["seasonId"];
	        this.playerAwards = this.convertValues(source["playerAwards"], PlayerAwardEntryDTO);
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

