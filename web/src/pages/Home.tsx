import { Link } from "react-router-dom";

export default function Home() {
  return (
    <section className="center">
      <h1>Expense Tracker</h1>
      <div className="button-row">
        <Link className="button" to="/login">LOG IN</Link>
        <Link className="button" to="/demo">DEMO</Link>
      </div>
    </section>
  );
}
