// Shapes the demo shows. They mirror the Go export (jev_states.json and Decide).

export interface Category20Answer {
  choice: string;
  confidence: number;
  probabilities: Record<string, number>;
}

export interface DemoTransaction {
  key: string;
  date: string; // MM/DD/YY
  title: string;
  amount: number; // dollars, negative for outflow
  category20: Category20Answer;
  essentialProbability: number;
  // Review and edits. Editing a row only ever changes category or essential.
  needsReview: boolean; // category confidence below 0.5
  category: string; // the category in use (JEV's choice until the user changes it)
  essential: boolean; // the essential flag in use
}

export interface CategoryOption {
  name: string;
  description: string;
}

// What the user changed on a row. An absent field is unchanged.
export interface RowEdit {
  category?: string;
  essential?: boolean;
}

// The dashboard numbers from POST /api/demo/summary. Amounts are dollars of money that left the account.
export interface DemoSummary {
  transactionCount: number;
  totalExpenditure: number;
  byMonth: Record<string, number>; // "2026-06" -> dollars
  byCategory: Record<string, number>;
  essential: { essential: number; nonEssential: number };
}

export type EditableField = "category" | "essential";

// A statement the user has added in the Upload step (GET /api/demo/statements lists the sample ones).
export interface StatementInfo {
  id: string; // "2026-06"
  label: string; // "June 2026"
  fileName: string;
  sizeBytes: number;
  pages: number;
}
