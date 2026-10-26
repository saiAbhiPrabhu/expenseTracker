import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { getCategories, getTransactions } from "../api";
import type { DemoTransaction, RowEdit } from "../types";

interface DemoState {
  transactions: DemoTransaction[];
  categories: string[];
  edits: Record<string, RowEdit>;
  /** The category in use for a row: the user's edit, otherwise JEV's answer. */
  categoryOf: (transaction: DemoTransaction) => string;
  /** The essential flag in use for a row: the user's edit, otherwise JEV's answer. */
  essentialOf: (transaction: DemoTransaction) => boolean;
  /** Change a row's category. Choosing JEV's own answer again clears that edit. */
  setCategory: (transaction: DemoTransaction, category: string) => void;
  /** Change a row's essential flag. Choosing JEV's own answer again clears that edit. */
  setEssential: (transaction: DemoTransaction, essential: boolean) => void;
  /** Undo only the category change on a row. */
  undoCategory: (transaction: DemoTransaction) => void;
  /** Undo every change on a row. */
  undoRow: (transaction: DemoTransaction) => void;
  /** Undo every change on every row. */
  resetAll: () => void;
  isEdited: (transaction: DemoTransaction) => boolean;
  /** Whether transfers, savings and debt payments are counted in the totals (dashboard and Edit data). */
  includeMoneyMovement: boolean;
  setIncludeMoneyMovement: (include: boolean) => void;
  /** The month the Review step is showing, kept so coming back lands where the user left off. */
  reviewPage: number;
  setReviewPage: (page: number) => void;
}

const DemoContext = createContext<DemoState | null>(null);

export function useDemo(): DemoState {
  const state = useContext(DemoContext);
  if (!state) throw new Error("useDemo must be used inside DemoProvider");
  return state;
}

type Loaded = { transactions: DemoTransaction[]; categories: string[] };

/** Apply a change to a row's edits, dropping any field that is back to JEV's answer. */
function withEdit(previous: Record<string, RowEdit>, transaction: DemoTransaction, change: RowEdit): Record<string, RowEdit> {
  const merged: RowEdit = { ...previous[transaction.key], ...change };
  if (merged.category === transaction.category20.choice) delete merged.category;
  if (merged.essential === transaction.essential) delete merged.essential;
  const next = { ...previous };
  if (Object.keys(merged).length === 0) delete next[transaction.key];
  else next[transaction.key] = merged;
  return next;
}

// Loads the saved demo run once and holds the user's edits. Nothing is sent back to the server here.
export function DemoProvider({ children }: { children: ReactNode }) {
  const [loaded, setLoaded] = useState<Loaded | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [edits, setEdits] = useState<Record<string, RowEdit>>({});
  const [includeMoneyMovement, setIncludeMoneyMovement] = useState(true);
  const [reviewPage, setReviewPage] = useState(0);

  useEffect(() => {
    Promise.all([getTransactions(), getCategories()])
      .then(([transactions, categories]) => setLoaded({ transactions, categories: categories.map((c) => c.name) }))
      .catch(() => setError("Could not load the demo data. Is the server running?"));
  }, []);

  const categoryOf = useCallback(
    (transaction: DemoTransaction) => edits[transaction.key]?.category ?? transaction.category20.choice,
    [edits],
  );
  const essentialOf = useCallback(
    (transaction: DemoTransaction) => edits[transaction.key]?.essential ?? transaction.essential,
    [edits],
  );
  const setCategory = useCallback(
    (transaction: DemoTransaction, category: string) => setEdits((previous) => withEdit(previous, transaction, { category })),
    [],
  );
  const setEssential = useCallback(
    (transaction: DemoTransaction, essential: boolean) => setEdits((previous) => withEdit(previous, transaction, { essential })),
    [],
  );
  const undoCategory = useCallback(
    (transaction: DemoTransaction) => setCategory(transaction, transaction.category20.choice),
    [setCategory],
  );
  const undoRow = useCallback((transaction: DemoTransaction) => {
    setEdits((previous) => {
      const next = { ...previous };
      delete next[transaction.key];
      return next;
    });
  }, []);
  const resetAll = useCallback(() => setEdits({}), []);
  const isEdited = useCallback((transaction: DemoTransaction) => transaction.key in edits, [edits]);

  const value = useMemo<DemoState | null>(
    () => (loaded ? { ...loaded, edits, categoryOf, essentialOf, setCategory, setEssential, undoCategory, undoRow, resetAll, isEdited, includeMoneyMovement, setIncludeMoneyMovement, reviewPage, setReviewPage } : null),
    [loaded, edits, categoryOf, essentialOf, setCategory, setEssential, undoCategory, undoRow, resetAll, isEdited, includeMoneyMovement, reviewPage],
  );

  if (error) return <p className="notice">{error}</p>;
  if (!value) return <p className="notice">Loading the demo…</p>;
  return <DemoContext.Provider value={value}>{children}</DemoContext.Provider>;
}
