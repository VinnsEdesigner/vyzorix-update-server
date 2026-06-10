// @vyzorix/config/tailwind/index.js - Tailwind CSS Theme Preset
// Vyzorix brand colors and design system with Rose-500 as primary

const defaultConfig = {
  // Vyzorix Brand Colors - Rose-500 Primary
  colors: {
    // Primary - Rose-500 (MOST IMPORTANT BRAND COLOR)
    primary: {
      DEFAULT: "oklch(0.645 0.246 16.439)",
      foreground: "oklch(0.985 0 0)",
      50: "oklch(0.95 0.03 16.439)",
      100: "oklch(0.90 0.06 16.439)",
      200: "oklch(0.85 0.09 16.439)",
      300: "oklch(0.78 0.14 16.439)",
      400: "oklch(0.71 0.18 16.439)",
      500: "oklch(0.645 0.246 16.439)",
      600: "oklch(0.60 0.20 16.439)",
      700: "oklch(0.55 0.18 16.439)",
      800: "oklch(0.50 0.16 16.439)",
      900: "oklch(0.45 0.14 16.439)",
      950: "oklch(0.40 0.12 16.439)",
    },
    
    // Secondary - Cool blue-gray
    secondary: {
      DEFAULT: "oklch(0.968 0.007 247.896)",
      foreground: "oklch(0.208 0.042 265.755)",
      50: "oklch(0.98 0.005 247.896)",
      100: "oklch(0.96 0.008 247.896)",
      200: "oklch(0.94 0.010 247.896)",
      300: "oklch(0.92 0.012 247.896)",
      400: "oklch(0.90 0.014 247.896)",
      500: "oklch(0.968 0.007 247.896)",
      600: "oklch(0.85 0.010 247.896)",
      700: "oklch(0.75 0.008 247.896)",
      800: "oklch(0.65 0.006 247.896)",
      900: "oklch(0.55 0.004 247.896)",
    },
    
    // Accent - Teal/Cyan
    accent: {
      DEFAULT: "oklch(0.704 0.191 22.216)",
      foreground: "oklch(0.984 0.003 247.858)",
      50: "oklch(0.85 0.15 22.216)",
      100: "oklch(0.80 0.17 22.216)",
      200: "oklch(0.75 0.18 22.216)",
      300: "oklch(0.73 0.19 22.216)",
      400: "oklch(0.71 0.195 22.216)",
      500: "oklch(0.704 0.191 22.216)",
      600: "oklch(0.65 0.17 22.216)",
      700: "oklch(0.55 0.14 22.216)",
      800: "oklch(0.45 0.11 22.216)",
      900: "oklch(0.35 0.08 22.216)",
    },
    
    // Destructive - Red
    destructive: {
      DEFAULT: "oklch(0.577 0.245 27.325)",
      foreground: "oklch(0.984 0.003 247.858)",
    },
    
    // Success - Green
    success: {
      DEFAULT: "oklch(0.627 0.265 303.9)",
      foreground: "oklch(0.985 0 0)",
    },
    
    // Warning - Amber
    warning: {
      DEFAULT: "oklch(0.769 0.188 70.08)",
      foreground: "oklch(0.208 0.042 265.755)",
    },
    
    // Chart colors (Rose-based)
    chart: {
      1: "oklch(0.645 0.246 16.439)",   // Rose-500
      2: "oklch(0.6 0.118 184.704)",
      3: "oklch(0.398 0.07 227.392)",
      4: "oklch(0.828 0.189 84.429)",
      5: "oklch(0.769 0.188 70.08)",
    },
    
    // Sidebar colors
    sidebar: {
      DEFAULT: "oklch(0.984 0.003 247.858)",
      foreground: "oklch(0.129 0.042 264.695)",
      primary: "oklch(0.645 0.246 16.439)",
      "primary-foreground": "oklch(0.985 0 0)",
      accent: "oklch(0.968 0.007 247.896)",
      "accent-foreground": "oklch(0.208 0.042 265.755)",
      border: "oklch(0.929 0.013 255.508)",
      ring: "oklch(0.551 0.027 264.364)",
    },
    
    // Light mode (default)
    background: "oklch(1 0 0)",
    foreground: "oklch(0.129 0.042 264.695)",
    card: "oklch(1 0 0)",
    "card-foreground": "oklch(0.129 0.042 264.695)",
    popover: "oklch(1 0 0)",
    "popover-foreground": "oklch(0.129 0.042 264.695)",
    muted: "oklch(0.968 0.007 247.896)",
    "muted-foreground": "oklch(0.554 0.046 257.417)",
    border: "oklch(0.929 0.013 255.508)",
    input: "oklch(0.929 0.013 255.508)",
    ring: "oklch(0.645 0.246 16.439)",
  },
  
  // Border radius
  borderRadius: {
    none: "0px",
    sm: "calc(var(--radius) - 4px)",
    DEFAULT: "calc(var(--radius) - 2px)",
    md: "var(--radius)",
    lg: "calc(var(--radius) + 2px)",
    xl: "calc(var(--radius) + 4px)",
    "2xl": "calc(var(--radius) + 8px)",
    "3xl": "calc(var(--radius) + 12px)",
    "4xl": "calc(var(--radius) + 16px)",
    full: "9999px",
  },
  
  // Animations
  animate: {
    spin: "spin 1s linear infinite",
    ping: "ping 1s cubic-bezier(0, 0, 0.2, 1) infinite",
    pulse: "pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite",
    bounce: "bounce 1s infinite",
    "fade-in": "fadeIn 0.2s ease-in",
    "fade-out": "fadeOut 0.2s ease-out",
    "slide-in-from-top": "slideInFromTop 0.3s ease-out",
    "slide-in-from-bottom": "slideInFromBottom 0.3s ease-out",
    "slide-in-from-left": "slideInFromLeft 0.3s ease-out",
    "slide-in-from-right": "slideInFromRight 0.3s ease-out",
  },
  
  // Keyframes
  keyframes: {
    fadeIn: {
      from: { opacity: "0" },
      to: { opacity: "1" },
    },
    fadeOut: {
      from: { opacity: "1" },
      to: { opacity: "0" },
    },
    slideInFromTop: {
      from: { transform: "translateY(-100%)" },
      to: { transform: "translateY(0)" },
    },
    slideInFromBottom: {
      from: { transform: "translateY(100%)" },
      to: { transform: "translateY(0)" },
    },
    slideInFromLeft: {
      from: { transform: "translateX(-100%)" },
      to: { transform: "translateX(0)" },
    },
    slideInFromRight: {
      from: { transform: "translateX(100%)" },
      to: { transform: "translateX(0)" },
    },
  },
};

