import { NgModule } from "@angular/core";
import { BrowserModule } from "@angular/platform-browser";

import { AllocationsDrawerWithLogsModule } from "./allocations-drawer-with-logs/allocations-drawer-with-logs.module";
import { AppComponent } from "./app.component";
import { TestModule } from "./test/test.module";
import { YhsHelloWorldModule } from "./yhs-hello-world/yhs-hello-world.module";

@NgModule({
  declarations: [AppComponent],
  imports: [BrowserModule, YhsHelloWorldModule, TestModule, AllocationsDrawerWithLogsModule],
  providers: [],
  bootstrap: [AppComponent],
})
export class AppModule {}
