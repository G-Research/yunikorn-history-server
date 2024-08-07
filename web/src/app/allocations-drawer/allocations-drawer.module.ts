import { CommonModule } from "@angular/common";
import { NgModule } from "@angular/core";
import { MatPaginatorModule } from "@angular/material/paginator";
import { MatSidenavModule } from "@angular/material/sidenav";
import { MatSortModule } from "@angular/material/sort";
import { MatTableModule } from "@angular/material/table";
import { MatTooltipModule } from "@angular/material/tooltip";
import { BrowserModule } from "@angular/platform-browser";
import { AllocationsDrawerComponent } from "./allocations-drawer.component";

@NgModule({
  declarations: [AllocationsDrawerComponent],
  imports: [CommonModule, MatSortModule, MatSidenavModule, MatPaginatorModule, MatTableModule, MatTooltipModule, BrowserModule],
  exports: [AllocationsDrawerComponent],
})
export class AllocationsDrawerModule {}
