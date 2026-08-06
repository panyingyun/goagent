/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        background: '#faf8ff',
        surface: '#faf8ff',
        'surface-dim': '#d9d9e5',
        'surface-bright': '#faf8ff',
        'surface-container-lowest': '#ffffff',
        'surface-container-low': '#f3f3fe',
        'surface-container': '#ededf9',
        'surface-container-high': '#e7e7f3',
        'surface-container-highest': '#e1e2ed',
        'on-surface': '#191b23',
        'on-surface-variant': '#434655',
        outline: '#737686',
        'outline-variant': '#c3c6d7',
        primary: '#004ac6',
        'primary-container': '#2563eb',
        'on-primary': '#ffffff',
        'on-primary-container': '#eeefff',
        secondary: '#b61722',
        'secondary-container': '#da3437',
        error: '#ba1a1a',
        'user-bubble': '#fff1e8',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'PingFang SC', 'Microsoft YaHei', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
      },
      spacing: {
        sidebar: '260px',
      },
      borderRadius: {
        xl: '0.75rem',
      },
    },
  },
  plugins: [],
}
