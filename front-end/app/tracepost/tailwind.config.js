/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/**/*.{js,jsx,ts,tsx}", "./components/**/*.{js,jsx,ts,tsx}"],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: "#f97316", // orange-500
          light: "#fdba74", // orange-300
          dark: "#c2410c", // orange-700
        },
        secondary: {
          DEFAULT: "#4338ca", // indigo-700
          light: "#818cf8", // indigo-400
          dark: "#312e81", // indigo-900
        },
      },
    },
  },
  plugins: [],
};
