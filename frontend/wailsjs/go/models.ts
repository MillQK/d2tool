export namespace config {
	
	export class FileConfig {
	    filePath: string;
	    enabled: boolean;
	    attributes: Record<string, string>;
	    lastUpdateTimestampMillis: number;
	    lastUpdateErrorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new FileConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.enabled = source["enabled"];
	        this.attributes = source["attributes"];
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

}

export namespace main {
	
	export class AppUpdateState {
	    currentVersion: string;
	    latestVersion: string;
	    lastCheckTime: string;
	    updateAvailable: boolean;
	    isCheckingForUpdate: boolean;
	    isDownloadingUpdate: boolean;
	    autoUpdateEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppUpdateState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.lastCheckTime = source["lastCheckTime"];
	        this.updateAvailable = source["updateAvailable"];
	        this.isCheckingForUpdate = source["isCheckingForUpdate"];
	        this.isDownloadingUpdate = source["isDownloadingUpdate"];
	        this.autoUpdateEnabled = source["autoUpdateEnabled"];
	    }
	}

}

