export namespace application {
	
	export class ApplicationSummary {
	    id: string;
	    name: string;
	    description: string;
	    category: string;
	    url: string;
	    optional: boolean;
	
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
	    }
	}

}

