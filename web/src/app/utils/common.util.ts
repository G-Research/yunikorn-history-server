import { NOT_AVAILABLE } from "@app/utils/constants";
import * as moment from "moment";

export class CommonUtil {
  static createUniqId(prefix?: string): string {
    const uniqid = Math.random().toString(36).substr(2);

    if (prefix) {
      return prefix + uniqid;
    }

    return uniqid;
  }

  static formatMemoryBytes(value: number | string): string {
    const units: readonly string[] = ["KiB", "MiB", "GiB", "TiB", "PiB", "EiB"];
    let unit: string = "B";
    let toValue = +value;
    for (let i = 0, unitslen = units.length; toValue / 1024 >= 1 && i < unitslen; i = i + 1) {
      toValue = toValue / 1024;
      unit = units[i];
    }
    return `${toValue.toLocaleString(undefined, { maximumFractionDigits: 2 })} ${unit}`;
  }

  static formatEphemeralStorageBytes(value: number | string): string {
    const units: readonly string[] = ["kB", "MB", "GB", "TB", "PB", "EB"];
    let unit: string = "B";
    let toValue = +value;
    for (let i = 0, unitslen = units.length; toValue / 1000 >= 1 && i < unitslen; i = i + 1) {
      toValue = toValue / 1000;
      unit = units[i];
    }
    return `${toValue.toLocaleString(undefined, { maximumFractionDigits: 2 })} ${unit}`;
  }

  static isEmpty(arg: object | any[]): boolean {
    return Object.keys(arg).length === 0;
  }

  static formatCpuCore(value: number | string): string {
    const units: readonly string[] = ["m", "", "k", "M", "G", "T", "P", "E"];
    let unit: string = "";
    let toValue = +value;
    if (toValue > 0) {
      unit = units[0];
    }
    for (let i = 1, unitslen = units.length; toValue / 1000 >= 1 && i < unitslen; i = i + 1) {
      toValue = toValue / 1000;
      unit = units[i];
    }
    return `${toValue.toLocaleString(undefined, { maximumFractionDigits: 2 })}${unit}`;
  }

  static formatOtherResource(value: number | string): string {
    const units: readonly string[] = ["k", "M", "G", "T", "P", "E"];
    let unit: string = "";
    let toValue = +value;
    for (let i = 0, unitslen = units.length; toValue / 1000 >= 1 && i < unitslen; i = i + 1) {
      toValue = toValue / 1000;
      unit = units[i];
    }
    return `${toValue.toLocaleString(undefined, { maximumFractionDigits: 2 })}${unit}`;
  }

  static resourceColumnFormatter(value: string): string {
    return value.split(", ").join("<br/>");
  }

  static formatPercent(value: number | string): string {
    const toValue = +value;
    return `${toValue.toFixed(0)}%`;
  }

  static timeColumnFormatter(value: null | number) {
    if (value) {
      const millisecs = Math.round(value / (1000 * 1000));
      return moment(millisecs).format("YYYY/MM/DD HH:mm:ss");
    }
    return NOT_AVAILABLE;
  }
}
