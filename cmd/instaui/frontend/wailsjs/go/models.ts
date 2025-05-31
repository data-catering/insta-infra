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
	
	export class ContainerInfo {
	    name: string;
	    serviceName: string;
	    status: string;
	    role: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ContainerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.serviceName = source["serviceName"];
	        this.status = source["status"];
	        this.role = source["role"];
	        this.description = source["description"];
	    }
	}
	export class EdgeData {
	    label: string;
	    type: string;
	    animated: boolean;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new EdgeData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.type = source["type"];
	        this.animated = source["animated"];
	        this.color = source["color"];
	    }
	}
	export class GraphEdge {
	    id: string;
	    source: string;
	    target: string;
	    type: string;
	    data: EdgeData;
	
	    static createFrom(source: any = {}) {
	        return new GraphEdge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source = source["source"];
	        this.target = source["target"];
	        this.type = source["type"];
	        this.data = this.convertValues(source["data"], EdgeData);
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
	export class NodeData {
	    serviceName: string;
	    type: string;
	    status: string;
	    health: string;
	    dependencies: string[];
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new NodeData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.health = source["health"];
	        this.dependencies = source["dependencies"];
	        this.color = source["color"];
	    }
	}
	export class NodePosition {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new NodePosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class GraphNode {
	    id: string;
	    label: string;
	    type: string;
	    status: string;
	    position: NodePosition;
	    data: NodeData;
	
	    static createFrom(source: any = {}) {
	        return new GraphNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.position = this.convertValues(source["position"], NodePosition);
	        this.data = this.convertValues(source["data"], NodeData);
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
	export class DependencyGraph {
	    nodes: GraphNode[];
	    edges: GraphEdge[];
	
	    static createFrom(source: any = {}) {
	        return new DependencyGraph(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodes = this.convertValues(source["nodes"], GraphNode);
	        this.edges = this.convertValues(source["edges"], GraphEdge);
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
	export class DependencyInfo {
	    serviceName: string;
	    status: string;
	    type: string;
	    health: string;
	    required: boolean;
	    startupOrder: number;
	    error?: string;
	    failureReason?: string;
	    containerStatus?: string;
	    exitCode?: number;
	    hasLogs: boolean;
	    lastFailureTime?: string;
	
	    static createFrom(source: any = {}) {
	        return new DependencyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.status = source["status"];
	        this.type = source["type"];
	        this.health = source["health"];
	        this.required = source["required"];
	        this.startupOrder = source["startupOrder"];
	        this.error = source["error"];
	        this.failureReason = source["failureReason"];
	        this.containerStatus = source["containerStatus"];
	        this.exitCode = source["exitCode"];
	        this.hasLogs = source["hasLogs"];
	        this.lastFailureTime = source["lastFailureTime"];
	    }
	}
	export class DependencyStatus {
	    serviceName: string;
	    dependencies: DependencyInfo[];
	    allDependenciesReady: boolean;
	    canStart: boolean;
	    requiredCount: number;
	    runningCount: number;
	    errorCount: number;
	    failedDependencies?: string[];
	
	    static createFrom(source: any = {}) {
	        return new DependencyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.dependencies = this.convertValues(source["dependencies"], DependencyInfo);
	        this.allDependenciesReady = source["allDependenciesReady"];
	        this.canStart = source["canStart"];
	        this.requiredCount = source["requiredCount"];
	        this.runningCount = source["runningCount"];
	        this.errorCount = source["errorCount"];
	        this.failedDependencies = source["failedDependencies"];
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
	export class ServiceContainerDetails {
	    serviceName: string;
	    containers: ContainerInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ServiceContainerDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.containers = this.convertValues(source["containers"], ContainerInfo);
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

