export type StepId = "upload" | "process" | "review" | "dashboard" | "edit";

export interface StepInfo {
  id: StepId;
  title: string;
  instruction: string; // one line, shown under the step title
}

// The order is the order of the flow. The first time through, the user moves forward along it.
export const STEPS: StepInfo[] = [
  { id: "upload", title: "Upload", instruction: "Add one statement per month." },
  { id: "process", title: "Process", instruction: "Your statements are read automatically." },
  { id: "review", title: "Review", instruction: "Check the categories we were unsure about." },
  { id: "dashboard", title: "Dashboard", instruction: "Where your money went. Press Edit data to change anything." },
  { id: "edit", title: "Edit data", instruction: "Change a category or essential flag, then press Done." },
];
