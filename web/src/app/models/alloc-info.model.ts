export class AllocationInfo {
  constructor(
    public displayName: string,
    public allocationKey: string,
    public allocationTags: null | string,
    public resource: string,
    public priority: string,
    public queueName: string,
    public nodeId: string,
    public applicationId: string,
    public partition: string,
    public requestTime: string,
    public allocationTime: string
  ) {}
}
