"use client";

import { useTheme, ThemeProvider } from "@/hooks/useTheme";

function ThemeToggleInner() {
  const { theme, setTheme } = useTheme();

  const cycle = () => {
    const order: Array<"light" | "dark" | "system"> = ["system", "light", "dark"];
    const idx = order.indexOf(theme);
    setTheme(order[(idx + 1) % order.length]);
  };

  const icons: Record<string, string> = {
    light: "☀️",
    dark: "🌙",
    system: "💻",
  };

  const labels: Record<string, string> = {
    light: "Light",
    dark: "Dark",
    system: "System",
  };

  return (
    <button
      onClick={cycle}
      className="flex items-center gap-1.5 text-xs text-muted hover:text-foreground transition-colors px-2 py-1 rounded-lg hover:bg-card-hover"
      title={`Theme: ${labels[theme]}`}
    >
      <span className="text-base">{icons[theme]}</span>
      <span className="hidden sm:inline">{labels[theme]}</span>
    </button>
  );
}

function ThemeToggle() {
  return (
    <ThemeProvider>
      <ThemeToggleInner />
    </ThemeProvider>
  );
}

export { ThemeToggle, ThemeProvider };
