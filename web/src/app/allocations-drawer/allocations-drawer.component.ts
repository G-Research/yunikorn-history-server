import { Component, EventEmitter, Injectable, Input, OnInit, Output, ViewChild } from "@angular/core";
import { FormControl } from "@angular/forms";
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


  disableSelect = new FormControl(false);
  selectedAllocation: any;
  showDetails: boolean = false;
  allocColumnDef: ColumnDef[] = [];
  allocColumnIds: string[] = [];
  selectedAllocationsRow: number = -1;

  displayedColumns: string[] = ['key', 'value'];

  dataSource = new MatTableDataSource<{ key: string, value: string|undefined }>([]);

  filteredDataSource = new MatTableDataSource<AllocationInfo & { expanded: boolean }>([]);

  selectedState: string = '';
  selectedNode: string = '';
  selectedInstance: string = '';

  states = ['Running', 'Unknown', 'Failed', 'Succeeded'];
  nodes = ['lima-rancher-desktop', 'Custom name 1', 'Custom name 2', 'Custom name 3'];
  instances: string[] = [];

  ngOnChanges(): void {
    if (this.allocDataSource) {
      this.updateInstances();
      this.allocDataSource.paginator = this.allocPaginator;
      this.allocDataSource.sort = this.allocSort;
      this.applyFilter();
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

  applyFilter(): void {
    this.filteredDataSource.data = this.allocDataSource.data.filter(item => {
      /*
      const matchesState = this.selectedState ? item.state === this.selectedState : true;
      const matchesNode = this.selectedNode ? item.node === this.selectedNode : true;
      const matchesInstance = this.selectedInstance ? item.instance === this.selectedInstance : true;
      return matchesState && matchesNode && matchesInstance;

    */
    const matchesInstance = this.selectedInstance ? item.displayName === this.selectedInstance : true;
    return matchesInstance;
    });
  }
  
  updateInstances(): void {
    if (this.allocDataSource && this.allocDataSource.data) {
      this.instances = [...new Set(this.allocDataSource.data.map(item => item.displayName))];
    }
  }

  formatResources(colValue: string): string[] {
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
    const newData = [
      { key: 'User', value: undefined },
      { key: 'Name', value: undefined },
      { key: 'Application Type', value: "spark" },
      { key: 'Application Tags', value: undefined },
      { key: 'YarnApplication State', value: this.selectedRow?.applicationState },
      { key: 'FinalStatus Reported by AM', value: undefined },
      { key: 'Started', value: undefined },
      { key: 'Launched', value: undefined },
      { key: 'Finished', value: undefined },
      { key: 'Elapsed', value: undefined },
    ];

    this.dataSource.data = newData;
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
