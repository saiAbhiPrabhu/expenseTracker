import { createPortal } from "react-dom";
import { useActionSlot } from "./ActionSlot";
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

// The one place a step puts its Back and primary buttons. It renders into the page's sticky action bar,
// so every step's primary action sits in the same place.
export default function ControlBar({ continueLabel, onContinue, disabledReason, onBack, hideBack }: Props) {
  const { index, back } = useFlow();
  const slot = useActionSlot();
  const showBack = !hideBack && (onBack !== undefined || index > 0);
  if (!slot) return null;
  return createPortal(
    <div className="button-row control-bar">
      {showBack && <button onClick={onBack ?? back}>Back</button>}
      <button onClick={onContinue} disabled={disabledReason !== undefined}>{continueLabel}</button>
      {disabledReason !== undefined && <span className="notice">{disabledReason}</span>}
    </div>,
    slot,
  );
}
