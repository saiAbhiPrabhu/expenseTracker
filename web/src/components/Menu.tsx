import { useState } from "react";
import { Link } from "react-router-dom";

// A small menu bar that is folded until the button is pressed.
export default function Menu() {
  const [open, setOpen] = useState(false);
  const close = () => setOpen(false);
  return (
    <header className="menu">
      <button
        className="menu-toggle"
        aria-expanded={open}
        aria-controls="menu-links"
        aria-label={open ? "Close menu" : "Open menu"}
        onClick={() => setOpen(!open)}
      >
        ☰
      </button>
      {open && (
        <nav id="menu-links" className="menu-links">
          <Link to="/" onClick={close}>Home</Link>
          <Link to="/demo" onClick={close}>Demo</Link>
          <Link to="/login" onClick={close}>Log in</Link>
        </nav>
      )}
    </header>
  );
}
