import { KeyboardEvent, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import "../ledger/ledger.css";

const GITHUB_URL = "https://github.com/saiAbhiPrabhu/expenseTracker";

// A small closed book that opens on click, sliding a bar of links out beside it.
export default function Menu() {
  const [open, setOpen] = useState(false);
  const toggle = useRef<HTMLButtonElement>(null);
  const onHome = useLocation().pathname === "/";
  const close = () => setOpen(false);
  const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Escape" && open) {
      close();
      toggle.current?.focus();
    }
  };
  return (
    <header className={open ? "book-menu open" : "book-menu"} onKeyDown={onKeyDown}>
      <div className="book-row">
        <button ref={toggle} className="book-toggle" aria-expanded={open} aria-controls="menu-links" aria-label="Menu" onClick={() => setOpen(!open)}>
          <span className="book" aria-hidden="true">
            <span className="cover"><span className="face front" /><span className="face back" /></span>
            <span className="leaf l1" />
            <span className="leaf l2" />
            <span className="leaf l3" />
          </span>
        </button>
        <nav id="menu-links" className="book-bar" aria-label="Menu">
          {onHome ? (
            <>
              {/* No ABOUT or CONTACT pages exist yet. */}
              <span className="unavailable">About</span>
              <a href={GITHUB_URL} target="_blank" rel="noopener noreferrer" onClick={close}>GitHub</a>
              <span className="unavailable">Contact</span>
            </>
          ) : (
            <>
              <Link to="/" onClick={close}>Home</Link>
              <Link to="/demo" onClick={close}>Demo</Link>
              <Link to="/login" onClick={close}>Log in</Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
