import { useEffect, useRef, useState } from "react";
import { useDemo } from "../demo/DemoContext";
import { useStatements } from "../demo/StatementsContext";
import DashboardStep from "../steps/dashboard/DashboardStep";
import EditStep from "../steps/edit/EditStep";
import ProcessStep from "../steps/process/ProcessStep";
import ReviewStep from "../steps/review/ReviewStep";
import UploadStep from "../steps/upload/UploadStep";
import { ActionSlotContext } from "./ActionSlot";
import { useFlow } from "./FlowContext";
import { STEPS, StepId } from "./steps";
import "./flow.css";

const VIEWS: Record<StepId, () => JSX.Element> = {
  upload: UploadStep,
  process: ProcessStep,
  review: ReviewStep,
  dashboard: DashboardStep,
  edit: EditStep,
};

// The left page: the steps, and the counts that follow the user from step to step.
function Rail() {
  const { current, reached, goTo } = useFlow();
  const { statements } = useStatements();
  const { transactions, edits } = useDemo();
  const toReview = transactions.filter((t) => t.needsReview).length;
  const changed = Object.keys(edits).length;
  return (
    <aside className="rail">
      <div className="rail-inner">
      <nav aria-label="Steps">
        <ol className="rail-steps">
          {STEPS.map((step, i) => (
            <li key={step.id} className={step.id === current ? "current" : ""} aria-current={step.id === current ? "step" : undefined}>
              {reached.has(step.id) && step.id !== current ? (
                <button className="stepper-link" onClick={() => goTo(step.id)}>{i + 1}. {step.title}</button>
              ) : (
                <span className="rail-label">{i + 1}. {step.title}</span>
              )}
            </li>
          ))}
        </ol>
      </nav>
      <dl className="rail-facts">
        <div><dt>Statements</dt><dd>{statements.length}</dd></div>
        <div className={toReview > 0 ? "attention" : ""}><dt>Rows to review</dt><dd>{toReview}</dd></div>
        <div><dt>Rows changed</dt><dd>{changed}</dd></div>
      </dl>
      </div>
    </aside>
  );
}

// The two-page spread: the rail on the left, the current step on the right, and the action bar pinned below.
export default function Flow() {
  const { current, index } = useFlow();
  const View = VIEWS[current];
  const heading = useRef<HTMLHeadingElement>(null);
  const firstRender = useRef(true);
  const [actionSlot, setActionSlot] = useState<HTMLElement | null>(null);
  // When the step changes, keyboard focus moves to the new step's title instead of being lost.
  useEffect(() => {
    if (firstRender.current) {
      firstRender.current = false;
      return;
    }
    heading.current?.focus();
  }, [current]);
  return (
    <div className="spread">
      <Rail />
      <div className="page">
        <h2 ref={heading} tabIndex={-1}>Step {index + 1} of {STEPS.length}: {STEPS[index].title}</h2>
        <p className="notice step-instruction">{STEPS[index].instruction}</p>
        <ActionSlotContext.Provider value={actionSlot}>
          <div className="page-body"><View /></div>
        </ActionSlotContext.Provider>
        <div className="action-bar" ref={setActionSlot} />
      </div>
    </div>
  );
}
