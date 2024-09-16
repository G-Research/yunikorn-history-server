import { NOT_AVAILABLE } from "@app/utils/constants";
import * as moment from "moment";
import { AllocationInfo } from "./alloc-info.model";

export class AppInfo {
  isSelected = false;

  constructor(
    public applicationId: string,
    public usedResource: string,
    public pendingResource: string,
    public maxUsedResource: string,
    public submissionTime: number,
    public finishedTime: null | number,
    public stateLog: Array<StateLog>,
    public lastStateChangeTime: null | number,
    public applicationState: string,
    public allocations: AllocationInfo[] | null
  ) {}

  get formattedSubmissionTime() {
    const millisecs = Math.round(this.submissionTime / (1000 * 1000));
    return moment(millisecs).format("YYYY/MM/DD HH:mm:ss");
  }

  get formattedlastStateChangeTime() {
    if (this.lastStateChangeTime == null) {
      return "n/a";
    }
    const millisecs = Math.round(this.lastStateChangeTime! / (1000 * 1000));
    return moment(millisecs).format("YYYY/MM/DD HH:mm:ss");
  }

  get formattedFinishedTime() {
    if (this.finishedTime) {
      const millisecs = Math.round(this.finishedTime / (1000 * 1000));
      return moment(millisecs).format("YYYY/MM/DD HH:mm:ss");
    }

    return NOT_AVAILABLE;
  }

  setAllocations(allocs: AllocationInfo[]) {
    this.allocations = allocs;
  }

  setLastStateChangeTime() {
    let time = 0;
    this.stateLog.forEach((log) => {
      if (log.time > time) {
        time = log.time;
      }
    });
    this.lastStateChangeTime = time;
  }
}

export class StateLog {
  constructor(public time: number, public applicationState: string) {}
}
