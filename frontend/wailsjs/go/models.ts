export namespace config {
	
	export class Config {
	    github_token: string;
	    github_username: string;
	    poll_interval_minutes: number;
	    schedule_days: string[];
	    schedule_start_hour: number;
	    schedule_start_minute: number;
	    schedule_end_hour: number;
	    schedule_end_minute: number;
	    sound_enabled: boolean;
	    sound_path: string;
	    auto_hide_seconds: number;
	    disable_drag: boolean;
	    dnd_enabled: boolean;
	    dnd_hours: number;
	    panel_opacity: number;
	    window_x: number;
	    window_y: number;
	    mock_mode: boolean;
	    theme: string;
	    custom_theme_css: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.github_token = source["github_token"];
	        this.github_username = source["github_username"];
	        this.poll_interval_minutes = source["poll_interval_minutes"];
	        this.schedule_days = source["schedule_days"];
	        this.schedule_start_hour = source["schedule_start_hour"];
	        this.schedule_start_minute = source["schedule_start_minute"];
	        this.schedule_end_hour = source["schedule_end_hour"];
	        this.schedule_end_minute = source["schedule_end_minute"];
	        this.sound_enabled = source["sound_enabled"];
	        this.sound_path = source["sound_path"];
	        this.auto_hide_seconds = source["auto_hide_seconds"];
	        this.disable_drag = source["disable_drag"];
	        this.dnd_enabled = source["dnd_enabled"];
	        this.dnd_hours = source["dnd_hours"];
	        this.panel_opacity = source["panel_opacity"];
	        this.window_x = source["window_x"];
	        this.window_y = source["window_y"];
	        this.mock_mode = source["mock_mode"];
	        this.theme = source["theme"];
	        this.custom_theme_css = source["custom_theme_css"];
	    }
	}

}

export namespace gh {
	
	export class Notification {
	    id: string;
	    reason: string;
	    unread: boolean;
	    title: string;
	    repo: string;
	    url: string;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Notification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.reason = source["reason"];
	        this.unread = source["unread"];
	        this.title = source["title"];
	        this.repo = source["repo"];
	        this.url = source["url"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
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

