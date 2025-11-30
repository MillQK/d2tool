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
	export class HomeState {
	    lastUpdateTime: string;
	    lastUpdateError: string;
	    isUpdating: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HomeState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastUpdateTime = source["lastUpdateTime"];
	        this.lastUpdateError = source["lastUpdateError"];
	        this.isUpdating = source["isUpdating"];
	    }
	}

}

