import { DemoProvider } from "../demo/DemoContext";
import { StatementsProvider } from "../demo/StatementsContext";
import Flow from "../flow/Flow";
import { FlowProvider } from "../flow/FlowContext";

// The demo is the same flow a logged-in session will use; only the data source differs.
export default function Demo() {
  return (
    <section>
      <h1>Demo</h1>
      <p className="notice">Sample data only.</p>
      <DemoProvider>
        <StatementsProvider>
          <FlowProvider>
            <Flow />
          </FlowProvider>
        </StatementsProvider>
      </DemoProvider>
    </section>
  );
}
