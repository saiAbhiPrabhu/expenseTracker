export type StepId = "upload" | "process" | "review" | "dashboard" | "edit";

export interface StepInfo {
  id: StepId;
  title: string;
  instruction: string; // one line, shown under the step title
}

// The order is the order of the flow. The first time through, the user moves forward along it.
export const STEPS: StepInfo[] = [
  { id: "upload", title: "Upload", instruction: "Add your monthly statements, or use the sample statements." },
  { id: "process", title: "Process", instruction: "We read each statement and pull out the transactions. Nothing for you to do." },
  { id: "review", title: "Review", instruction: "Check the categories we were not sure about, one month at a time." },
  { id: "dashboard", title: "Dashboard", instruction: "See where your money went. Use Edit data to fix anything that looks wrong." },
  { id: "edit", title: "Edit data", instruction: "Change a category or the essential flag. Press Done to return to the dashboard." },
];
