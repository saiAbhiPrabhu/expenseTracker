import type { DemoTransaction } from "../types";

export interface MonthPage {
  key: string; // "2026-06"
  label: string; // "June 2026"
  rows: DemoTransaction[]; // the rows that need review in that month
}

// "06/21/26" -> "2026-06"
export const monthKey = (date: string) => `20${date.slice(6, 8)}-${date.slice(0, 2)}`;

export const monthLabel = (key: string) => {
  const [year, month] = key.split("-").map(Number);
  return new Date(year, month - 1, 1).toLocaleString("en-US", { month: "long", year: "numeric" });
};

/** One page per month that appears in the data, oldest first, holding that month's rows needing review. */
export function reviewPages(transactions: DemoTransaction[]): MonthPage[] {
  const pages = new Map<string, MonthPage>();
  for (const transaction of transactions) {
    const key = monthKey(transaction.date);
    if (!pages.has(key)) pages.set(key, { key, label: monthLabel(key), rows: [] });
    if (transaction.needsReview) pages.get(key)!.rows.push(transaction);
  }
  return [...pages.values()].sort((a, b) => a.key.localeCompare(b.key));
}
