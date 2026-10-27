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
      <h2>Review: {current.label}</h2>
      <p className="notice">
        Month {page + 1} of {pages.length}. {current.rows.length === 0
          ? "Nothing to review."
          : `${current.rows.length} row${current.rows.length === 1 ? "" : "s"} to review.`}
      </p>
      {current.rows.length > 0 && (
        <table className="review-table">
          <thead>
            <tr>
              <th>Date</th>
              <th className="amount">Amount</th>
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
                  <td>{row.date}</td>
                  <td className="amount">{money.format(row.amount)}</td>
                  <td>{row.title}</td>
                  <td>
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
                  <td>
                    <button onClick={() => undoCategory(row)} disabled={!edited} aria-label={`Undo change for ${row.title}`}>
                      Undo
                    </button>
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
