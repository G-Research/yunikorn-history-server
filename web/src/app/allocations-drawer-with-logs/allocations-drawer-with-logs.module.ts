import { CommonModule } from "@angular/common";
import { NgModule } from "@angular/core";
import { AllocationsDrawerWithLogsComponent } from "./allocations-drawer-with-logs.component";

@NgModule({
  declarations: [AllocationsDrawerWithLogsComponent],
  imports: [CommonModule],
  exports: [AllocationsDrawerWithLogsComponent],
})
export class AllocationsDrawerWithLogsModule {}
