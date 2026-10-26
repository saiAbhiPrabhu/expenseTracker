import { useEffect } from "react";

// Paints the page cream while a ledger page is shown.
export function useLedgerBody() {
  useEffect(() => {
    document.body.classList.add("ledger");
    return () => document.body.classList.remove("ledger");
  }, []);
}
