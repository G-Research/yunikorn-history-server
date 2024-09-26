import { CommonModule } from "@angular/common";
import { NgModule } from "@angular/core";
import { AppsViewComponent } from "./apps-view.component";

@NgModule({
  declarations: [AppsViewComponent],
  imports: [CommonModule],
  exports: [AppsViewComponent],
})
export class AppsViewModule {}
