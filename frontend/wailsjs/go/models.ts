export namespace main {

  export class FranchiseDTO {
    id: string
    name: string
    gameVersion: string
    saveFilePath: string
    lastSynced: string
    lastSeason: number

    static createFrom(source: any = {}) { return new FranchiseDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.id = source['id']
      this.name = source['name']
      this.gameVersion = source['gameVersion']
      this.saveFilePath = source['saveFilePath']
      this.lastSynced = source['lastSynced']
      this.lastSeason = source['lastSeason']
    }
  }

  export class SyncSeasonResult {
    seasonId: number
    seasonNum: number
    players: number
    teams: number
    games: number
    playoffGames: number

    static createFrom(source: any = {}) { return new SyncSeasonResult(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.seasonId = source['seasonId']
      this.seasonNum = source['seasonNum']
      this.players = source['players']
      this.teams = source['teams']
      this.games = source['games']
      this.playoffGames = source['playoffGames']
    }
  }

  // ── Phase 5 types ────────────────────────────────────────────────────────────

  export class SeasonSummaryDTO {
    id: number
    seasonNum: number
    numGames: number
    importedAt: string
    championTeamName: string
    championHistoryId: number | null

    static createFrom(source: any = {}) { return new SeasonSummaryDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.id = source['id']
      this.seasonNum = source['seasonNum']
      this.numGames = source['numGames']
      this.importedAt = source['importedAt']
      this.championTeamName = source['championTeamName']
      this.championHistoryId = source['championHistoryId']
    }
  }

  export class TeamStandingDTO {
    historyId: number
    teamId: number
    teamName: string
    divisionName: string
    conferenceName: string
    wins: number
    losses: number
    winPct: number
    gamesBack: number
    runsFor: number
    runsAgainst: number
    runDiff: number
    playoffSeed: number | null

    static createFrom(source: any = {}) { return new TeamStandingDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.historyId = source['historyId']
      this.teamId = source['teamId']
      this.teamName = source['teamName']
      this.divisionName = source['divisionName']
      this.conferenceName = source['conferenceName']
      this.wins = source['wins']
      this.losses = source['losses']
      this.winPct = source['winPct']
      this.gamesBack = source['gamesBack']
      this.runsFor = source['runsFor']
      this.runsAgainst = source['runsAgainst']
      this.runDiff = source['runDiff']
      this.playoffSeed = source['playoffSeed']
    }
  }

  export class StatLeaderDTO {
    playerId: number
    firstName: string
    lastName: string
    teamName: string
    statValue: number

    static createFrom(source: any = {}) { return new StatLeaderDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.playerId = source['playerId']
      this.firstName = source['firstName']
      this.lastName = source['lastName']
      this.teamName = source['teamName']
      this.statValue = source['statValue']
    }
  }

  export class StatLeadersDTO {
    seasonNum: number
    ba: StatLeaderDTO | null
    hr: StatLeaderDTO | null
    rbi: StatLeaderDTO | null
    era: StatLeaderDTO | null
    wins: StatLeaderDTO | null
    strikeouts: StatLeaderDTO | null

    static createFrom(source: any = {}) { return new StatLeadersDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.seasonNum = source['seasonNum']
      this.ba = source['ba'] ? new StatLeaderDTO(source['ba']) : null
      this.hr = source['hr'] ? new StatLeaderDTO(source['hr']) : null
      this.rbi = source['rbi'] ? new StatLeaderDTO(source['rbi']) : null
      this.era = source['era'] ? new StatLeaderDTO(source['era']) : null
      this.wins = source['wins'] ? new StatLeaderDTO(source['wins']) : null
      this.strikeouts = source['strikeouts'] ? new StatLeaderDTO(source['strikeouts']) : null
    }
  }

  export class CareerLeaderDTO {
    playerId: number
    firstName: string
    lastName: string
    statValue: number
    seasonsPlayed: number

    static createFrom(source: any = {}) { return new CareerLeaderDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.playerId = source['playerId']
      this.firstName = source['firstName']
      this.lastName = source['lastName']
      this.statValue = source['statValue']
      this.seasonsPlayed = source['seasonsPlayed']
    }
  }

  export class CareerLeadersDTO {
    hr: CareerLeaderDTO[]
    hits: CareerLeaderDTO[]
    rbi: CareerLeaderDTO[]
    wins: CareerLeaderDTO[]
    strikeouts: CareerLeaderDTO[]
    saves: CareerLeaderDTO[]

    static createFrom(source: any = {}) { return new CareerLeadersDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.hr = (source['hr'] ?? []).map((x: any) => new CareerLeaderDTO(x))
      this.hits = (source['hits'] ?? []).map((x: any) => new CareerLeaderDTO(x))
      this.rbi = (source['rbi'] ?? []).map((x: any) => new CareerLeaderDTO(x))
      this.wins = (source['wins'] ?? []).map((x: any) => new CareerLeaderDTO(x))
      this.strikeouts = (source['strikeouts'] ?? []).map((x: any) => new CareerLeaderDTO(x))
      this.saves = (source['saves'] ?? []).map((x: any) => new CareerLeaderDTO(x))
    }
  }

  export class CareerBattingStatsDTO {
    gamesPlayed: number
    gamesBatting: number
    atBats: number
    runs: number
    hits: number
    doubles: number
    triples: number
    homeRuns: number
    rbi: number
    stolenBases: number
    caughtStealing: number
    walks: number
    strikeouts: number
    hitByPitch: number
    sacHits: number
    sacFlies: number
    errors: number
    passedBalls: number
    ba: number | null
    obp: number | null
    slg: number | null
    ops: number | null
    iso: number | null
    babip: number | null
    kPct: number | null
    bbPct: number | null
    abPerHr: number | null

    static createFrom(source: any = {}) { return new CareerBattingStatsDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.gamesPlayed = source['gamesPlayed']
      this.gamesBatting = source['gamesBatting']
      this.atBats = source['atBats']
      this.runs = source['runs']
      this.hits = source['hits']
      this.doubles = source['doubles']
      this.triples = source['triples']
      this.homeRuns = source['homeRuns']
      this.rbi = source['rbi']
      this.stolenBases = source['stolenBases']
      this.caughtStealing = source['caughtStealing']
      this.walks = source['walks']
      this.strikeouts = source['strikeouts']
      this.hitByPitch = source['hitByPitch']
      this.sacHits = source['sacHits']
      this.sacFlies = source['sacFlies']
      this.errors = source['errors']
      this.passedBalls = source['passedBalls']
      this.ba = source['ba']
      this.obp = source['obp']
      this.slg = source['slg']
      this.ops = source['ops']
      this.iso = source['iso']
      this.babip = source['babip']
      this.kPct = source['kPct']
      this.bbPct = source['bbPct']
      this.abPerHr = source['abPerHr']
    }
  }

  export class CareerPitchingStatsDTO {
    wins: number
    losses: number
    games: number
    gamesStarted: number
    completeGames: number
    shutouts: number
    saves: number
    outsPitched: number
    hitsAllowed: number
    earnedRuns: number
    homeRunsAllowed: number
    walks: number
    strikeouts: number
    hitBatters: number
    battersFaced: number
    gamesFinished: number
    runsAllowed: number
    wildPitches: number
    totalPitches: number
    era: number | null
    whip: number | null
    k9: number | null
    bb9: number | null
    h9: number | null
    hr9: number | null
    kPerBb: number | null
    kPct: number | null
    winPct: number | null
    pPerIp: number | null

    static createFrom(source: any = {}) { return new CareerPitchingStatsDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.wins = source['wins']
      this.losses = source['losses']
      this.games = source['games']
      this.gamesStarted = source['gamesStarted']
      this.completeGames = source['completeGames']
      this.shutouts = source['shutouts']
      this.saves = source['saves']
      this.outsPitched = source['outsPitched']
      this.hitsAllowed = source['hitsAllowed']
      this.earnedRuns = source['earnedRuns']
      this.homeRunsAllowed = source['homeRunsAllowed']
      this.walks = source['walks']
      this.strikeouts = source['strikeouts']
      this.hitBatters = source['hitBatters']
      this.battersFaced = source['battersFaced']
      this.gamesFinished = source['gamesFinished']
      this.runsAllowed = source['runsAllowed']
      this.wildPitches = source['wildPitches']
      this.totalPitches = source['totalPitches']
      this.era = source['era']
      this.whip = source['whip']
      this.k9 = source['k9']
      this.bb9 = source['bb9']
      this.h9 = source['h9']
      this.hr9 = source['hr9']
      this.kPerBb = source['kPerBb']
      this.kPct = source['kPct']
      this.winPct = source['winPct']
      this.pPerIp = source['pPerIp']
    }
  }

  export class PlayerSearchResultDTO {
    playerId: number
    firstName: string
    lastName: string
    isHallOfFamer: boolean
    seasonsPlayed: number
    firstSeason: number
    lastSeason: number

    static createFrom(source: any = {}) { return new PlayerSearchResultDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.playerId = source['playerId']
      this.firstName = source['firstName']
      this.lastName = source['lastName']
      this.isHallOfFamer = source['isHallOfFamer']
      this.seasonsPlayed = source['seasonsPlayed']
      this.firstSeason = source['firstSeason']
      this.lastSeason = source['lastSeason']
    }
  }

  export class PlayerCareerDTO {
    playerId: number
    firstName: string
    lastName: string
    isHallOfFamer: boolean
    batting: CareerBattingStatsDTO | null
    pitching: CareerPitchingStatsDTO | null

    static createFrom(source: any = {}) { return new PlayerCareerDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.playerId = source['playerId']
      this.firstName = source['firstName']
      this.lastName = source['lastName']
      this.isHallOfFamer = source['isHallOfFamer']
      this.batting = source['batting'] ? new CareerBattingStatsDTO(source['batting']) : null
      this.pitching = source['pitching'] ? new CareerPitchingStatsDTO(source['pitching']) : null
    }
  }

  export class PlayerSeasonLogDTO {
    seasonNum: number
    seasonId: number
    teamName: string
    age: number
    salary: number
    primaryPosition: string
    secondaryPosition: string
    pitcherRole: string
    batHand: string
    throwHand: string
    chemistryType: string
    traitsJson: string
    pitchesJson: string
    power: number
    contact: number
    speed: number
    fielding: number
    arm: number
    velocity: number
    junk: number
    accuracy: number
    batting: CareerBattingStatsDTO | null
    pitching: CareerPitchingStatsDTO | null
    playoffBatting: CareerBattingStatsDTO | null
    playoffPitching: CareerPitchingStatsDTO | null

    static createFrom(source: any = {}) { return new PlayerSeasonLogDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.seasonNum = source['seasonNum']
      this.seasonId = source['seasonId']
      this.teamName = source['teamName']
      this.age = source['age']
      this.salary = source['salary']
      this.primaryPosition = source['primaryPosition']
      this.secondaryPosition = source['secondaryPosition']
      this.pitcherRole = source['pitcherRole']
      this.batHand = source['batHand']
      this.throwHand = source['throwHand']
      this.chemistryType = source['chemistryType']
      this.traitsJson = source['traitsJson']
      this.pitchesJson = source['pitchesJson']
      this.power = source['power']
      this.contact = source['contact']
      this.speed = source['speed']
      this.fielding = source['fielding']
      this.arm = source['arm']
      this.velocity = source['velocity']
      this.junk = source['junk']
      this.accuracy = source['accuracy']
      this.batting = source['batting'] ? new CareerBattingStatsDTO(source['batting']) : null
      this.pitching = source['pitching'] ? new CareerPitchingStatsDTO(source['pitching']) : null
      this.playoffBatting = source['playoffBatting'] ? new CareerBattingStatsDTO(source['playoffBatting']) : null
      this.playoffPitching = source['playoffPitching'] ? new CareerPitchingStatsDTO(source['playoffPitching']) : null
    }
  }

  export class TeamSearchResultDTO {
    teamId: number
    teamName: string
    seasons: number
    firstSeason: number
    lastSeason: number

    static createFrom(source: any = {}) { return new TeamSearchResultDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.teamId = source['teamId']
      this.teamName = source['teamName']
      this.seasons = source['seasons']
      this.firstSeason = source['firstSeason']
      this.lastSeason = source['lastSeason']
    }
  }

  export class TeamSeasonSummaryDTO {
    historyId: number
    seasonId: number
    seasonNum: number
    teamName: string
    divisionName: string
    conferenceName: string
    wins: number
    losses: number
    winPct: number
    gamesBack: number
    runsFor: number
    runsAgainst: number
    budget: number
    payroll: number
    playoffSeed: number | null
    playoffWins: number | null
    playoffLosses: number | null
    playoffRunsFor: number | null
    playoffRunsAgainst: number | null
    totalPower: number
    totalContact: number
    totalSpeed: number
    totalFielding: number
    totalArm: number
    totalVelocity: number
    totalJunk: number
    totalAccuracy: number
    isChampion: boolean

    static createFrom(source: any = {}) { return new TeamSeasonSummaryDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.historyId = source['historyId']
      this.seasonId = source['seasonId']
      this.seasonNum = source['seasonNum']
      this.teamName = source['teamName']
      this.divisionName = source['divisionName']
      this.conferenceName = source['conferenceName']
      this.wins = source['wins']
      this.losses = source['losses']
      this.winPct = source['winPct']
      this.gamesBack = source['gamesBack']
      this.runsFor = source['runsFor']
      this.runsAgainst = source['runsAgainst']
      this.budget = source['budget']
      this.payroll = source['payroll']
      this.playoffSeed = source['playoffSeed']
      this.playoffWins = source['playoffWins']
      this.playoffLosses = source['playoffLosses']
      this.playoffRunsFor = source['playoffRunsFor']
      this.playoffRunsAgainst = source['playoffRunsAgainst']
      this.totalPower = source['totalPower']
      this.totalContact = source['totalContact']
      this.totalSpeed = source['totalSpeed']
      this.totalFielding = source['totalFielding']
      this.totalArm = source['totalArm']
      this.totalVelocity = source['totalVelocity']
      this.totalJunk = source['totalJunk']
      this.totalAccuracy = source['totalAccuracy']
      this.isChampion = source['isChampion']
    }
  }

  export class TeamHistoryDTO {
    teamId: number
    gameGuid: string
    seasons: TeamSeasonSummaryDTO[]

    static createFrom(source: any = {}) { return new TeamHistoryDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.teamId = source['teamId']
      this.gameGuid = source['gameGuid']
      this.seasons = (source['seasons'] ?? []).map((x: any) => new TeamSeasonSummaryDTO(x))
    }
  }

  export class TeamSeasonListDTO {
    seasonNum: number
    historyId: number
    teamId: number
    teamName: string
    conferenceName: string
    divisionName: string
    wins: number
    losses: number
    winPct: number
    runsFor: number
    runsAgainst: number
    playoffSeed: number | null
    playoffWins: number | null
    playoffLosses: number | null
    isChampion: boolean

    static createFrom(source: any = {}) { return new TeamSeasonListDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.seasonNum = source['seasonNum']
      this.historyId = source['historyId']
      this.teamId = source['teamId']
      this.teamName = source['teamName']
      this.conferenceName = source['conferenceName']
      this.divisionName = source['divisionName']
      this.wins = source['wins']
      this.losses = source['losses']
      this.winPct = source['winPct']
      this.runsFor = source['runsFor']
      this.runsAgainst = source['runsAgainst']
      this.playoffSeed = source['playoffSeed']
      this.playoffWins = source['playoffWins']
      this.playoffLosses = source['playoffLosses']
      this.isChampion = source['isChampion']
    }
  }

  export class RosterPlayerDTO {
    playerId: number
    firstName: string
    lastName: string
    isHallOfFamer: boolean
    age: number
    salary: number
    primaryPosition: string
    secondaryPosition: string
    pitcherRole: string
    batHand: string
    throwHand: string
    chemistryType: string
    traitsJson: string
    pitchesJson: string
    power: number
    contact: number
    speed: number
    fielding: number
    arm: number
    velocity: number
    junk: number
    accuracy: number
    batting: CareerBattingStatsDTO | null
    pitching: CareerPitchingStatsDTO | null

    static createFrom(source: any = {}) { return new RosterPlayerDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.playerId = source['playerId']
      this.firstName = source['firstName']
      this.lastName = source['lastName']
      this.isHallOfFamer = source['isHallOfFamer']
      this.age = source['age']
      this.salary = source['salary']
      this.primaryPosition = source['primaryPosition']
      this.secondaryPosition = source['secondaryPosition']
      this.pitcherRole = source['pitcherRole']
      this.batHand = source['batHand']
      this.throwHand = source['throwHand']
      this.chemistryType = source['chemistryType']
      this.traitsJson = source['traitsJson']
      this.pitchesJson = source['pitchesJson']
      this.power = source['power']
      this.contact = source['contact']
      this.speed = source['speed']
      this.fielding = source['fielding']
      this.arm = source['arm']
      this.velocity = source['velocity']
      this.junk = source['junk']
      this.accuracy = source['accuracy']
      this.batting = source['batting'] ? new CareerBattingStatsDTO(source['batting']) : null
      this.pitching = source['pitching'] ? new CareerPitchingStatsDTO(source['pitching']) : null
    }
  }

  export class ScheduleGameDTO {
    gameNumber: number
    day: number
    homeTeamHistoryId: number
    homeTeamName: string
    awayTeamHistoryId: number
    awayTeamName: string
    homeScore: number | null
    awayScore: number | null
    homePitcherName: string
    awayPitcherName: string

    static createFrom(source: any = {}) { return new ScheduleGameDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.gameNumber = source['gameNumber']
      this.day = source['day']
      this.homeTeamHistoryId = source['homeTeamHistoryId']
      this.homeTeamName = source['homeTeamName']
      this.awayTeamHistoryId = source['awayTeamHistoryId']
      this.awayTeamName = source['awayTeamName']
      this.homeScore = source['homeScore']
      this.awayScore = source['awayScore']
      this.homePitcherName = source['homePitcherName']
      this.awayPitcherName = source['awayPitcherName']
    }
  }

  export class PlayoffGameDTO {
    seriesNumber: number
    gameNumber: number
    homeTeamHistoryId: number
    homeTeamName: string
    awayTeamHistoryId: number
    awayTeamName: string
    homeScore: number | null
    awayScore: number | null
    homePitcherName: string
    awayPitcherName: string

    static createFrom(source: any = {}) { return new PlayoffGameDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.seriesNumber = source['seriesNumber']
      this.gameNumber = source['gameNumber']
      this.homeTeamHistoryId = source['homeTeamHistoryId']
      this.homeTeamName = source['homeTeamName']
      this.awayTeamHistoryId = source['awayTeamHistoryId']
      this.awayTeamName = source['awayTeamName']
      this.homeScore = source['homeScore']
      this.awayScore = source['awayScore']
      this.homePitcherName = source['homePitcherName']
      this.awayPitcherName = source['awayPitcherName']
    }
  }

  export class TeamSeasonDetailDTO {
    team: TeamSeasonSummaryDTO
    roster: RosterPlayerDTO[]
    schedule: ScheduleGameDTO[]
    playoffs: PlayoffGameDTO[]

    static createFrom(source: any = {}) { return new TeamSeasonDetailDTO(source) }
    constructor(source: any = {}) {
      if ('string' === typeof source) source = JSON.parse(source)
      this.team = new TeamSeasonSummaryDTO(source['team'] ?? {})
      this.roster = (source['roster'] ?? []).map((x: any) => new RosterPlayerDTO(x))
      this.schedule = (source['schedule'] ?? []).map((x: any) => new ScheduleGameDTO(x))
      this.playoffs = (source['playoffs'] ?? []).map((x: any) => new PlayoffGameDTO(x))
    }
  }
}
