import { ComponentFixture, TestBed } from '@angular/core/testing';

import { YhsHelloWorldComponent } from './yhs-hello-world.component';

describe('YhsHelloWorldComponent', () => {
  let component: YhsHelloWorldComponent;
  let fixture: ComponentFixture<YhsHelloWorldComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [YhsHelloWorldComponent]
    });
    fixture = TestBed.createComponent(YhsHelloWorldComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
