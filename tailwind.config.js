/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--bg-primary)",
        surface: "var(--bg-card)",
        primary: "var(--accent-red)",
        secondary: "var(--text-secondary)",
        border: "var(--bg-tertiary)",
      }
    },
  },
  plugins: [],
}
