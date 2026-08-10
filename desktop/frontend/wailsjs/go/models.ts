export namespace application {

	export class ApplicationSummary {
	    id: string;
	    name: string;
	    description: string;
	    category: string;
	    url: string;
	    optional: boolean;
	    dependencies: string[];

	    static createFrom(source: any = {}) {
	        return new ApplicationSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.url = source["url"];
	        this.optional = source["optional"];
	        this.dependencies = source["dependencies"];
	    }
	}
	export class EnvironmentStatus {
	    platform: string;
	    architecture: string;
	    runtime: runtime.Status;

	    static createFrom(source: any = {}) {
	        return new EnvironmentStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.architecture = source["architecture"];
	        this.runtime = this.convertValues(source["runtime"], runtime.Status);
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
	export class InstallationItem {
	    applicationId: string;
	    status: runtime.ContainerStatus;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new InstallationItem(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.status = this.convertValues(source["status"], runtime.ContainerStatus);
	        this.error = source["error"];
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
	export class InstallationResult {
	    items: InstallationItem[];
	    complete: boolean;

	    static createFrom(source: any = {}) {
	        return new InstallationResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], InstallationItem);
	        this.complete = source["complete"];
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
	export class SetupStatus {
	    storagePath?: string;
	    applications: string[];
	    canPrepare: boolean;
	    canInstall: boolean;
	    termsVersion: string;
	    termsAccepted: boolean;

	    static createFrom(source: any = {}) {
	        return new SetupStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.storagePath = source["storagePath"];
	        this.applications = source["applications"];
	        this.canPrepare = source["canPrepare"];
	        this.canInstall = source["canInstall"];
	        this.termsVersion = source["termsVersion"];
	        this.termsAccepted = source["termsAccepted"];
	    }
	}

}

export namespace runtime {

	export class ContainerStatus {
	    applicationId: string;
	    state: string;
	    health?: string;
	    image?: string;

	    static createFrom(source: any = {}) {
	        return new ContainerStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.state = source["state"];
	        this.health = source["health"];
	        this.image = source["image"];
	    }
	}
	export class Status {
	    provider: string;
	    state: string;
	    version?: string;
	    technicalDetail?: string;

	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.state = source["state"];
	        this.version = source["version"];
	        this.technicalDetail = source["technicalDetail"];
	    }
	}

}

export namespace storage {

	export class LayoutStatus {
	    rootPath: string;
	    directories: string[];

	    static createFrom(source: any = {}) {
	        return new LayoutStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootPath = source["rootPath"];
	        this.directories = source["directories"];
	    }
	}
	export class Status {
	    path: string;
	    state: string;
	    writable: boolean;
	    hardlinks: boolean;
	    availableBytes?: number;
	    technicalDetail?: string;

	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.state = source["state"];
	        this.writable = source["writable"];
	        this.hardlinks = source["hardlinks"];
	        this.availableBytes = source["availableBytes"];
	        this.technicalDetail = source["technicalDetail"];
	    }
	}

}

