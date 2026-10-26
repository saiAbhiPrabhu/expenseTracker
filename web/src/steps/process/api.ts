// The Process step's calls: GET /api/demo/statements/{id}/process.
export interface PageReading {
  page: number;
  status: "read" | "skipped";
  reason?: string;
  beforeImage?: string;
  afterImage?: string;
  removedTopPercent?: number;
  ocrText?: string;
}

export interface SectionCheck {
  name: string;
  rows: number;
  printedTotal: number;
  rowsTotal: number;
  matches: boolean;
}

export interface ProcessedRow {
  date: string;
  amount: number;
  title: string;
  scrubbedTitle: string;
  section: string;
  kept: boolean;
}

export interface ProcessedStatement {
  id: string;
  label: string;
  fileName: string;
  pages: number;
  pageReadings: PageReading[];
  sections: SectionCheck[];
  rows: ProcessedRow[];
  summary: { rowsFound: number; depositsRemoved: number; spendingRows: number; flaggedForReview: number };
}

export async function getProcessed(id: string): Promise<ProcessedStatement> {
  const path = `/api/demo/statements/${id}/process`;
  const response = await fetch(path);
  if (!response.ok) throw new Error(`${path} returned ${response.status}`);
  return response.json() as Promise<ProcessedStatement>;
}
