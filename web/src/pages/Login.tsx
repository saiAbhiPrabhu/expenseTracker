// Intentionally non-functional for now: nothing is submitted, stored or checked.
export default function Login() {
  return (
    <main className="surface narrow">
      <h1>Log in</h1>
      <p className="lead" id="login-off">Log-in is not available in the demo.</p>
      <form className="form" onSubmit={(event) => event.preventDefault()}>
        <label>
          Email
          <input type="email" disabled aria-describedby="login-off" />
        </label>
        <label>
          Password
          <input type="password" disabled aria-describedby="login-off" />
        </label>
        <button type="submit" className="primary" disabled aria-describedby="login-off">Log in</button>
      </form>
    </main>
  );
}
