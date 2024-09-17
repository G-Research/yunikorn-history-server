import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SchedulerService } from './scheduler.service';

@NgModule({
    imports: [CommonModule],
    providers: [SchedulerService],
})
export class SchedulerModule { }
