import { NgModule } from "@angular/core";
import { BrowserModule } from "@angular/platform-browser";
import { AllocationsDrawerModule } from "./allocations-drawer/allocations-drawer.module";
import { AppComponent } from "./app.component";

@NgModule({
  declarations: [AppComponent],
  imports: [BrowserModule, AllocationsDrawerModule],
  providers: [],
  bootstrap: [AppComponent],
})
export class AppModule {}
