import { CommonUtil } from "./common.util";

describe("CommonUtil", () => {
  it("should have createUniqId method", () => {
    expect(CommonUtil.createUniqId).toBeTruthy();
  });

  it("checking formatMemoryBytes method result", () => {
    const inputs: number[] = [0, 100, 1100, 1200000, 1048576000, 1300000000, 1400000000000, 1500000000000000, 1500000000000000000];
    const expected: string[] = ["0 B", "100 B", "1.07 KiB", "1.14 MiB", "1,000 MiB", "1.21 GiB", "1.27 TiB", "1.33 PiB", "1.3 EiB"];
    for (let index = 0; index < inputs.length; index = index + 1) {
      expect(CommonUtil.formatMemoryBytes(inputs[index])).toEqual(expected[index]);
      expect(CommonUtil.formatMemoryBytes(inputs[index].toString())).toEqual(expected[index]);
    }
  });

  it("checking formatEphemeralStorageBytes method result", () => {
    const inputs: number[] = [0, 100, 1100, 1200000, 1048576000, 1300000000, 1400000000000, 1500000000000000, 1500000000000000000];
    const expected: string[] = ["0 B", "100 B", "1.1 kB", "1.2 MB", "1.05 GB", "1.3 GB", "1.4 TB", "1.5 PB", "1.5 EB"];
    for (let index = 0; index < inputs.length; index = index + 1) {
      expect(CommonUtil.formatEphemeralStorageBytes(inputs[index])).toEqual(expected[index]);
      expect(CommonUtil.formatEphemeralStorageBytes(inputs[index].toString())).toEqual(expected[index]);
    }
  });

  it("checking formatCpuCore method result", () => {
    const inputs: number[] = [
      0, 100, 1000, 1555, 1555555, 1555555555, 1555555555555, 1555555555555555, 1555555555555555555, 1555555555555555555555,
    ];
    const expected: string[] = ["0", "100m", "1", "1.56", "1.56k", "1.56M", "1.56G", "1.56T", "1.56P", "1.56E"];
    for (let index = 0; index < inputs.length; index = index + 1) {
      expect(CommonUtil.formatCpuCore(inputs[index])).toEqual(expected[index]);
      expect(CommonUtil.formatCpuCore(inputs[index].toString())).toEqual(expected[index]);
    }
  });

  it("checking formatOtherResource method result", () => {
    const inputs: number[] = [0, 100, 1000, 1555, 1555555, 1555555555, 1555555555555, 1555555555555555, 1555555555555555555];
    const expected: string[] = ["0", "100", "1k", "1.56k", "1.56M", "1.56G", "1.56T", "1.56P", "1.56E"];
    for (let index = 0; index < inputs.length; index = index + 1) {
      expect(CommonUtil.formatOtherResource(inputs[index])).toEqual(expected[index]);
      expect(CommonUtil.formatOtherResource(inputs[index].toString())).toEqual(expected[index]);
    }
  });
});
