export namespace hub {
	
	export class CallResult {
	    server: string;
	    tool: string;
	    isError: boolean;
	    content: mcp.Content[];
	
	    static createFrom(source: any = {}) {
	        return new CallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.tool = source["tool"];
	        this.isError = source["isError"];
	        this.content = this.convertValues(source["content"], mcp.Content);
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
	export class ServerInfo {
	    name: string;
	    url: string;
	    transport: string;
	    status: string;
	    error?: string;
	    added_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.transport = source["transport"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.added_at = source["added_at"];
	    }
	}
	export class ToolInfo {
	    server: string;
	    name: string;
	    description?: string;
	    inputSchema: mcp.InputSchema;
	
	    static createFrom(source: any = {}) {
	        return new ToolInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.inputSchema = this.convertValues(source["inputSchema"], mcp.InputSchema);
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

export namespace mcp {
	
	export class Content {
	    type: string;
	    text?: string;
	    data?: string;
	    mimeType?: string;
	
	    static createFrom(source: any = {}) {
	        return new Content(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.text = source["text"];
	        this.data = source["data"];
	        this.mimeType = source["mimeType"];
	    }
	}
	export class InputSchema {
	    type: string;
	    properties?: Record<string, any>;
	    required?: string[];
	
	    static createFrom(source: any = {}) {
	        return new InputSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.properties = source["properties"];
	        this.required = source["required"];
	    }
	}

}

