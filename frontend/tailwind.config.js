// tailwind.config.js
module.exports = {
  darkMode: 'class', // 通过 class 切换暗模式
  content: [
    "./app.vue",
    "./components/**/*.{vue,js,ts}",
    "./layouts/**/*.vue",
    "./pages/**/*.vue",
    "./plugins/**/*.{js,ts}",
    "./nuxt.config.{js,ts}"
  ],
  theme: {
    extend: {
      fontFamily: {
        hack: ['"Hack Nerd Mono"', 'monospace'],
        hy12bit: ['"HY 12bit"', 'monospace'],
      },
      colors: {
        editorBackground: '#f8f9fa',
      },
    },
  },
  plugins: [],
};
