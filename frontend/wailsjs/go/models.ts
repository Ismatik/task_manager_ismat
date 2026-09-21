export namespace domain {
	
	export class Date {
	    Year: number;
	    Month: number;
	    Day: number;
	
	    static createFrom(source: any = {}) {
	        return new Date(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Year = source["Year"];
	        this.Month = source["Month"];
	        this.Day = source["Day"];
	    }
	}
	export class Node {
	    id: string;
	    parentId?: string;
	    type: string;
	    title: string;
	    descriptionMd: string;
	    status: string;
	    due?: Date;
	    dueSource: string;
	    priority: number;
	    estimateMin?: number;
	    recurrence?: string;
	    activity?: string;
	    sortOrder: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    // Go type: time
	    completedAt?: any;
	    // Go type: time
	    archivedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.parentId = source["parentId"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.descriptionMd = source["descriptionMd"];
	        this.status = source["status"];
	        this.due = this.convertValues(source["due"], Date);
	        this.dueSource = source["dueSource"];
	        this.priority = source["priority"];
	        this.estimateMin = source["estimateMin"];
	        this.recurrence = source["recurrence"];
	        this.activity = source["activity"];
	        this.sortOrder = source["sortOrder"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.completedAt = this.convertValues(source["completedAt"], null);
	        this.archivedAt = this.convertValues(source["archivedAt"], null);
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
	export class Tag {
	    id: string;
	    name: string;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	    }
	}

}

export namespace service {
	
	export class TimerView {
	    running: boolean;
	    entryId: string;
	    // Go type: time
	    startedAt?: any;
	    elapsedSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new TimerView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.entryId = source["entryId"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.elapsedSeconds = source["elapsedSeconds"];
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
	export class ProgressView {
	    done: number;
	    total: number;
	    percent: number;
	    defined: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProgressView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.done = source["done"];
	        this.total = source["total"];
	        this.percent = source["percent"];
	        this.defined = source["defined"];
	    }
	}
	export class NodeView {
	    node: domain.Node;
	    status: string;
	    progress: ProgressView;
	    overdue: boolean;
	    isLeaf: boolean;
	    tags: domain.Tag[];
	    timer: TimerView;
	    children: NodeView[];
	
	    static createFrom(source: any = {}) {
	        return new NodeView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.node = this.convertValues(source["node"], domain.Node);
	        this.status = source["status"];
	        this.progress = this.convertValues(source["progress"], ProgressView);
	        this.overdue = source["overdue"];
	        this.isLeaf = source["isLeaf"];
	        this.tags = this.convertValues(source["tags"], domain.Tag);
	        this.timer = this.convertValues(source["timer"], TimerView);
	        this.children = this.convertValues(source["children"], NodeView);
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
	export class ColumnView {
	    status: string;
	    nodes: NodeView[];
	
	    static createFrom(source: any = {}) {
	        return new ColumnView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.nodes = this.convertValues(source["nodes"], NodeView);
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
	export class HabitView {
	    node: domain.Node;
	    scheduledToday: boolean;
	    checkedToday: boolean;
	    streak: number;
	
	    static createFrom(source: any = {}) {
	        return new HabitView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.node = this.convertValues(source["node"], domain.Node);
	        this.scheduledToday = source["scheduledToday"];
	        this.checkedToday = source["checkedToday"];
	        this.streak = source["streak"];
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
	export class NewNode {
	    ParentID?: string;
	    Type: string;
	    Title: string;
	    DescriptionMD: string;
	    Status: string;
	    Due?: domain.Date;
	    Priority: number;
	    EstimateMin?: number;
	    Recurrence?: string;
	    Activity?: string;
	
	    static createFrom(source: any = {}) {
	        return new NewNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ParentID = source["ParentID"];
	        this.Type = source["Type"];
	        this.Title = source["Title"];
	        this.DescriptionMD = source["DescriptionMD"];
	        this.Status = source["Status"];
	        this.Due = this.convertValues(source["Due"], domain.Date);
	        this.Priority = source["Priority"];
	        this.EstimateMin = source["EstimateMin"];
	        this.Recurrence = source["Recurrence"];
	        this.Activity = source["Activity"];
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
	
	
	export class SearchOptions {
	    IncludeArchived: boolean;
	    Limit: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IncludeArchived = source["IncludeArchived"];
	        this.Limit = source["Limit"];
	    }
	}
	export class SettingsView {
	    palette: string;
	    theme: string;
	    accent: string;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.palette = source["palette"];
	        this.theme = source["theme"];
	        this.accent = source["accent"];
	        this.language = source["language"];
	    }
	}

}

