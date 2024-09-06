export class QueueInfo {
  queueName: string = '';
  status: string = '';
  partitionName: string = '';
  maxResource: string = '';
  guaranteedResource: string = '';
  allocatedResource: string = '';
  pendingResource: string='';
  absoluteUsedCapacity: string = '';
  absoluteUsedPercent: number = 0;
  parentQueue: null | QueueInfo = null;
  children: null | QueueInfo[] = null;
  properties: QueuePropertyItem[] = [];
  template: null | QueueTemplate = null;
  isLeaf: boolean = false;
  isManaged: boolean = false;
  isExpanded: boolean = false;
  isSelected: boolean = false;
}

export interface QueuePropertyItem {
  name: string;
  value: string;
}

export interface QueueTemplate {
  maxResource: string;
  guaranteedResource: string;
  properties: { [key: string]: string };
}

export interface SchedulerInfo {
  rootQueue: QueueInfo;
}

export interface ToggleQueueChildrenEvent {
  queueItem: QueueInfo;
  nextLevel: string;
}
