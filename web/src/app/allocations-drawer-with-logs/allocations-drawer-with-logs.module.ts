import { CommonModule } from "@angular/common";
import { NgModule } from "@angular/core";
import { MatPaginatorModule } from "@angular/material/paginator";
import { MatSidenavModule } from "@angular/material/sidenav";
import { MatSortModule } from "@angular/material/sort";
import { MatTableModule } from "@angular/material/table";
import { MatTooltipModule } from "@angular/material/tooltip";
import { AllocationsDrawerWithLogsComponent } from "./allocations-drawer-with-logs.component";

@NgModule({
  declarations: [AllocationsDrawerWithLogsComponent],
  imports: [CommonModule, MatSortModule, MatSidenavModule, MatPaginatorModule, MatTableModule, MatTooltipModule],
  exports: [AllocationsDrawerWithLogsComponent],
})
export class AllocationsDrawerWithLogsModule {}
