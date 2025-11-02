export namespace main {
	
	export class AlertConfig {
	    maps: string[];
	    gamemodes: string[];
	    regions: string[];
	    minPlayers: number;
	
	    static createFrom(source: any = {}) {
	        return new AlertConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maps = source["maps"];
	        this.gamemodes = source["gamemodes"];
	        this.regions = source["regions"];
	        this.minPlayers = source["minPlayers"];
	    }
	}
	export class FilterLists {
	    gamemodes: string[];
	    regions: string[];
	
	    static createFrom(source: any = {}) {
	        return new FilterLists(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gamemodes = source["gamemodes"];
	        this.regions = source["regions"];
	    }
	}

}

