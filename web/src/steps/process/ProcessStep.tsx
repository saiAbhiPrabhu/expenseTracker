import { useEffect, useState } from "react";
import { useStatements } from "../../demo/StatementsContext";
import ControlBar from "../../flow/ControlBar";
import { useFlow } from "../../flow/FlowContext";
import { money } from "../../format";
import { getProcessed, ProcessedStatement } from "./api";
import "./process.css";

const plural = (n: number, word: string) => `${n} ${word}${n === 1 ? "" : "s"}`;

function Summary({ s }: { s: ProcessedStatement }) {
  const read = s.pageReadings.filter((p) => p.status === "read").length;
  const skipped = s.pageReadings.length - read;
  const bad = s.sections.filter((x) => !x.matches);
  return (
    <ol className="process-list">
      <li><strong>Pages read:</strong> {plural(s.pages, "page")}: {read} read, {skipped} skipped (no transactions).</li>
      <li><strong>Private details removed:</strong> the top of each page that holds your name and account number is cut off before any text is read.</li>
      <li><strong>Transactions found:</strong> {plural(s.summary.rowsFound, "row")}; {plural(s.summary.depositsRemoved, "deposit")} taken out; {s.summary.spendingRows} kept.</li>
      <li>
        <strong>Totals checked:</strong> every section's printed total equals the sum of its rows (
        {bad.length === 0
          ? `all ${s.sections.length} sections match`
          : `${bad.map((x) => x.name).join(", ")} ${bad.length === 1 ? "does" : "do"} not match`}
        ).
      </li>
      <li><strong>Names and numbers hidden:</strong> titles are cleaned before they go to the AI.</li>
      <li><strong>Categories:</strong> {plural(s.summary.spendingRows, "transaction")} {s.summary.spendingRows === 1 ? "was" : "were"} sorted into categories; {s.summary.flaggedForReview} need your review.</li>
    </ol>
  );
}

function Details({ s }: { s: ProcessedStatement }) {
  const read = s.pageReadings.filter((p) => p.status === "read");
  return (
    <details>
      <summary>Show how this statement was read</summary>

      <h4>Pages, before and after the top is cut off</h4>
      <p className="notice">Left is the page as printed; right is the same page with the top cut off before any text is read.</p>
      {s.pageReadings.map((p) =>
        p.status === "read" ? (
          <div key={p.page}>
            <p>Page {p.page}{p.removedTopPercent !== undefined && ` (top ${p.removedTopPercent}% cut off)`}</p>
            <div className="process-pair">
              <figure>
                <img src={p.beforeImage} alt={`${s.label} page ${p.page} as printed, with your name and account number at the top`} />
                <figcaption className="notice">As printed</figcaption>
              </figure>
              <figure>
                <img src={p.afterImage} alt={`${s.label} page ${p.page} with the top cut off, name and account number gone`} />
                <figcaption className="notice">Top cut off</figcaption>
              </figure>
            </div>
          </div>
        ) : (
          <p key={p.page} className="notice">Page {p.page} skipped: {p.reason}</p>
        ),
      )}

      <h4>Text read from each page</h4>
      <p className="notice">This is the text the reader found on each page, after the top was cut off.</p>
      {read.map((p) => (
        <div key={p.page}>
          <p>Page {p.page}</p>
          <pre className="process-text" aria-label={`Text read from ${s.label} page ${p.page}`} tabIndex={0}>{p.ocrText}</pre>
        </div>
      ))}

      <h4>Section totals</h4>
      <p className="notice">Each section's printed total is compared with the sum of its rows.</p>
      <div className="process-scroll">
        <table className="data-table" aria-label={`Section totals for ${s.label}`}>
          <thead><tr><th>Section</th><th>Rows</th><th>Printed total</th><th>Sum of rows</th><th>Matches</th></tr></thead>
          <tbody>
            {s.sections.map((x) => (
              <tr key={x.name}>
                <td>{x.name}</td>
                <td className="amount">{x.rows}</td>
                <td className="amount">{money.format(x.printedTotal)}</td>
                <td className="amount">{money.format(x.rowsTotal)}</td>
                <td>{x.matches ? "Yes" : "No"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <h4>Every row</h4>
      <p className="notice">Each line found on the statement, the title as printed, and the title sent to the AI. Removed rows are deposits.</p>
      <div className="process-scroll">
        <table className="data-table" aria-label={`Every row of ${s.label}`}>
          <thead><tr><th>Date</th><th>Title as printed</th><th>Title sent to the AI</th><th>Amount</th><th>Section</th><th>Status</th></tr></thead>
          <tbody>
            {s.rows.map((r, i) => (
              <tr key={i}>
                <td>{r.date}</td>
                <td>{r.title}</td>
                <td>{r.scrubbedTitle}</td>
                <td className="amount">{money.format(r.amount)}</td>
                <td>{r.section}</td>
                <td>{r.kept ? "Kept" : "Removed"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </details>
  );
}

// Step 2: show what was done to each added statement, with the real numbers.
export default function ProcessStep() {
  const { next } = useFlow();
  const { statements } = useStatements();
  const [results, setResults] = useState<ProcessedStatement[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (statements.length === 0) return;
    let current = true;
    setResults(null);
    setError(null);
    const ordered = [...statements].sort((a, b) => a.id.localeCompare(b.id));
    Promise.all(ordered.map((s) => getProcessed(s.id)))
      .then((r) => current && setResults(r))
      .catch(() => current && setError("Could not read your statements. Press Back to return to Upload and try again."));
    return () => { current = false; };
  }, [statements]);

  if (statements.length === 0) {
    return (
      <div>
        <p className="notice">No statements were added. Press Back to return to Upload and add them.</p>
        <ControlBar continueLabel="Continue" onContinue={() => {}} disabledReason="Go back to Upload and add statements first." />
      </div>
    );
  }

  const reason = error ?? (results ? undefined : "Reading your statements…");
  return (
    <div>
      <p className="notice">This is what was done to each statement before it reached your dashboard.</p>
      {error && <p className="notice">{error}</p>}
      {!error && !results && <p className="notice">Reading your statements…</p>}
      {results && results.map((s) => (
        <div key={s.id} className="process-block">
          <h3>{s.label}</h3>
          <p className="notice">{s.fileName}</p>
          <Summary s={s} />
          <Details s={s} />
        </div>
      ))}
      <ControlBar continueLabel="Continue" onContinue={next} disabledReason={reason} />
    </div>
  );
}
