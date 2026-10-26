import type { StatementInfo } from "../../types";

/** The sample statements the demo offers in place of the user's own files. */
export async function getSampleStatements(): Promise<StatementInfo[]> {
  const response = await fetch("/api/demo/statements");
  if (!response.ok) throw new Error(`/api/demo/statements returned ${response.status}`);
  return response.json() as Promise<StatementInfo[]>;
}
