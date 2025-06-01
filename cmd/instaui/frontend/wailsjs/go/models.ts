export namespace container {
	
	export class RuntimeStatus {
	    name: string;
	    isInstalled: boolean;
	    isRunning: boolean;
	    isAvailable: boolean;
	    hasCompose: boolean;
	    version: string;
	    error: string;
	    canAutoStart: boolean;
	    installationGuide: string;
	    startupCommand: string;
	    requiresMachine: boolean;
	    machineStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isInstalled = source["isInstalled"];
	        this.isRunning = source["isRunning"];
	        this.isAvailable = source["isAvailable"];
	        this.hasCompose = source["hasCompose"];
	        this.version = source["version"];
	        this.error = source["error"];
	        this.canAutoStart = source["canAutoStart"];
	        this.installationGuide = source["installationGuide"];
	        this.startupCommand = source["startupCommand"];
	        this.requiresMachine = source["requiresMachine"];
	        this.machineStatus = source["machineStatus"];
	    }
	}
	export class StartupResult {
	    success: boolean;
	    message: string;
	    error: string;
	    requiresManualAction: boolean;
	
	    static createFrom(source: any = {}) {
	        return new StartupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.requiresManualAction = source["requiresManualAction"];
	    }
	}
	export class SystemRuntimeStatus {
	    hasAnyRuntime: boolean;
	    preferredRuntime: string;
	    availableRuntimes: string[];
	    runtimeStatuses: RuntimeStatus[];
	    recommendedAction: string;
	    canProceed: boolean;
	    platform: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemRuntimeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasAnyRuntime = source["hasAnyRuntime"];
	        this.preferredRuntime = source["preferredRuntime"];
	        this.availableRuntimes = source["availableRuntimes"];
	        this.runtimeStatuses = this.convertValues(source["runtimeStatuses"], RuntimeStatus);
	        this.recommendedAction = source["recommendedAction"];
	        this.canProceed = source["canProceed"];
	        this.platform = source["platform"];
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

export namespace models {
	
	export class ImagePullProgress {
	    status: string;
	    progress: number;
	    currentLayer: string;
	    totalLayers: number;
	    downloaded: number;
	    total: number;
	    speed: string;
	    eta: string;
	    error: string;
	    serviceName: string;
	
	    static createFrom(source: any = {}) {
	        return new ImagePullProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.currentLayer = source["currentLayer"];
	        this.totalLayers = source["totalLayers"];
	        this.downloaded = source["downloaded"];
	        this.total = source["total"];
	        this.speed = source["speed"];
	        this.eta = source["eta"];
	        this.error = source["error"];
	        this.serviceName = source["serviceName"];
	    }
	}
	export class ServiceConnectionInfo {
	    serviceName: string;
	    hasWebUI: boolean;
	    webURL?: string;
	    hostPort?: string;
	    containerPort?: string;
	    username?: string;
	    password?: string;
	    connectionCommand?: string;
	    connectionString?: string;
	    available: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceConnectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.hasWebUI = source["hasWebUI"];
	        this.webURL = source["webURL"];
	        this.hostPort = source["hostPort"];
	        this.containerPort = source["containerPort"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.connectionCommand = source["connectionCommand"];
	        this.connectionString = source["connectionString"];
	        this.available = source["available"];
	        this.error = source["error"];
	    }
	}
	export class ServiceDetailInfo {
	    name: string;
	    type: string;
	    status: string;
	    statusError?: string;
	    dependencies: string[];
	    dependenciesError?: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceDetailInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.statusError = source["statusError"];
	        this.dependencies = source["dependencies"];
	        this.dependenciesError = source["dependenciesError"];
	    }
	}
	export class ServiceInfo {
	    Name: string;
	    Type: string;
	    ConnectionCmd: string;
	    DefaultUser: string;
	    DefaultPassword: string;
	    RequiresPassword: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServiceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.ConnectionCmd = source["ConnectionCmd"];
	        this.DefaultUser = source["DefaultUser"];
	        this.DefaultPassword = source["DefaultPassword"];
	        this.RequiresPassword = source["RequiresPassword"];
	    }
	}

}

