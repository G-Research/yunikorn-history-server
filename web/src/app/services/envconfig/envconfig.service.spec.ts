import {HttpClientTestingModule} from '@angular/common/http/testing';
import {TestBed} from '@angular/core/testing';

import {EnvconfigService} from './envconfig.service';

describe('EnvconfigService', () => {
  let service: EnvconfigService;

  beforeAll(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [EnvconfigService],
    }).compileComponents();
  });

  beforeEach(() => {
    service = TestBed.inject(EnvconfigService);
  });

  it('should create the service', () => {
    expect(service).toBeTruthy();
  });
});
