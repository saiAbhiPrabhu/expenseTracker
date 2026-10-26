import { useMemo } from "react";
import { useDemo } from "../../demo/DemoContext";
import { reviewPages } from "../../demo/months";
import ControlBar from "../../flow/ControlBar";
import { useFlow } from "../../flow/FlowContext";
import { money } from "../../format";

// One page per month listing the rows JEV was unsure about. Only the category can be changed.
export default function ReviewStep() {
  const { next } = useFlow();
  const { transactions, categories, categoryOf, setCategory, undoCategory, reviewPage, setReviewPage } = useDemo();
  const pages = useMemo(() => reviewPages(transactions), [transactions]);
  if (pages.length === 0) {
    return (
      <div>
        <p className="notice">No transactions to review.</p>
        <ControlBar continueLabel="Finish" onContinue={next} />
      </div>
    );
  }
  const page = Math.min(reviewPage, pages.length - 1);
  const current = pages[page];
  const isLast = page === pages.length - 1;

  return (
    <div>
      <div className="tab-row" role="group" aria-label="Months">
        {pages.map((p, i) => (
          <button key={p.key} className="tab" aria-pressed={i === page} onClick={() => setReviewPage(i)}>
            {p.label.split(" ")[0]}{p.rows.length > 0 && <span className="tab-count"> {p.rows.length}</span>}
          </button>
        ))}
      </div>
      <h3>Review: {current.label}</h3>
      <p className="notice">
        Month {page + 1} of {pages.length}. {current.rows.length === 0
          ? "Nothing to review."
          : `${current.rows.length} row${current.rows.length === 1 ? "" : "s"} to review.`}
      </p>
      {current.rows.length > 0 && (
        <table className="review-table" aria-label={`Rows to review for ${current.label}`}>
          <thead>
            <tr>
              <th>Date</th>
              <th>Amount</th>
              <th>Title</th>
              <th>Category</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {current.rows.map((row) => {
              const category = categoryOf(row);
              const edited = category !== row.category20.choice;
              return (
                <tr key={row.key} className={edited ? "edited" : ""}>
                  <td className="rv-date">{row.date}</td>
                  <td className="amount rv-amount">{money.format(row.amount)}</td>
                  <td className="rv-title">{row.title}</td>
                  <td className="rv-cat">
                    <select
                      value={category}
                      aria-label={`Category for ${row.title}`}
                      onChange={(event) => setCategory(row, event.target.value)}
                    >
                      {categories.map((name) => (
                        <option key={name} value={name}>{name}</option>
                      ))}
                    </select>
                  </td>
                  <td className="rv-undo">
                    {edited && (
                      <button onClick={() => undoCategory(row)} aria-label={`Undo change for ${row.title}`}>
                        Undo
                      </button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
      <ControlBar
        continueLabel={isLast ? "Finish" : "Continue"}
        onContinue={isLast ? next : () => setReviewPage(page + 1)}
        onBack={page > 0 ? () => setReviewPage(page - 1) : undefined}
      />
    </div>
  );
}
