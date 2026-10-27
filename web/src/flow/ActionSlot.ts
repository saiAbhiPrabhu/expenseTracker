import { createContext, useContext } from "react";

// The place in the control column where the current step's Back and primary buttons are drawn.
export const ActionSlot = createContext<HTMLElement | null>(null);
export const useActionSlot = () => useContext(ActionSlot);
