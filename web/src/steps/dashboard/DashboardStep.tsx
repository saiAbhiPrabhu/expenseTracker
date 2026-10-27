import { useEffect, useState } from "react";
import { useDemo } from "../../demo/DemoContext";
import { monthLabel } from "../../demo/months";
import ControlBar from "../../flow/ControlBar";
import { useFlow } from "../../flow/FlowContext";
import { money, percent } from "../../format";
import { DashboardData, postDashboard } from "./api";
import { downloadCsv } from "./csv";
import "./dashboard.css";

const longDate = (iso: string) => new Date(+iso.slice(0, 4), +iso.slice(5, 7) - 1, +iso.slice(8, 10)).toLocaleDateString("en-US", { dateStyle: "long" });
const plural = (n: number, word: string) => `${n} ${word}${n === 1 ? "" : "s"}`;

/** A plain bar: its length is `part` out of `whole`. The numbers always sit next to it. */
function Bar({ part, whole }: { part: number; whole: number }) {
  return (
    <div className="dash-bar" aria-hidden="true">
      <div style={{ width: `${whole > 0 ? (part / whole) * 100 : 0}%` }} />
    </div>
  );
}

// The home screen: the numbers from POST /api/demo/dashboard with the user's edits applied.
export default function DashboardStep() {
  const { next } = useFlow();
  const { edits, transactions, categoryOf, essentialOf, includeMoneyMovement, setIncludeMoneyMovement } = useDemo();
  const [data, setData] = useState<DashboardData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let current = true;
    setError(null);
    postDashboard(edits, includeMoneyMovement)
      .then((result) => current && setData(result))
      .catch(() => current && setError("Could not load the dashboard numbers."));
    return () => { current = false; };
  }, [edits, includeMoneyMovement, attempt]);

  if (error) {
    return (
      <div>
        <p className="notice">{error}</p>
        <p><button onClick={() => setAttempt(attempt + 1)}>Try again</button></p>
        <ControlBar continueLabel="Edit data" onContinue={next} />
      </div>
    );
  }
  if (!data) return <p className="notice">Calculating…</p>;

  const total = data.totalExpenditure;
  const changed = Object.keys(edits).length;
  const largestMonth = Math.max(0, ...data.months.map((m) => m.expenditure));
  const largestCategory = data.categories.length > 0 ? data.categories[0].expenditure : 0;
  const essentialShare = total > 0 ? (data.essential.essential / total) * 100 : 0;

  return (
    <div>
      <p className="notice">
        Expenditure is money that left your account.
        {changed > 0 && ` Includes your ${plural(changed, "change")}.`}
      </p>
      <p className="dash-total" aria-live="polite">
        Total expenditure: <strong>{money.format(total)}</strong> across {plural(data.transactionCount, "transaction")}
      </p>
      <p>{longDate(data.firstDate)} to {longDate(data.lastDate)}</p>
      <label className="dash-check">
        <input type="checkbox" checked={includeMoneyMovement} onChange={(event) => setIncludeMoneyMovement(event.target.checked)} />
        {" "}Include transfers, savings and debt payments ({money.format(data.moneyMovement)})
      </label>

      <h2>By month</h2>
      <p className="notice">Bars are relative to the largest month.</p>
      <table className="data-table" aria-label="Expenditure by month">
        <thead><tr><th>Month</th><th className="amount">Expenditure</th><th className="amount">Share</th><th aria-label="Bar compared with the largest month" /></tr></thead>
        <tbody>
          {data.months.map((row) => (
            <tr key={row.month}>
              <th scope="row">{monthLabel(row.month)}</th>
              <td className="amount">{money.format(row.expenditure)}</td>
              <td className="amount">{percent(row.expenditure, total)}</td>
              <td><Bar part={row.expenditure} whole={largestMonth} /></td>
            </tr>
          ))}
        </tbody>
        <tfoot><tr><th scope="row">Total</th><td className="amount">{money.format(total)}</td><td className="amount">{percent(total, total)}</td><td /></tr></tfoot>
      </table>

      <h2>By category</h2>
      <p className="notice">Bars are relative to the largest category.</p>
      <table className="data-table" aria-label="Expenditure by category">
        <thead><tr><th>Category</th><th className="amount">Expenditure</th><th className="amount">Share</th><th aria-label="Bar compared with the largest category" /></tr></thead>
        <tbody>
          {data.categories.map((row) => (
            <tr key={row.name}>
              <th scope="row">{row.name}</th>
              <td className="amount">{money.format(row.expenditure)}</td>
              <td className="amount">{percent(row.expenditure, total)}</td>
              <td><Bar part={row.expenditure} whole={largestCategory} /></td>
            </tr>
          ))}
        </tbody>
        <tfoot><tr><th scope="row">Total</th><td className="amount">{money.format(total)}</td><td className="amount">{percent(total, total)}</td><td /></tr></tfoot>
      </table>

      <h2>By necessity</h2>
      <p className="notice">The dark part of the bar is essential.</p>
      <div className="dash-split" aria-hidden="true"><div style={{ width: `${essentialShare}%` }} /></div>
      <table className="data-table" aria-label="Expenditure by necessity">
        <thead><tr><th>Type</th><th className="amount">Expenditure</th><th className="amount">Share</th></tr></thead>
        <tbody>
          <tr><th scope="row">Essential</th><td className="amount">{money.format(data.essential.essential)}</td><td className="amount">{percent(data.essential.essential, total)}</td></tr>
          <tr><th scope="row">Non-essential</th><td className="amount">{money.format(data.essential.nonEssential)}</td><td className="amount">{percent(data.essential.nonEssential, total)}</td></tr>
        </tbody>
      </table>

      <details className="dash-details">
        <summary>Show category by month</summary>
        <div className="dash-scroll">
          <table className="data-table" aria-label="Expenditure by category and month">
            <thead>
              <tr><th>Category</th>{data.months.map((m) => <th key={m.month} className="amount">{monthLabel(m.month)}</th>)}<th className="amount">Total</th></tr>
            </thead>
            <tbody>
              {data.categoryByMonth.map((row) => (
                <tr key={row.category}>
                  <th scope="row">{row.category}</th>
                  {row.byMonth.map((amount, i) => <td key={data.months[i].month} className="amount">{money.format(amount)}</td>)}
                  <td className="amount">{money.format(row.total)}</td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr><th scope="row">Total</th>{data.months.map((m) => <td key={m.month} className="amount">{money.format(m.expenditure)}</td>)}<td className="amount">{money.format(total)}</td></tr>
            </tfoot>
          </table>
        </div>
      </details>

      <details className="dash-details">
        <summary>Show largest transactions</summary>
        <table className="data-table" aria-label="Largest transactions">
          <thead><tr><th>Date</th><th>Title</th><th>Category</th><th className="amount">Amount</th></tr></thead>
          <tbody>
            {data.largest.map((row, i) => (
              <tr key={i}><td>{row.date}</td><td className="dash-wrap">{row.title}</td><td>{row.category}</td><td className="amount">{money.format(row.amount)}</td></tr>
            ))}
          </tbody>
        </table>
      </details>

      <details className="dash-details">
        <summary>Show top merchants</summary>
        <p className="notice">Similar titles are grouped.</p>
        <table className="data-table" aria-label="Top merchants">
          <thead><tr><th>Merchant</th><th className="amount">Transactions</th><th className="amount">Total</th></tr></thead>
          <tbody>
            {data.merchants.map((row) => (
              <tr key={row.name}><th scope="row">{row.name}</th><td className="amount">{row.count}</td><td className="amount">{money.format(row.total)}</td></tr>
            ))}
          </tbody>
        </table>
      </details>

      <details className="dash-details">
        <summary>Show recurring charges</summary>
        <p className="notice">Charged in 3 or more months. Typical amount is the median.</p>
        {data.recurring.length === 0 ? (
          <p className="notice">None.</p>
        ) : (
          <table className="data-table" aria-label="Recurring charges">
            <thead><tr><th>Merchant</th><th className="amount">Typical amount</th><th className="amount">Months</th><th>Category</th></tr></thead>
            <tbody>
              {data.recurring.map((row) => (
                <tr key={row.name}><th scope="row">{row.name}</th><td className="amount">{money.format(row.typicalAmount)}</td><td className="amount">{row.months}</td><td>{row.category}</td></tr>
              ))}
            </tbody>
          </table>
        )}
      </details>

      <button className="dash-download" onClick={() => downloadCsv(transactions, categoryOf, essentialOf)}>Download data (CSV)</button>
      <ControlBar continueLabel="Edit data" onContinue={next} />
    </div>
  );
}
