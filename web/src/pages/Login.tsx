// Intentionally non-functional for now: nothing is submitted, stored or checked.
export default function Login() {
  return (
    <section className="center">
      <h1>Log in</h1>
      <form className="form" onSubmit={(event) => event.preventDefault()}>
        <label>
          Email
          <input type="email" disabled />
        </label>
        <label>
          Password
          <input type="password" disabled />
        </label>
        <button type="submit" disabled>Log in</button>
      </form>
      <p className="notice">Not available yet.</p>
    </section>
  );
}
