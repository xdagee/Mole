export namespace analyze {
	
	export class Node {
	    name: string;
	    path: string;
	    size: number;
	    is_dir: boolean;
	    children?: Node[];
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.is_dir = source["is_dir"];
	        this.children = this.convertValues(source["children"], Node);
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

export namespace status {
	
	export class BatteryStatus {
	    percent: number;
	    status: string;
	    time_left: string;
	    health: string;
	    cycle_count: number;
	    capacity: number;
	
	    static createFrom(source: any = {}) {
	        return new BatteryStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.percent = source["percent"];
	        this.status = source["status"];
	        this.time_left = source["time_left"];
	        this.health = source["health"];
	        this.cycle_count = source["cycle_count"];
	        this.capacity = source["capacity"];
	    }
	}
	export class BluetoothDevice {
	    name: string;
	    connected: boolean;
	    battery: string;
	
	    static createFrom(source: any = {}) {
	        return new BluetoothDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.connected = source["connected"];
	        this.battery = source["battery"];
	    }
	}
	export class CPUStatus {
	    usage: number;
	    per_core: number[];
	    per_core_estimated: boolean;
	    load1: number;
	    load5: number;
	    load15: number;
	    core_count: number;
	    logical_cpu: number;
	    p_core_count: number;
	    e_core_count: number;
	
	    static createFrom(source: any = {}) {
	        return new CPUStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.usage = source["usage"];
	        this.per_core = source["per_core"];
	        this.per_core_estimated = source["per_core_estimated"];
	        this.load1 = source["load1"];
	        this.load5 = source["load5"];
	        this.load15 = source["load15"];
	        this.core_count = source["core_count"];
	        this.logical_cpu = source["logical_cpu"];
	        this.p_core_count = source["p_core_count"];
	        this.e_core_count = source["e_core_count"];
	    }
	}
	export class DiskIOStatus {
	    read_rate: number;
	    write_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskIOStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.read_rate = source["read_rate"];
	        this.write_rate = source["write_rate"];
	    }
	}
	export class DiskStatus {
	    mount: string;
	    device: string;
	    used: number;
	    total: number;
	    used_percent: number;
	    fstype: string;
	    external: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiskStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mount = source["mount"];
	        this.device = source["device"];
	        this.used = source["used"];
	        this.total = source["total"];
	        this.used_percent = source["used_percent"];
	        this.fstype = source["fstype"];
	        this.external = source["external"];
	    }
	}
	export class GPUStatus {
	    name: string;
	    usage: number;
	    memory_used: number;
	    memory_total: number;
	    core_count: number;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new GPUStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.usage = source["usage"];
	        this.memory_used = source["memory_used"];
	        this.memory_total = source["memory_total"];
	        this.core_count = source["core_count"];
	        this.note = source["note"];
	    }
	}
	export class HardwareInfo {
	    model: string;
	    cpu_model: string;
	    total_ram: string;
	    disk_size: string;
	    os_version: string;
	    refresh_rate: string;
	
	    static createFrom(source: any = {}) {
	        return new HardwareInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.cpu_model = source["cpu_model"];
	        this.total_ram = source["total_ram"];
	        this.disk_size = source["disk_size"];
	        this.os_version = source["os_version"];
	        this.refresh_rate = source["refresh_rate"];
	    }
	}
	export class MemoryStatus {
	    used: number;
	    total: number;
	    available: number;
	    used_percent: number;
	    swap_used: number;
	    swap_total: number;
	    cached: number;
	    pressure: string;
	
	    static createFrom(source: any = {}) {
	        return new MemoryStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.used = source["used"];
	        this.total = source["total"];
	        this.available = source["available"];
	        this.used_percent = source["used_percent"];
	        this.swap_used = source["swap_used"];
	        this.swap_total = source["swap_total"];
	        this.cached = source["cached"];
	        this.pressure = source["pressure"];
	    }
	}
	export class ProcessAlert {
	    pid: number;
	    name: string;
	    command?: string;
	    cpu: number;
	    threshold: number;
	    window: string;
	    // Go type: time
	    triggered_at: any;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessAlert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.command = source["command"];
	        this.cpu = source["cpu"];
	        this.threshold = source["threshold"];
	        this.window = source["window"];
	        this.triggered_at = this.convertValues(source["triggered_at"], null);
	        this.status = source["status"];
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
	export class ProcessWatchConfig {
	    enabled: boolean;
	    cpu_threshold: number;
	    window: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessWatchConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.cpu_threshold = source["cpu_threshold"];
	        this.window = source["window"];
	    }
	}
	export class ProcessInfo {
	    pid: number;
	    ppid: number;
	    name: string;
	    command: string;
	    cpu: number;
	    memory: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.ppid = source["ppid"];
	        this.name = source["name"];
	        this.command = source["command"];
	        this.cpu = source["cpu"];
	        this.memory = source["memory"];
	    }
	}
	export class SensorReading {
	    label: string;
	    value: number;
	    unit: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new SensorReading(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.note = source["note"];
	    }
	}
	export class ThermalStatus {
	    cpu_temp: number;
	    gpu_temp: number;
	    battery_temp: number;
	    fan_speed: number;
	    fan_count: number;
	    system_power: number;
	    adapter_power: number;
	    battery_power: number;
	
	    static createFrom(source: any = {}) {
	        return new ThermalStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu_temp = source["cpu_temp"];
	        this.gpu_temp = source["gpu_temp"];
	        this.battery_temp = source["battery_temp"];
	        this.fan_speed = source["fan_speed"];
	        this.fan_count = source["fan_count"];
	        this.system_power = source["system_power"];
	        this.adapter_power = source["adapter_power"];
	        this.battery_power = source["battery_power"];
	    }
	}
	export class ProxyStatus {
	    enabled: boolean;
	    type: string;
	    host: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.type = source["type"];
	        this.host = source["host"];
	    }
	}
	export class NetworkHistory {
	    rx_history: number[];
	    tx_history: number[];
	
	    static createFrom(source: any = {}) {
	        return new NetworkHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rx_history = source["rx_history"];
	        this.tx_history = source["tx_history"];
	    }
	}
	export class NetworkStatus {
	    name: string;
	    rx_rate_mbs: number;
	    tx_rate_mbs: number;
	    ip: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.rx_rate_mbs = source["rx_rate_mbs"];
	        this.tx_rate_mbs = source["tx_rate_mbs"];
	        this.ip = source["ip"];
	    }
	}
	export class MetricsSnapshot {
	    // Go type: time
	    collected_at: any;
	    host: string;
	    platform: string;
	    uptime: string;
	    uptime_seconds: number;
	    procs: number;
	    hardware: HardwareInfo;
	    health_score: number;
	    health_score_msg: string;
	    cpu: CPUStatus;
	    gpu: GPUStatus[];
	    memory: MemoryStatus;
	    disks: DiskStatus[];
	    trash_size: number;
	    trash_approx: boolean;
	    disk_io: DiskIOStatus;
	    network: NetworkStatus[];
	    network_history: NetworkHistory;
	    proxy: ProxyStatus;
	    batteries: BatteryStatus[];
	    thermal: ThermalStatus;
	    sensors: SensorReading[];
	    bluetooth: BluetoothDevice[];
	    top_processes: ProcessInfo[];
	    process_watch: ProcessWatchConfig;
	    process_alerts: ProcessAlert[];
	
	    static createFrom(source: any = {}) {
	        return new MetricsSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collected_at = this.convertValues(source["collected_at"], null);
	        this.host = source["host"];
	        this.platform = source["platform"];
	        this.uptime = source["uptime"];
	        this.uptime_seconds = source["uptime_seconds"];
	        this.procs = source["procs"];
	        this.hardware = this.convertValues(source["hardware"], HardwareInfo);
	        this.health_score = source["health_score"];
	        this.health_score_msg = source["health_score_msg"];
	        this.cpu = this.convertValues(source["cpu"], CPUStatus);
	        this.gpu = this.convertValues(source["gpu"], GPUStatus);
	        this.memory = this.convertValues(source["memory"], MemoryStatus);
	        this.disks = this.convertValues(source["disks"], DiskStatus);
	        this.trash_size = source["trash_size"];
	        this.trash_approx = source["trash_approx"];
	        this.disk_io = this.convertValues(source["disk_io"], DiskIOStatus);
	        this.network = this.convertValues(source["network"], NetworkStatus);
	        this.network_history = this.convertValues(source["network_history"], NetworkHistory);
	        this.proxy = this.convertValues(source["proxy"], ProxyStatus);
	        this.batteries = this.convertValues(source["batteries"], BatteryStatus);
	        this.thermal = this.convertValues(source["thermal"], ThermalStatus);
	        this.sensors = this.convertValues(source["sensors"], SensorReading);
	        this.bluetooth = this.convertValues(source["bluetooth"], BluetoothDevice);
	        this.top_processes = this.convertValues(source["top_processes"], ProcessInfo);
	        this.process_watch = this.convertValues(source["process_watch"], ProcessWatchConfig);
	        this.process_alerts = this.convertValues(source["process_alerts"], ProcessAlert);
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

