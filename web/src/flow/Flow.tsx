import { useEffect, useRef, useState } from "react";
import DashboardStep from "../steps/dashboard/DashboardStep";
import EditStep from "../steps/edit/EditStep";
import ProcessStep from "../steps/process/ProcessStep";
import ReviewStep from "../steps/review/ReviewStep";
import UploadStep from "../steps/upload/UploadStep";
import { ActionSlot } from "./ActionSlot";
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

// The control column (step keys and the current step's buttons) beside the working surface.
export default function Flow() {
  const { current, index, reached, goTo } = useFlow();
  const View = VIEWS[current];
  const heading = useRef<HTMLHeadingElement>(null);
  const firstRender = useRef(true);
  const [slot, setSlot] = useState<HTMLElement | null>(null);
  // When the step changes, keyboard focus moves to the new step's title instead of being lost.
  useEffect(() => {
    if (firstRender.current) {
      firstRender.current = false;
      return;
    }
    heading.current?.focus();
  }, [current]);
  return (
    <div className="unit">
      <aside className="control">
        <nav aria-label="Steps">
          <ol className="keys">
            {STEPS.map((step, i) => {
              const label = (
                <>
                  <span className="key-number">{i + 1}</span>
                  <span className="key-title">{step.title}</span>
                </>
              );
              const isCurrent = step.id === current;
              return (
                <li key={step.id} aria-current={isCurrent ? "step" : undefined}>
                  {reached.has(step.id) && !isCurrent ? (
                    <button className="key" onClick={() => goTo(step.id)}>{label}</button>
                  ) : (
                    <span className={isCurrent ? "key key-current" : "key key-off"}>{label}</span>
                  )}
                </li>
              );
            })}
          </ol>
        </nav>
        <div className="control-actions" ref={setSlot} />
        <p className="control-note">Sample data only.</p>
      </aside>
      <main className="surface">
        <h1 ref={heading} tabIndex={-1}>{STEPS[index].title}</h1>
        <p className="lead">{STEPS[index].instruction}</p>
        <ActionSlot.Provider value={slot}>
          <View />
        </ActionSlot.Provider>
      </main>
    </div>
  );
}
