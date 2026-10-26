import type { DemoTransaction } from "../../types";

const cell = (value: string) => (/[",\r\n]/.test(value) ? `"${value.replace(/"/g, '""')}"` : value);

/** date, amount, title, category, essential for every transaction, with the user's edits applied. */
export function buildCsv(
  transactions: DemoTransaction[],
  categoryOf: (t: DemoTransaction) => string,
  essentialOf: (t: DemoTransaction) => boolean,
): string {
  const rows = transactions.map((t) =>
    [t.date, t.amount.toFixed(2), t.title, categoryOf(t), essentialOf(t) ? "yes" : "no"].map(cell).join(","),
  );
  return ["date,amount,title,category,essential", ...rows].join("\r\n") + "\r\n";
}

export function downloadCsv(
  transactions: DemoTransaction[],
  categoryOf: (t: DemoTransaction) => string,
  essentialOf: (t: DemoTransaction) => boolean,
) {
  const url = URL.createObjectURL(new Blob([buildCsv(transactions, categoryOf, essentialOf)], { type: "text/csv" }));
  const link = document.createElement("a");
  link.href = url;
  link.download = "expenses.csv";
  link.click();
  URL.revokeObjectURL(url);
}
