import {HttpClientTestingModule} from '@angular/common/http/testing';
import {TestBed} from '@angular/core/testing';
import {EnvconfigService} from '@app/services/envconfig/envconfig.service';
import {MockEnvconfigService} from '@app/testing/mocks';

import {SchedulerService} from './scheduler.service';
import {SchedulerResourceInfo} from '@app/models/resource-info.model';
import {NOT_AVAILABLE} from '@app/utils/constants';

describe('SchedulerService', () => {
  let service: SchedulerService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [SchedulerService, { provide: EnvconfigService, useValue: MockEnvconfigService }],
    });
    service = TestBed.inject(SchedulerService);
  });

  it('should create the service', () => {
    expect(service).toBeTruthy();
    
  });

  it('should format SchedulerResourceInfo correctly', () => {
    type TestCase = {
      description: string;
      schedulerResourceInfo: SchedulerResourceInfo;
      expected: string;
    };
  
    const testCases: TestCase[] = [
      {
        description: 'test simple resourceInfo',
        schedulerResourceInfo: {
          'memory': 1024,
          'vcore': 2,
        },
        expected: 'Memory: 1 KiB, CPU: 2m'
      },
      {
        description: 'test undefined resourceInfo',
        schedulerResourceInfo : undefined as any,
        expected: `${NOT_AVAILABLE}`
      },
      {
        description: 'test empty resourceInfo',
        schedulerResourceInfo : {} as any,
        expected: `${NOT_AVAILABLE}`
      },
      {
        description: 'Test zero values',
        schedulerResourceInfo: {
          'memory': 0,
          'vcore': 0,
          'ephemeral-storage': 0,
          'hugepages-2Mi': 0,
          'hugepages-1Gi': 0,
          'pods': 0
        },
        expected: 'Memory: 0 B, CPU: 0, pods: 0, ephemeral-storage: 0 B, hugepages-1Gi: 0 B, hugepages-2Mi: 0 B'
      },
      {
        description: 'Test resource ordering',
        schedulerResourceInfo: {
          'ephemeral-storage': 2048,
          'memory': 1024,
          'vcore': 2,
          'TPU': 30000,
          'GPU': 40000,
          'hugepages-2Mi':2097152,
          'hugepages-1Gi':1073741824,
          'pods': 10000
        },
        expected: 'Memory: 1 KiB, CPU: 2m, pods: 10k, ephemeral-storage: 2.05 kB, GPU: 40k, hugepages-1Gi: 1 GiB, hugepages-2Mi: 2 MiB, TPU: 30k'
      }
    ];

    testCases.forEach((testCase: TestCase) => {
      const result = (service as any).formatResource(testCase.schedulerResourceInfo); // ignore type typecheck to access private method
      expect(result).toEqual(testCase.expected);
    });
  });


  it('should format SchedulerResourceInfo percentage correctly', () => {
    type TestCase = {
      description: string;
      schedulerResourceInfo: SchedulerResourceInfo;
      expected: string;
    };
  
    const testCases: TestCase[] = [
      {
        description: 'test simple resourceInfo',
        schedulerResourceInfo: {
          'memory': 10,
          'vcore': 50,
        },
        expected: 'Memory: 10%, CPU: 50%'
      },
      {
        description: 'test undefined resourceInfo',
        schedulerResourceInfo : undefined as any,
        expected: `${NOT_AVAILABLE}`
      },
      {
        description: 'test empty resourceInfo',
        schedulerResourceInfo : {} as any,
        expected: `${NOT_AVAILABLE}`
      },
      {
        description: 'Test zero values and will only show memory and cpu',
        schedulerResourceInfo: {
          'memory': 0,
          'vcore': 0,
          'pods': 0,
          'ephemeral-storage': 0,
        },
        expected: 'Memory: 0%, CPU: 0%'
      }
    ];

    testCases.forEach((testCase: TestCase) => {
      const result = (service as any).formatPercent(testCase.schedulerResourceInfo); // ignore type typecheck to access private method
      expect(result).toEqual(testCase.expected);
    });
  });
});
