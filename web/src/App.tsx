import { Route, Routes } from "react-router-dom";
import Menu from "./components/Menu";
import Demo from "./pages/Demo";
import Home from "./pages/Home";
import Login from "./pages/Login";

export default function App() {
  return (
    <>
      <Menu />
      <div className="page">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/demo" element={<Demo />} />
          <Route path="*" element={<main className="surface narrow"><p className="notice">Page not found.</p></main>} />
        </Routes>
      </div>
    </>
  );
}
