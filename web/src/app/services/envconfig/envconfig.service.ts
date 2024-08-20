import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';

import { EnvConfig } from '@app/models/envconfig.model';
import { LoadRemoteModuleEsmOptions } from '@angular-architects/module-federation';

const ENV_CONFIG_JSON_URL = './assets/config/envconfig.json';

export function envConfigFactory(envConfig: EnvconfigService) {
  return () => envConfig.loadEnvConfig();
}

@Injectable({
  providedIn: 'root',
})
export class EnvconfigService {
  private envConfig: EnvConfig;
  private uiProtocol: string;
  private uiHostname: string;
  private uiPort: string;

  constructor(private httpClient: HttpClient) {
    this.uiProtocol = window.location.protocol;
    this.uiHostname = window.location.hostname;
    this.uiPort = window.location.port;
    this.envConfig = {
      localSchedulerWebAddress: 'http://localhost:9889',
    };
  }

  loadEnvConfig(): Promise<void> {
    return new Promise((resolve) => {
      this.httpClient.get<EnvConfig>(ENV_CONFIG_JSON_URL).subscribe((data) => {
        this.envConfig = data;
        resolve();
      });
    });
  }

  getSchedulerWebAddress() {
    if (!environment.production) {
      return this.envConfig.localSchedulerWebAddress;
    }

    return `${this.uiProtocol}//${this.uiHostname}:${this.uiPort}`;
  }

  getAllocationsDrawerComponentRemoteConfig(): LoadRemoteModuleEsmOptions | null {
    if (
      this.envConfig.allocationsDrawerRemoteComponent &&
      this.envConfig.moduleFederationRemoteEntry
    )
      return {
        type: 'module',
        remoteEntry: this.envConfig.moduleFederationRemoteEntry,
        exposedModule: this.envConfig.allocationsDrawerRemoteComponent,
      };
    return null;
  }
}
