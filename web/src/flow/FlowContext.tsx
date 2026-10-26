import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { STEPS, StepId } from "./steps";

interface FlowState {
  current: StepId;
  index: number; // position of the current step in STEPS
  reached: ReadonlySet<StepId>;
  /** Move to the next step in order. */
  next: () => void;
  /** Move to the previous step in order. */
  back: () => void;
  /** Jump to a step the user has already reached. Ignored for steps not reached yet. */
  goTo: (id: StepId) => void;
}

const FlowContext = createContext<FlowState | null>(null);

export function useFlow(): FlowState {
  const state = useContext(FlowContext);
  if (!state) throw new Error("useFlow must be used inside FlowProvider");
  return state;
}

// The current step lives in the browser history (location state), so the browser's Back and Forward
// buttons move between steps. A step that has not been reached in this visit falls back to Upload,
// which is what a reload shows, because the progress is held in memory.
export function FlowProvider({ children }: { children: ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const [reached, setReached] = useState<ReadonlySet<StepId>>(new Set([STEPS[0].id]));
  const requested = (location.state as { step?: StepId } | null)?.step;
  const current: StepId = requested !== undefined && reached.has(requested) ? requested : STEPS[0].id;
  const index = STEPS.findIndex((step) => step.id === current);

  const move = useCallback(
    (id: StepId) => {
      if (id === current) return;
      setReached((previous) => new Set(previous).add(id));
      navigate(location.pathname, { state: { step: id } });
    },
    [current, navigate, location.pathname],
  );

  // Once the user has moved past Upload, a reload or closing the tab would lose their progress.
  useEffect(() => {
    if (reached.size < 2) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [reached]);

  const value = useMemo<FlowState>(
    () => ({
      current,
      index,
      reached,
      next: () => move(STEPS[Math.min(index + 1, STEPS.length - 1)].id),
      back: () => move(STEPS[Math.max(index - 1, 0)].id),
      goTo: (id) => reached.has(id) && move(id),
    }),
    [current, index, reached, move],
  );
  return <FlowContext.Provider value={value}>{children}</FlowContext.Provider>;
}
