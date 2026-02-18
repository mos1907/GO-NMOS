/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{svelte,js,ts}"],
  theme: {
    extend: {
      colors: {
        "nmos-bg": "#343434",
        "nmos-card": "#343434",
        svelte: "#ff3e00",
        "svelte-soft": "#ffb199"
      },
    },
  },
  plugins: [],
};

