import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// In dev, the Go server (go run . -serve :8080) answers /api; Vite serves the pages.
export default defineConfig({
  plugins: [react()],
  server: { proxy: { "/api": "http://localhost:8080" } },
});
