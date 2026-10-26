import { useState } from "react";
import { useStatements } from "../../demo/StatementsContext";
import ControlBar from "../../flow/ControlBar";
import { useFlow } from "../../flow/FlowContext";
import { getSampleStatements } from "./api";

const megabytes = (bytes: number) => `${(bytes / 1_000_000).toFixed(2)} MB`;

// Step 1: choose the statements. In the demo the files come from the sample set instead of the user's computer.
export default function UploadStep() {
  const { next } = useFlow();
  const { statements, setStatements } = useStatements();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const useSamples = () => {
    setLoading(true);
    setError(null);
    getSampleStatements()
      .then(setStatements)
      .catch(() => setError("Could not load the sample statements. Try again."))
      .finally(() => setLoading(false));
  };

  return (
    <div>
      <div className="dropzone">
      <p>
        <label>
          Statement PDFs{" "}
          <input type="file" accept="application/pdf" multiple disabled aria-describedby="upload-off" />
        </label>{" "}
        <span id="upload-off" className="notice">Disabled in the demo.</span>
      </p>
      <p>
        <button onClick={loading ? undefined : useSamples} aria-disabled={loading}>Use the sample statements</button>
        {loading && <span className="notice"> Loading…</span>}
      </p>
      </div>
      {error && <p className="notice error"><span className="error-label">Error</span>{error}</p>}
      <table className="data-table" aria-label="Your statements">
        <thead>
          <tr><th>File</th><th>Month</th><th>Pages</th><th>Size</th></tr>
        </thead>
        <tbody>
          {statements.length === 0 && (
            <tr><td colSpan={4} className="notice">No statements added yet.</td></tr>
          )}
          {statements.map((statement) => (
            <tr key={statement.id}>
              <td><a href={`/api/demo/statements/${statement.id}/file`} target="_blank" rel="noreferrer">{statement.fileName}</a></td>
              <td>{statement.label}</td>
              <td className="amount">{statement.pages}</td>
              <td className="amount">{megabytes(statement.sizeBytes)}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <ControlBar
        continueLabel="Continue"
        onContinue={next}
        disabledReason={statements.length === 0 ? "Add the sample statements to continue." : undefined}
      />
    </div>
  );
}
