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

}

