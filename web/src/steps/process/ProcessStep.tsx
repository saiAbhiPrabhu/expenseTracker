import { useEffect, useState } from "react";
import { useStatements } from "../../demo/StatementsContext";
import ControlBar from "../../flow/ControlBar";
import { useFlow } from "../../flow/FlowContext";
import { money } from "../../format";
import { getProcessed, ProcessedStatement } from "./api";
import "./process.css";

function Summary({ s }: { s: ProcessedStatement }) {
  const parsed = s.pageReadings.filter((p) => p.status === "read").length;
  return (
    <ul className="process-list">
      <li>Pages: {s.pages} total, {parsed} parsed, {s.pages - parsed} skipped</li>
      <li>Transactions: {s.summary.rowsFound} total rows, {s.summary.depositsRemoved} removed rows, {s.summary.spendingRows} kept rows</li>
      <li>Categories: {s.summary.flaggedForReview} under review</li>
    </ul>
  );
}

type Tab = "pages" | "text" | "rows";
const TABS: { id: Tab; label: string }[] = [
  { id: "pages", label: "Pages" },
  { id: "text", label: "Text from pages" },
  { id: "rows", label: "Rows" },
];

// Three tabs per statement. Pressing the open tab again closes it, so the page stays short by default.
function Details({ s }: { s: ProcessedStatement }) {
  const [tab, setTab] = useState<Tab | null>(null);
  const parsed = s.pageReadings.filter((p) => p.status === "read");
  return (
    <div>
      <div className="tab-row" role="group" aria-label={`Details of ${s.label}`}>
        {TABS.map((t) => (
          <button key={t.id} className="tab" aria-expanded={tab === t.id} aria-controls={`${s.id}-${t.id}`} onClick={() => setTab(tab === t.id ? null : t.id)}>
            {t.label}
          </button>
        ))}
      </div>

      {tab === "pages" && (
        <div id={`${s.id}-pages`} className="tab-panel">
          {s.pageReadings.map((p) =>
            p.status === "read" ? (
              <div key={p.page}>
                <p>Page {p.page}</p>
                <div className="process-pair">
                  <figure>
                    <img loading="lazy" src={p.beforeImage} alt={`${s.label} page ${p.page} as printed, with your name and account number at the top`} />
                    <figcaption className="notice">As printed</figcaption>
                  </figure>
                  <figure>
                    <img loading="lazy" src={p.afterImage} alt={`${s.label} page ${p.page} with the top cut off, name and account number gone`} />
                    <figcaption className="notice">Top cut off</figcaption>
                  </figure>
                </div>
              </div>
            ) : (
              <p key={p.page} className="notice">Page {p.page}: skipped</p>
            ),
          )}
        </div>
      )}

      {tab === "text" && (
        <div id={`${s.id}-text`} className="tab-panel">
          {parsed.map((p) => (
            <div key={p.page}>
              <p>Page {p.page}</p>
              <pre className="process-text" aria-label={`Text read from ${s.label} page ${p.page}`} tabIndex={0}>{p.ocrText}</pre>
            </div>
          ))}
        </div>
      )}

      {tab === "rows" && (
        <div id={`${s.id}-rows`} className="tab-panel">
          <table className="data-table process-rows" aria-label={`Rows of ${s.label}`}>
            <thead><tr><th>Date</th><th>Title</th><th>Cleaned title</th><th>Amount</th><th>Status</th></tr></thead>
            <tbody>
              {s.rows.map((r, i) => (
                <tr key={i}>
                  <td data-label="Date">{r.date}</td>
                  <td data-label="Title">{r.title}</td>
                  <td data-label="Cleaned title">{r.scrubbedTitle}</td>
                  <td className="amount" data-label="Amount">{money.format(r.amount)}</td>
                  <td data-label="Status">{r.kept ? "Kept" : "Removed"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

// Step 2: what was done to each added statement, with the real numbers.
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
      .catch(() => current && setError("Could not read your statements. Press Back and try again."));
    return () => { current = false; };
  }, [statements]);

  if (statements.length === 0) {
    return (
      <div>
        <p className="notice">No statements added.</p>
        <ControlBar continueLabel="Continue" onContinue={() => {}} disabledReason="Add statements on the Upload step." />
      </div>
    );
  }

  const reason = error ?? (results ? undefined : "Reading your statements…");
  return (
    <div>
      {error && <p className="notice error"><span className="error-label">Error</span>{error}</p>}
      {!error && !results && <p className="notice">Reading your statements…</p>}
      {results && results.map((s) => (
        <div key={s.id} className="process-block">
          <h3>{s.label}</h3>
          <p className="notice file-name">{s.fileName}</p>
          <Summary s={s} />
          <Details s={s} />
        </div>
      ))}
      <ControlBar continueLabel="Continue" onContinue={next} disabledReason={reason} />
    </div>
  );
}
