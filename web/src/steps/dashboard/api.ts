// The dashboard's own call to the Go server (POST /api/demo/dashboard, documented in demo_dashboard.go).
import type { RowEdit } from "../../types";

export interface DashboardData {
  includeMoneyMovement: boolean;
  totalExpenditure: number;
  transactionCount: number;
  moneyMovement: number; // dollars of transfers, savings and debt payments in the data, counted or not
  firstDate: string; // YYYY-MM-DD
  lastDate: string;
  months: { month: string; expenditure: number }[]; // "2026-06", oldest first
  categories: { name: string; expenditure: number; count: number }[]; // largest first
  essential: { essential: number; nonEssential: number };
  categoryByMonth: { category: string; total: number; byMonth: number[] }[]; // byMonth lines up with months
  merchants: { name: string; total: number; count: number }[];
  largest: { date: string; title: string; amount: number; category: string }[];
  recurring: { name: string; typicalAmount: number; months: number; category: string }[];
}

export async function postDashboard(edits: Record<string, RowEdit>, includeMoneyMovement: boolean): Promise<DashboardData> {
  const response = await fetch("/api/demo/dashboard", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ edits, includeMoneyMovement }),
  });
  if (!response.ok) throw new Error(`/api/demo/dashboard returned ${response.status}`);
  return response.json() as Promise<DashboardData>;
}
