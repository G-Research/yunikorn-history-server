export interface ResourceInfo {
  memory: string;
  vcore: string;
  [key: string]: string;
}

export interface SchedulerResourceInfo {
  memory: number;
  vcore: number;
  [key: string]: number;
}
