export namespace config {
	
	export class D2PTConfig {
	    period: string;
	
	    static createFrom(source: any = {}) {
	        return new D2PTConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.period = source["period"];
	    }
	}
	export class FileConfig {
	    filePath: string;
	    enabled: boolean;
	    lastUpdateTimestampMillis: number;
	    lastUpdateErrorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new FileConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.enabled = source["enabled"];
	        this.lastUpdateTimestampMillis = source["lastUpdateTimestampMillis"];
	        this.lastUpdateErrorMessage = source["lastUpdateErrorMessage"];
	    }
	}
	export class PositionConfig {
	    id: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PositionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	    }
	}
	export class SteamAccountConfig {
	    steamId64: string;
	    enabled: boolean;
	    lastUpdateTimestampMillis: number;
	    lastUpdateErrorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new SteamAccountConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.steamId64 = source["steamId64"];
	        this.enabled = source["enabled"];
	        this.lastUpdateTimestampMillis = source["lastUpdateTimestampMillis"];
	        this.lastUpdateErrorMessage = source["lastUpdateErrorMessage"];
	    }
	}
	export class SteamConfig {
	    steamPath: string;
	    autoEnableNewAccounts: boolean;
	    accounts: SteamAccountConfig[];
	
	    static createFrom(source: any = {}) {
	        return new SteamConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.steamPath = source["steamPath"];
	        this.autoEnableNewAccounts = source["autoEnableNewAccounts"];
	        this.accounts = this.convertValues(source["accounts"], SteamAccountConfig);
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

}

export namespace main {
	
	export class AppUpdateState {
	    currentVersion: string;
	    latestVersion: string;
	    lastCheckTimeMillis: number;
	    updateAvailable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppUpdateState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.lastCheckTimeMillis = source["lastCheckTimeMillis"];
	        this.updateAvailable = source["updateAvailable"];
	    }
	}

}

export namespace steam {
	
	export class SteamAccountView {
	    steamId64: string;
	    steamId3: string;
	    accountName: string;
	    personaName: string;
	    avatarBase64: string;
	    enabled: boolean;
	    lastUpdateTimestampMillis: number;
	    lastUpdateErrorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new SteamAccountView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.steamId64 = source["steamId64"];
	        this.steamId3 = source["steamId3"];
	        this.accountName = source["accountName"];
	        this.personaName = source["personaName"];
	        this.avatarBase64 = source["avatarBase64"];
	        this.enabled = source["enabled"];
	        this.lastUpdateTimestampMillis = source["lastUpdateTimestampMillis"];
	        this.lastUpdateErrorMessage = source["lastUpdateErrorMessage"];
	    }
	}

}

