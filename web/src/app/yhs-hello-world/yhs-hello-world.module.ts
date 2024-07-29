import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import {YhsHelloWorldComponent} from './yhs-hello-world.component';



@NgModule({
  declarations: [YhsHelloWorldComponent],
  imports: [
    CommonModule,
  ],
  exports: [YhsHelloWorldComponent]
})
export class YhsHelloWorldModule {  }
