export class NodeUtilization {
  constructor(
    public type: string,
    public utilization: {
      bucketName: string;
      numOfNodes: number;
      nodeNames: null | string[];
    }[],
  ) { }
}

export class NodeUtilizationsInfo {
  constructor(
    public clusterId: string,
    public partition: string,
    public utilizations: NodeUtilization[],
  ) { }
}
