import { CommonModule } from "@angular/common";
import { NgModule } from "@angular/core";
import { MatPaginatorModule } from "@angular/material/paginator";
import { MatSidenavModule } from "@angular/material/sidenav";
import { MatIconModule } from '@angular/material/icon';
import { MatSortModule } from "@angular/material/sort";
import { MatTableModule } from "@angular/material/table";
import { MatTooltipModule } from "@angular/material/tooltip";
import { BrowserModule } from "@angular/platform-browser";
import { AllocationsDrawerComponent } from "./allocations-drawer.component";
import { MatSelectModule } from "@angular/material/select";
import { MatFormFieldModule } from "@angular/material/form-field";

@NgModule({
  declarations: [AllocationsDrawerComponent],
  imports: [CommonModule, MatSortModule, MatSidenavModule, MatPaginatorModule, MatTableModule, MatTooltipModule, MatIconModule, BrowserModule, MatFormFieldModule, MatSelectModule],
  exports: [AllocationsDrawerComponent],
})
export class AllocationsDrawerModule {}
