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
	export class ApplicationUpdateResult {
	    applicationId: string;
	    previousImage: string;
	    approvedImage: string;
	    backup: storage.BackupResult;
	    status: runtime.ContainerStatus;
	    updated: boolean;
	    rolledBack: boolean;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new ApplicationUpdateResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.previousImage = source["previousImage"];
	        this.approvedImage = source["approvedImage"];
	        this.backup = this.convertValues(source["backup"], storage.BackupResult);
	        this.status = this.convertValues(source["status"], runtime.ContainerStatus);
	        this.updated = source["updated"];
	        this.rolledBack = source["rolledBack"];
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
	export class ManagedApplicationStatus {
	    applicationId: string;
	    state: string;
	    health?: string;
	    image?: string;
	    approvedImage?: string;
	    updateAvailable: boolean;
	    technicalDetail?: string;

	    static createFrom(source: any = {}) {
	        return new ManagedApplicationStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.state = source["state"];
	        this.health = source["health"];
	        this.image = source["image"];
	        this.approvedImage = source["approvedImage"];
	        this.updateAvailable = source["updateAvailable"];
	        this.technicalDetail = source["technicalDetail"];
	    }
	}
	export class ServiceAccessStatus {
	    applicationId: string;
	    username: string;
	    available: boolean;

	    static createFrom(source: any = {}) {
	        return new ServiceAccessStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.username = source["username"];
	        this.available = source["available"];
	    }
	}
	export class SetupStatus {
	    storagePath?: string;
	    applications: string[];
	    canPrepare: boolean;
	    canInstall: boolean;
	    termsVersion: string;
	    termsAccepted: boolean;
	    startAtLogin: boolean;
	    startAtLoginSupported: boolean;
	    startAtLoginRequiresApproval: boolean;
	    jellyfinLanEnabled: boolean;

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
	        this.startAtLogin = source["startAtLogin"];
	        this.startAtLoginSupported = source["startAtLoginSupported"];
	        this.startAtLoginRequiresApproval = source["startAtLoginRequiresApproval"];
	        this.jellyfinLanEnabled = source["jellyfinLanEnabled"];
	    }
	}

}

export namespace legal {

	export class LinkSummary {
	    kind: string;
	    label: string;

	    static createFrom(source: any = {}) {
	        return new LinkSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	    }
	}
	export class Notice {
	    id: string;
	    name: string;
	    purpose: string;
	    componentType: string;
	    license: string;
	    copyrightNotice: string;
	    imageMaintainer?: string;
	    approvedImage?: string;
	    affiliationStatement: string;
	    links: LinkSummary[];

	    static createFrom(source: any = {}) {
	        return new Notice(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.purpose = source["purpose"];
	        this.componentType = source["componentType"];
	        this.license = source["license"];
	        this.copyrightNotice = source["copyrightNotice"];
	        this.imageMaintainer = source["imageMaintainer"];
	        this.approvedImage = source["approvedImage"];
	        this.affiliationStatement = source["affiliationStatement"];
	        this.links = this.convertValues(source["links"], LinkSummary);
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

	export class DiagnosticExportResult {
	    exported: boolean;
	    path?: string;

	    static createFrom(source: any = {}) {
	        return new DiagnosticExportResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exported = source["exported"];
	        this.path = source["path"];
	    }
	}
	export class JellyfinNetworkStatus {
	    enabled: boolean;
	    urls: string[];

	    static createFrom(source: any = {}) {
	        return new JellyfinNetworkStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.urls = source["urls"];
	    }
	}

}

export namespace onboarding {

	export class PreparationResult {
	    ready: boolean;
	    installed: boolean;
	    started: boolean;
	    version?: string;

	    static createFrom(source: any = {}) {
	        return new PreparationResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ready = source["ready"];
	        this.installed = source["installed"];
	        this.started = source["started"];
	        this.version = source["version"];
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

	export class ApplicationDataStatus {
	    applicationId: string;
	    present: boolean;

	    static createFrom(source: any = {}) {
	        return new ApplicationDataStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.present = source["present"];
	    }
	}
	export class ArchivedApplicationData {
	    applicationId: string;
	    archived: boolean;
	    archivePath?: string;

	    static createFrom(source: any = {}) {
	        return new ArchivedApplicationData(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.archived = source["archived"];
	        this.archivePath = source["archivePath"];
	    }
	}
	export class BackupResult {
	    applicationId: string;
	    path: string;
	    sha256: string;
	    fileCount: number;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new BackupResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.path = source["path"];
	        this.sha256 = source["sha256"];
	        this.fileCount = source["fileCount"];
	        this.createdAt = source["createdAt"];
	    }
	}
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

