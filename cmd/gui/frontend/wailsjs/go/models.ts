export namespace clean {
	
	export class PreviewEntry {
	    Path: string;
	    Size: number;
	    Type: string;
	
	    static createFrom(source: any = {}) {
	        return new PreviewEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Size = source["Size"];
	        this.Type = source["Type"];
	    }
	}

}

export namespace platform {
	
	export class AppInfo {
	    Name: string;
	    BundleID: string;
	    Path: string;
	    UninstallString: string;
	    Size: number;
	    // Go type: time
	    LastUsed: any;
	    IsRunning: boolean;
	    IsBackground: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.BundleID = source["BundleID"];
	        this.Path = source["Path"];
	        this.UninstallString = source["UninstallString"];
	        this.Size = source["Size"];
	        this.LastUsed = this.convertValues(source["LastUsed"], null);
	        this.IsRunning = source["IsRunning"];
	        this.IsBackground = source["IsBackground"];
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

