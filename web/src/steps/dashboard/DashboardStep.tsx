import { useEffect, useState } from "react";
import { useDemo } from "../../demo/DemoContext";
import { monthLabel } from "../../demo/months";
import ControlBar from "../../flow/ControlBar";
import { useFlow } from "../../flow/FlowContext";
import { money, percent } from "../../format";
import { DashboardData, postDashboard } from "./api";
import { downloadCsv } from "./csv";
import "./dashboard.css";

const sign = (amount: number) => (amount > 0 ? "+" : "") + money.format(amount);
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
        Expenditure is money that left the account (any negative amount).
        {changed > 0 && ` Includes your ${plural(changed, "change")}.`}
      </p>
      <p className="dash-total" aria-live="polite">
        Total expenditure: <strong>{money.format(total)}</strong> across {plural(data.transactionCount, "transaction")}
      </p>
      <p>{longDate(data.firstDate)} to {longDate(data.lastDate)}</p>
      <label className="dash-check">
        <input type="checkbox" checked={includeMoneyMovement} onChange={(event) => setIncludeMoneyMovement(event.target.checked)} />
        {" "}Include transfers, savings and debt payments
      </label>
      <p className="notice">Moving money between accounts or paying off debt is not everyday spending; unchecking leaves it out of every number below.</p>
      <p>Money movement: <strong>{money.format(data.moneyMovement)}</strong>{includeMoneyMovement ? " (counted in the numbers below)" : " (left out of the numbers below)"}</p>

      <h3>By month</h3>
      <p className="notice">See the trend. Bars are drawn against the largest month.</p>
      <table className="data-table" aria-label="Expenditure by month">
        <thead><tr><th>Month</th><th>Expenditure</th><th>Change</th><th>Change %</th><th>Share</th><th aria-label="Bar compared with the largest month" /></tr></thead>
        <tbody>
          {data.months.map((row, i) => {
            const previous = i > 0 ? data.months[i - 1].expenditure : null;
            return (
              <tr key={row.month}>
                <th scope="row">{monthLabel(row.month)}</th>
                <td className="amount">{money.format(row.expenditure)}</td>
                <td className="amount">{previous === null ? "-" : sign(row.expenditure - previous)}</td>
                <td className="amount">{previous === null || previous === 0 ? "-" : `${row.expenditure >= previous ? "+" : ""}${(((row.expenditure - previous) / previous) * 100).toFixed(1)}%`}</td>
                <td className="amount">{percent(row.expenditure, total)}</td>
                <td><Bar part={row.expenditure} whole={largestMonth} /></td>
              </tr>
            );
          })}
        </tbody>
        <tfoot><tr><th scope="row">Total</th><td className="amount">{money.format(total)}</td><td /><td /><td className="amount">{percent(total, total)}</td><td /></tr></tfoot>
      </table>

      <h3>By category</h3>
      <p className="notice">See where the money goes. Bars are drawn against the largest category.</p>
      <table className="data-table" aria-label="Expenditure by category">
        <thead><tr><th>Category</th><th>Expenditure</th><th>Share</th><th aria-label="Bar compared with the largest category" /></tr></thead>
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

      <h3>Essential or not</h3>
      <p className="notice">See what is optional. The dark part of the bar is essential spending.</p>
      <div className="dash-split" aria-hidden="true"><div style={{ width: `${essentialShare}%` }} /></div>
      <table className="data-table" aria-label="Essential versus non-essential">
        <thead><tr><th>Type</th><th>Expenditure</th><th>Share</th></tr></thead>
        <tbody>
          <tr><th scope="row">Essential</th><td className="amount">{money.format(data.essential.essential)}</td><td className="amount">{percent(data.essential.essential, total)}</td></tr>
          <tr><th scope="row">Non-essential</th><td className="amount">{money.format(data.essential.nonEssential)}</td><td className="amount">{percent(data.essential.nonEssential, total)}</td></tr>
        </tbody>
      </table>

      <details className="dash-details">
        <summary>Show category by month</summary>
        <p className="notice">Every category in every month, to spot when a category grew or shrank.</p>
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
        <summary>Show top merchants</summary>
        <p className="notice">The ten places you spent the most in total, with similar statement titles grouped together.</p>
        <table className="data-table" aria-label="Top merchants">
          <thead><tr><th>Merchant</th><th>Transactions</th><th>Total</th></tr></thead>
          <tbody>
            {data.merchants.map((row) => (
              <tr key={row.name}><th scope="row">{row.name}</th><td className="amount">{row.count}</td><td className="amount">{money.format(row.total)}</td></tr>
            ))}
          </tbody>
        </table>
      </details>

      <details className="dash-details">
        <summary>Show largest transactions</summary>
        <p className="notice">The ten biggest single payments, in case one looks wrong or can be avoided.</p>
        <table className="data-table" aria-label="Largest transactions">
          <thead><tr><th>Date</th><th>Title</th><th>Category</th><th>Amount</th></tr></thead>
          <tbody>
            {data.largest.map((row, i) => (
              <tr key={i}><td>{row.date}</td><td className="dash-wrap">{row.title}</td><td>{row.category}</td><td className="amount">{money.format(row.amount)}</td></tr>
            ))}
          </tbody>
        </table>
      </details>

      <details className="dash-details">
        <summary>Show recurring charges</summary>
        <p className="notice">Merchants charged in at least 3 different months, with their typical (median) amount, to spot rent and subscriptions.</p>
        {data.recurring.length === 0 ? (
          <p className="notice">No merchant appears in 3 or more months.</p>
        ) : (
          <table className="data-table" aria-label="Recurring charges">
            <thead><tr><th>Merchant</th><th>Typical amount</th><th>Months</th><th>Category</th></tr></thead>
            <tbody>
              {data.recurring.map((row) => (
                <tr key={row.name}><th scope="row">{row.name}</th><td className="amount">{money.format(row.typicalAmount)}</td><td className="amount">{row.months}</td><td>{row.category}</td></tr>
              ))}
            </tbody>
          </table>
        )}
      </details>

      <button className="dash-download" onClick={() => downloadCsv(transactions, categoryOf, essentialOf)}>Download data (CSV)</button>
      <p className="notice">Every transaction with the category and essential flag in use after your edits.</p>
      <ControlBar continueLabel="Edit data" onContinue={next} />
    </div>
  );
}
