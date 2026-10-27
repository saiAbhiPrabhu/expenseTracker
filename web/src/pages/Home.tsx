import { Link } from "react-router-dom";

export default function Home() {
  return (
    <main className="surface narrow">
      <h1>Expense Tracker</h1>
      <p className="lead">Reads bank statements and shows where the money went. The demo uses sample data only.</p>
      <div className="button-row">
        <Link className="button primary" to="/demo">Start demo</Link>
        <Link className="button" to="/login">Log in</Link>
      </div>
    </main>
  );
}
