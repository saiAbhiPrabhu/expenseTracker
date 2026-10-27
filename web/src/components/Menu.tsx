import { Link, NavLink } from "react-router-dom";

// The top bar: the name, and the three places the site has. Always visible, so nothing is folded away.
export default function Menu() {
  return (
    <header className="site-bar">
      <Link to="/" className="wordmark">Expense Tracker</Link>
      <nav aria-label="Site">
        <NavLink to="/" end>Home</NavLink>
        <NavLink to="/demo">Demo</NavLink>
        <NavLink to="/login">Log in</NavLink>
      </nav>
    </header>
  );
}
