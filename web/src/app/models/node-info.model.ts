import { AllocationInfo } from './alloc-info.model';

export class NodeInfo {
  isSelected = false;
  constructor(
    public nodeId: string,
    public hostName: string,
    public rackName: string,
    public partitionName: string,
    public capacity: string,
    public allocated: string,
    public occupied: string,
    public available: string,
    public utilized: string,
    public allocations: AllocationInfo[] | null,
    public attributes: Attributes,
  ) {}

  setAllocations(allocs: AllocationInfo[]) {
    this.allocations = allocs;
  }
}

export interface Attributes{
  [key: string]: string;
}
