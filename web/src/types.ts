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

export type EditableField = "category" | "essential";

export const DEMO_STEPS = [
  "Upload",
  "Process",
  "Review",
  "Dashboard",
  "Edit data",
] as const;
