import { useFlow } from "./FlowContext";

interface Props {
  continueLabel: string; // "Continue", "Finish", "Edit data", "Done"
  onContinue: () => void;
  /** When set, Continue is disabled and this says why. */
  disabledReason?: string;
  /** Replaces the default Back (the previous step), for example to go to the previous month. */
  onBack?: () => void;
  /** For a step whose only way out is the primary button. */
  hideBack?: boolean;
}

// The one place a step puts its Back and primary buttons. Every step uses it, so the order of events is always clear.
export default function ControlBar({ continueLabel, onContinue, disabledReason, onBack, hideBack }: Props) {
  const { index, back } = useFlow();
  const showBack = !hideBack && (onBack !== undefined || index > 0);
  return (
    <div className="button-row control-bar">
      {showBack && <button onClick={onBack ?? back}>Back</button>}
      <button onClick={onContinue} disabled={disabledReason !== undefined}>{continueLabel}</button>
      {disabledReason !== undefined && <span className="notice">{disabledReason}</span>}
    </div>
  );
}
