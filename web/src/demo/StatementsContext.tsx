import { createContext, ReactNode, useContext, useState } from "react";
import type { StatementInfo } from "../types";

interface StatementsState {
  /** The statements the user added in the Upload step. The Process step reads them. */
  statements: StatementInfo[];
  setStatements: (statements: StatementInfo[]) => void;
}

const StatementsContext = createContext<StatementsState | null>(null);

export function useStatements(): StatementsState {
  const state = useContext(StatementsContext);
  if (!state) throw new Error("useStatements must be used inside StatementsProvider");
  return state;
}

export function StatementsProvider({ children }: { children: ReactNode }) {
  const [statements, setStatements] = useState<StatementInfo[]>([]);
  return <StatementsContext.Provider value={{ statements, setStatements }}>{children}</StatementsContext.Provider>;
}
