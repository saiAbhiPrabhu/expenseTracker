import { useEffect, useMemo, useState } from "react";
import { useDemo } from "../../demo/DemoContext";
import { monthKey, monthLabel } from "../../demo/months";
import ControlBar from "../../flow/ControlBar";
import { useFlow } from "../../flow/FlowContext";
import { money } from "../../format";
import { DashboardData, postDashboard } from "../dashboard/api";

// All rows, with the category and the essential flag editable. The dashboard recomputes from the shared edits.
export default function EditStep() {
  const { goTo } = useFlow();
  const { transactions, categories, edits, categoryOf, essentialOf, setCategory, setEssential, undoRow, resetAll, isEdited, includeMoneyMovement } = useDemo();
  const [month, setMonth] = useState("");
  const [onlyChanged, setOnlyChanged] = useState(false);
  const [summary, setSummary] = useState<DashboardData | null>(null);
  const [summaryError, setSummaryError] = useState(false);
  const changed = Object.keys(edits).length;

  // The previous numbers stay on screen until the new ones arrive.
  useEffect(() => {
    let current = true;
    postDashboard(edits, includeMoneyMovement)
      .then((result) => { if (current) { setSummary(result); setSummaryError(false); } })
      .catch(() => { if (current) setSummaryError(true); });
    return () => { current = false; };
  }, [edits, includeMoneyMovement]);

  const months = useMemo(() => [...new Set(transactions.map((t) => monthKey(t.date)))].sort(), [transactions]);
  const shown = transactions.filter((t) => (month === "" || monthKey(t.date) === month) && (!onlyChanged || isEdited(t)));

  return (
    <div>
      <p className="dash-total" aria-live="polite">
        {summary
          ? `Total expenditure ${money.format(summary.totalExpenditure)}${includeMoneyMovement ? "" : " (without transfers, savings and debt payments)"}. Essential ${money.format(summary.essential.essential)}, non-essential ${money.format(summary.essential.nonEssential)}. ${changed} row${changed === 1 ? "" : "s"} changed.`
          : summaryError ? "" : "Loading the totals…"}
        {summaryError && <span className="notice error"><span className="error-label">Error</span>Could not update the totals.</span>}
      </p>
      <p className="filter-row">
        <label>
          Month{" "}
          <select value={month} onChange={(event) => setMonth(event.target.value)}>
            <option value="">All months</option>
            {months.map((key) => (
              <option key={key} value={key}>{monthLabel(key)}</option>
            ))}
          </select>
        </label>{" "}
        <label>
          <input type="checkbox" checked={onlyChanged} onChange={(event) => setOnlyChanged(event.target.checked)} /> Changed rows only
        </label>{" "}
        {changed > 0 && <button onClick={resetAll}>Undo all changes</button>}
      </p>
      <p className="notice">{shown.length} of {transactions.length} rows</p>
      {shown.length === 0 && (
        <p className="notice">{onlyChanged && changed === 0 ? "No changed rows." : "No rows match."}</p>
      )}
      <table className="review-table edit-table" aria-label="All transactions">
        <thead>
          <tr>
            <th>Date</th>
            <th>Amount</th>
            <th>Title</th>
            <th>Category</th>
            <th>Essential</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {shown.map((row) => (
            <tr key={row.key} className={isEdited(row) ? "edited" : ""}>
              <td className="rv-date">{row.date}</td>
              <td className="amount rv-amount">{money.format(row.amount)}</td>
              <td className="rv-title">{row.title}</td>
              <td className="rv-cat">
                <select
                  value={categoryOf(row)}
                  aria-label={`Category for ${row.title}`}
                  onChange={(event) => setCategory(row, event.target.value)}
                >
                  {categories.map((name) => (
                    <option key={name} value={name}>{name}</option>
                  ))}
                </select>
              </td>
              <td className="rv-ess" data-label="Essential">
                <input
                  type="checkbox"
                  checked={essentialOf(row)}
                  aria-label={`Essential for ${row.title}`}
                  onChange={(event) => setEssential(row, event.target.checked)}
                />
              </td>
              <td className="rv-undo">
                {isEdited(row) && (
                  <button onClick={() => undoRow(row)} aria-label={`Undo change for ${row.title}`}>
                    Undo
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <ControlBar continueLabel="Done" onContinue={() => goTo("dashboard")} hideBack />
    </div>
  );
}
