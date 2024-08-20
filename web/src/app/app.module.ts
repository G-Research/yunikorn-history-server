import { NgModule } from "@angular/core";
import { BrowserModule } from "@angular/platform-browser";
import { envConfigFactory, EnvConfigService } from "@app/services/envconfig/envconfig.service";
import { AllocationsDrawerModule } from "./allocations-drawer/allocations-drawer.module";
import { AppComponent } from "./app.component";

@NgModule({
  declarations: [AppComponent],
  imports: [BrowserModule, AllocationsDrawerModule],
  providers: [
    {
      useFactory: envConfigFactory,
      provide: EnvConfigService,
      deps: [EnvConfigService],
    }
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
