export class PartitionInfo {
  name: string;
  value: string;

  constructor(name: string, value: string) {
    this.name = name;
    this.value = value;
  }
}

export interface Partition {
  name: string;
  state: string;
  clusterId: string;
  capacity: Capacity;
  nodeSortingPolicy: NodeSortingPolicy;
  applications: Applications;
  lastStateTransitionTime: string;
  totalNodes: number;
  totalContainers: number;
}

export interface Capacity {
  capacity: string;
  usedCapacity: string;
}

export interface Applications {
	New: number;
	Accepted: number;
	Starting: number;
	Running: number;
	Rejected: number;
	Completing: number;
	Completed: number;
	Failing: number;
	Failed: number;
	Expired: number;
	Resuming: number;
  total: number;
}

export interface NodeSortingPolicy {
  type: string;
  resourceWeights: {
    memory: number;
    vcore: number;
  };
}
