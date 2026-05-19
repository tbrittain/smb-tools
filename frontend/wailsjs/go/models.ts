export namespace main {
	
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

}

