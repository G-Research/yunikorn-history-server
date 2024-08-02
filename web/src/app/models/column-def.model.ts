export interface ColumnDef {
  colId: string;
  colName: string;
  colFormatter?: (val: any) => any;
  colWidth?: number;
}
