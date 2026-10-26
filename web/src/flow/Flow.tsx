import { useEffect, useRef } from "react";
import DashboardStep from "../steps/dashboard/DashboardStep";
import EditStep from "../steps/edit/EditStep";
import ProcessStep from "../steps/process/ProcessStep";
import ReviewStep from "../steps/review/ReviewStep";
import UploadStep from "../steps/upload/UploadStep";
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

// The step list, the current step's title and instruction, and the current step itself.
export default function Flow() {
  const { current, index, reached, goTo } = useFlow();
  const View = VIEWS[current];
  const heading = useRef<HTMLHeadingElement>(null);
  const firstRender = useRef(true);
  // When the step changes, keyboard focus moves to the new step's title instead of being lost.
  useEffect(() => {
    if (firstRender.current) {
      firstRender.current = false;
      return;
    }
    heading.current?.focus();
  }, [current]);
  return (
    <div>
      <nav aria-label="Steps">
      <ol className="stepper">
        {STEPS.map((step, i) => (
          <li key={step.id} className={step.id === current ? "current" : ""} aria-current={step.id === current ? "step" : undefined}>
            {reached.has(step.id) && step.id !== current ? (
              <button className="stepper-link" onClick={() => goTo(step.id)}>{i + 1}. {step.title}</button>
            ) : (
              <>{i + 1}. {step.title}</>
            )}
          </li>
        ))}
      </ol>
      </nav>
      <h2 ref={heading} tabIndex={-1}>Step {index + 1} of {STEPS.length}: {STEPS[index].title}</h2>
      <p className="notice">{STEPS[index].instruction}</p>
      <div className="panel"><View /></div>
    </div>
  );
}
