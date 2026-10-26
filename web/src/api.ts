// Calls to the Go server's demo API (see demo.go).
import type { CategoryOption, DemoSummary, DemoTransaction, RowEdit } from "./types";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) {
    throw new Error(`${path} returned ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export const getTransactions = () => request<DemoTransaction[]>("/api/demo/transactions");
export const getCategories = () => request<CategoryOption[]>("/api/demo/categories");

/** The dashboard numbers with the user's edits applied. The server keeps no state. */
export const postSummary = (edits: Record<string, RowEdit>) =>
  request<DemoSummary>("/api/demo/summary", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ edits }),
  });
