import { Injectable, Injector } from '@angular/core';
import { loadRemoteModule, LoadRemoteModuleEsmOptions } from '@angular-architects/module-federation';
import { SchedulerService } from './scheduler.service';

@Injectable({
  providedIn: 'root',
})
export class SchedulerServiceLoader {
  constructor(private injector: Injector) {}

  async initializeSchedulerService(remoteComponentConfig: LoadRemoteModuleEsmOptions | null): Promise<SchedulerService | null> {
    if (remoteComponentConfig !== null) {
      try {
        const remoteModule = await loadRemoteModule(remoteComponentConfig);
        if (remoteModule && remoteModule.SchedulerService) {
          return this.injector.get(remoteModule.SchedulerService);
        } else {
          console.error('SchedulerService not found.');
          return null;
        }
      } catch (error) {
        console.error('Error loading the remote module:', error);
        return null;
      }
    }
    return null;
  }
}