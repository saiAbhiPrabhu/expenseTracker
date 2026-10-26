import { createContext, useContext } from "react";

// The sticky bar at the bottom of the page. Flow provides it; ControlBar renders into it, so each step keeps
// declaring its own Back and primary buttons while the bar stays in one place.
export const ActionSlotContext = createContext<HTMLElement | null>(null);
export const useActionSlot = () => useContext(ActionSlotContext);
