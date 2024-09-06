import { Component, EventEmitter, Injectable, Input, OnInit, Output, ViewChild } from "@angular/core";
import { MatPaginator } from "@angular/material/paginator";
import { MatDrawer } from "@angular/material/sidenav";
import { MatSort } from "@angular/material/sort";
import { MatTableDataSource } from "@angular/material/table";
import { AllocationInfo } from "@app/models/alloc-info.model";
import { AppInfo } from "@app/models/app-info.model";
import { ColumnDef } from "@app/models/column-def.model";
import { EnvConfigService } from "@app/services/envconfig/envconfig.service";
import { CommonUtil } from "@app/utils/common.util";

@Injectable()
@Component({
  selector: "app-allocations-drawer",
  templateUrl: "./allocations-drawer.component.html",
  styleUrls: ["./allocations-drawer.component.scss"],
})
export class AllocationsDrawerComponent implements OnInit {
  @ViewChild("matDrawer", { static: false }) matDrawer!: MatDrawer;
  @ViewChild("allocationMatPaginator", { static: true }) allocPaginator!: MatPaginator;
  @ViewChild("allocSort", { static: true }) allocSort!: MatSort;
  @Input() allocDataSource!: MatTableDataSource<AllocationInfo & { expanded: boolean }>;
  @Input() selectedRow!: AppInfo | null;
  @Input() externalLogsBaseUrl!: string | null;
  @Input() partitionSelected!: string;
  @Input() leafQueueSelected!: string;

  @Output() removeRowSelection = new EventEmitter<void>();

  selectedAllocation: any;
  showDetails: boolean = false;
  allocColumnDef: ColumnDef[] = [];
  allocColumnIds: string[] = [];
  selectedAllocationsRow: number = -1;

  displayedColumns: string[] = ['key', 'value'];

  dataSource = new MatTableDataSource<{ key: string, value: string, link?: string }>([]);



  ngOnChanges(): void {
    if (this.allocDataSource) {
      this.allocDataSource.paginator = this.allocPaginator;
      this.allocDataSource.sort = this.allocSort;
    }
  }
  constructor(private envConfig: EnvConfigService) {}

  ngOnInit(): void {
    this.allocColumnDef = [
      { colId: "displayName", colName: "Display Name", colWidth: 1 },
      { colId: "resource", colName: "Resource", colWidth: 1, colFormatter: CommonUtil.resourceColumnFormatter },
      { colId: "nodeId", colName: "Node ID", colWidth: 1 },
      { colId: "state", colName: "State", colWidth: 1 },
    ];
    this.allocColumnIds = this.allocColumnDef.map((col) => col.colId);
    this.externalLogsBaseUrl = this.envConfig.getExternalLogsBaseUrl();
  }
  

  formatResources(colValue: string): string[] {
    console.log(colValue);
    const arr: string[] = colValue.split("<br/>");
    // Check if there are "cpu" or "Memory" elements in the array
    const hasCpu = arr.some((item) => item.toLowerCase().includes("cpu"));
    const hasMemory = arr.some((item) => item.toLowerCase().includes("memory"));
    if (!hasCpu) {
      arr.unshift("CPU: n/a");
    }
    if (!hasMemory) {
      arr.unshift("Memory: n/a");
    }
    
    // Concatenate the two arrays, with "cpu" and "Memory" elements first
    const cpuAndMemoryElements = arr.filter((item) => item.toLowerCase().includes("cpu") || item.toLowerCase().includes("memory"));
    const otherElements = arr.filter((item) => !item.toLowerCase().includes("cpu") && !item.toLowerCase().includes("memory"));
    const result = cpuAndMemoryElements.concat(otherElements);

    return result;
  }

  isAllocDataSourceEmpty() {
    return this.allocDataSource?.data && this.allocDataSource.data.length === 0;
  }

  allocationsDetailToggle(row: any) {
    this.showDetails = true;
    this.selectedAllocation = row;
    console.log("data", row);
    const newData = [
      { key: 'User', value: row.user },
      { key: 'Name', value: row.name },
      { key: 'Application Type', value: row.applicationType },
      { key: 'Application Tags', value: row.applicationTags },
      { key: 'Application Priority', value: row.applicationPriority },
      { key: 'YarnApplication State', value: row.yarnApplicationState },
      { key: 'Queue', value: row.queue },
      { key: 'FinalStatus Reported by AM', value: row.finalStatusReportedByAM },
      { key: 'Started', value: row.started },
      { key: 'Launched', value: row.launched },
      { key: 'Finished', value: row.finished },
      { key: 'Elapsed', value: row.elapsed },
      { key: 'Tracking URL', value: 'History', link: row.trackingUrl },
      { key: 'Log Aggregation Status', value: row.logAggregationStatus },
      { key: 'Application Timeout (Remaining Time)', value: row.applicationTimeout },
      { key: 'Unmanaged Application', value: row.unmanagedApplication },
      { key: 'Application Node Label Expression', value: row.applicationNodeLabelExpression },
      { key: 'AM Container Node Label Expression', value: row.amContainerNodeLabelExpression }
    ];


    console.log('new data', newData);
    this.dataSource.data = newData;
    console.log(this.selectedAllocation);
    if (this.selectedAllocationsRow !== -1) {
      if (this.selectedAllocationsRow !== row) {
        this.allocDataSource.data[this.selectedAllocationsRow].expanded = false;
        this.selectedAllocationsRow = row;
        this.allocDataSource.data[row].expanded = true;
      } else {
        this.allocDataSource.data[this.selectedAllocationsRow].expanded = false;
        this.selectedAllocationsRow = -1;
      }
    } else {
      this.selectedAllocationsRow = row;
      this.allocDataSource.data[row].expanded = true;
    }
  }

  goBackToTable(){
    this.showDetails = false;
  }
  
  closeDrawer() {
    this.selectedAllocationsRow = -1;
    this.matDrawer.close();
    this.removeRowSelection.emit();
  }

  openDrawer() {
    this.matDrawer.open();
  }

  logClick(e: MouseEvent) {
    e.stopPropagation();
  }

  copyLinkToClipboard() {
    const url = window.location.href.split("?")[0];
    const copyString = `${url}?partition=${this.partitionSelected}&queue=${this.leafQueueSelected}&applicationId=${this?.selectedRow?.applicationId}`;
    navigator.clipboard.writeText(copyString).catch((error) => console.error("Writing to the clipboard is not allowed. ", error));
  }
}
