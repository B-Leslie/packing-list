/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}", // Scans all JS, JSX, TS, TSX files in src
    "./public/index.html"         // Include public/index.html if you use classes there
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'), // If you are using form styling plugin
  ],
}

