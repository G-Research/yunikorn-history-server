/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DebugElement } from '@angular/core';
import { MatDividerModule } from '@angular/material/divider';
import { MatInputModule } from '@angular/material/input';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSelectModule } from '@angular/material/select';
import { MatSortModule } from '@angular/material/sort';
import { MatTableModule } from '@angular/material/table';
import { By } from '@angular/platform-browser';
import { MatSidenavModule } from '@angular/material/sidenav';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { NgxSpinnerService } from 'ngx-spinner';

import { MockNgxSpinnerService } from '@app/testing/mocks';
import { AllocationsDrawerWithLogsComponent } from './allocations-drawer-with-logs.component';

describe('AllocationsDrawerWithLogsComponent', () => {
  let component: AllocationsDrawerWithLogsComponent;
  let fixture: ComponentFixture<AllocationsDrawerWithLogsComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [AllocationsDrawerWithLogsComponent],
      imports: [
        NoopAnimationsModule,
        MatSidenavModule,
        MatPaginatorModule,
        MatDividerModule,
        MatSortModule,
        MatInputModule,
        MatTableModule,
        MatSelectModule,
      ],
      providers: [{ provide: NgxSpinnerService, useValue: MockNgxSpinnerService }],
    });
    fixture = TestBed.createComponent(AllocationsDrawerWithLogsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should open drawer', () => {
    spyOn(component.matDrawer, 'open');
    component.openDrawer();
    expect(component.matDrawer.open).toHaveBeenCalled();
  });

  it('should close drawer', () => {
    spyOn(component.matDrawer, 'close');
    component.closeDrawer();
    expect(component.matDrawer.close).toHaveBeenCalled();
  });

  it('should toggle allocations detail', () => {
    const initialToggleState = component.allocationsToggle;
    component.allocationsDetailToggle();
    expect(component.allocationsToggle).toBe(!initialToggleState);
  });

  it('should copy the allocations URL to clipboard', () => {
    const debugEl: DebugElement = fixture.debugElement;
    const copyButton = debugEl.query(By.css('.copy-btn'));
    const copyButtonSpy = spyOn(component, 'copyLinkToClipboard');
    copyButton.triggerEventHandler('click', null);
    expect(copyButtonSpy).toHaveBeenCalled();
  });
});
