import { NgModule } from "@angular/core";
import { BrowserModule } from "@angular/platform-browser";
import { envConfigFactory, EnvConfigService } from "@app/services/envconfig/envconfig.service";
import { AllocationsDrawerModule } from "./allocations-drawer/allocations-drawer.module";
import { AppComponent } from "./app.component";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";

@NgModule({
  declarations: [AppComponent],
  imports: [BrowserModule, AllocationsDrawerModule, BrowserAnimationsModule],
  providers: [
    {
      useFactory: envConfigFactory,
      provide: EnvConfigService,
      deps: [EnvConfigService],
    },
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
