/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#2563eb',
        surface: '#faf8ff',
        'surface-container': '#ededf9',
        'on-surface': '#191b23',
        'on-surface-variant': '#434655',
        outline: '#c3c6d7',
      },
    },
  },
  plugins: [],
}
