import { useState } from "react";
import { DashboardStep, EditStep, ProcessStep, ReviewStep, UploadStep } from "../demo/steps";
import { DEMO_STEPS } from "../types";

const STEP_VIEWS = [UploadStep, ProcessStep, ReviewStep, DashboardStep, EditStep];

export default function Demo() {
  const [step, setStep] = useState(0);
  const View = STEP_VIEWS[step];
  return (
    <section>
      <h1>Demo</h1>
      <p className="notice">Runs on a made-up sample statement, never on real data.</p>
      <ol className="stepper">
        {DEMO_STEPS.map((name, index) => (
          <li key={name} className={index === step ? "current" : ""}>
            {index + 1}. {name}
          </li>
        ))}
      </ol>
      <div className="panel">
        <View />
      </div>
      <div className="button-row">
        <button onClick={() => setStep(step - 1)} disabled={step === 0}>Back</button>
        <button onClick={() => setStep(step + 1)} disabled={step === DEMO_STEPS.length - 1}>Next</button>
      </div>
    </section>
  );
}
