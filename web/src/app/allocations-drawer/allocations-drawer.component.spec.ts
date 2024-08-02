import { DebugElement } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MatDividerModule } from "@angular/material/divider";
import { MatInputModule } from "@angular/material/input";
import { MatPaginatorModule } from "@angular/material/paginator";
import { MatSelectModule } from "@angular/material/select";
import { MatSidenavModule } from "@angular/material/sidenav";
import { MatSortModule } from "@angular/material/sort";
import { MatTableModule } from "@angular/material/table";
import { By } from "@angular/platform-browser";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { AllocationsDrawerComponent } from "./allocations-drawer.component";

describe("AllocationsDrawerComponent", () => {
  let component: AllocationsDrawerComponent;
  let fixture: ComponentFixture<AllocationsDrawerComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [AllocationsDrawerComponent],
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
      providers: [],
    });
    fixture = TestBed.createComponent(AllocationsDrawerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should open drawer", () => {
    spyOn(component.matDrawer, "open");
    component.openDrawer();
    expect(component.matDrawer.open).toHaveBeenCalled();
  });

  it("should close drawer", () => {
    spyOn(component.matDrawer, "close");
    component.closeDrawer();
    expect(component.matDrawer.close).toHaveBeenCalled();
  });

  it("should copy the allocations URL to clipboard", () => {
    const debugEl: DebugElement = fixture.debugElement;
    const copyButton = debugEl.query(By.css(".copy-btn"));
    const copyButtonSpy = spyOn(component, "copyLinkToClipboard");
    copyButton.triggerEventHandler("click", null);
    expect(copyButtonSpy).toHaveBeenCalled();
  });
});