// Dark mode overrides
const darkMode = {
  colors: {
    background: "oklch(0.129 0.042 264.695)",
    foreground: "oklch(0.984 0.003 247.858)",
    card: "oklch(0.208 0.042 265.755)",
    "card-foreground": "oklch(0.984 0.003 247.858)",
    popover: "oklch(0.208 0.042 265.755)",
    "popover-foreground": "oklch(0.984 0.003 247.858)",
    primary: "oklch(0.645 0.246 16.439)",
    "primary-foreground": "oklch(0.985 0 0)",
    secondary: "oklch(0.279 0.041 260.031)",
    "secondary-foreground": "oklch(0.984 0.003 247.858)",
    muted: "oklch(0.279 0.041 260.031)",
    "muted-foreground": "oklch(0.704 0.04 256.788)",
    accent: "oklch(0.279 0.041 260.031)",
    "accent-foreground": "oklch(0.984 0.003 247.858)",
    destructive: "oklch(0.704 0.191 22.216)",
    "destructive-foreground": "oklch(0.984 0.003 247.858)",
    border: "oklch(1 0 0 / 10%)",
    input: "oklch(1 0 0 / 15%)",
    ring: "oklch(0.645 0.246 16.439)",
    chart: {
      1: "oklch(0.488 0.243 264.376)",
      2: "oklch(0.696 0.17 162.48)",
      3: "oklch(0.769 0.188 70.08)",
      4: "oklch(0.627 0.265 303.9)",
      5: "oklch(0.645 0.246 16.439)",
    },
    sidebar: {
      DEFAULT: "oklch(0.208 0.042 265.755)",
      foreground: "oklch(0.984 0.003 247.858)",
      primary: "oklch(0.645 0.246 16.439)",
      "primary-foreground": "oklch(0.985 0 0)",
      accent: "oklch(0.279 0.041 260.031)",
      "accent-foreground": "oklch(0.984 0.003 247.858)",
      border: "oklch(1 0 0 / 10%)",
      ring: "oklch(0.551 0.027 264.364)",
    },
  },
};

// Export Rose-500 as default (most important brand color)
export default { ...defaultConfig, darkMode };

// Export themes
export { rose500Theme } from "./themes/rose-500";
export { defaultTheme } from "./themes/default";