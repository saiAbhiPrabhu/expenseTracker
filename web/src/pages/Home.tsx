import { Link } from "react-router-dom";
import { useLedgerBody } from "../ledger/useLedgerBody";

// The front door: two equal choices on an almost empty page. Real login is not built, so LOG IN is shown disabled.
export default function Home() {
  useLedgerBody();
  return (
    <section className="home">
      <h1 className="sr-only">Expense Tracker</h1>
      <div className="choices">
        <div className="choice">
          <button className="ledger-button" disabled aria-describedby="login-note">Log in</button>
          <p className="choice-note" id="login-note">Not available yet.</p>
        </div>
        <div className="choice">
          <Link className="ledger-button" to="/demo">Try the demo</Link>
        </div>
      </div>
      <p className="privacy">Your name and account number never leave your machine.</p>
    </section>
  );
}
