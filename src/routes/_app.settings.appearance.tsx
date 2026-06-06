import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Monitor, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";

export const Route = createFileRoute("/_app/settings/appearance")({
  component: AppearanceSettings,
});

type Theme = "system" | "light" | "dark";
const KEY = "vyzorix.theme";

function apply(theme: Theme) {
  const root = document.documentElement;
  const sysDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  const dark = theme === "dark" || (theme === "system" && sysDark);
  root.classList.toggle("dark", dark);
}

function AppearanceSettings() {
  const [theme, setTheme] = useState<Theme>(() => {
    try {
      return (localStorage.getItem(KEY) as Theme) || "system";
    } catch {
      return "system";
    }
  });

  useEffect(() => {
    try {
      localStorage.setItem(KEY, theme);
    } catch {
      // ignore storage error
    }
    apply(theme);
  }, [theme]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Appearance</CardTitle>
        <CardDescription>Theme and density preferences for this browser.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label className="text-sm">Theme</Label>
          <div className="grid grid-cols-3 gap-2">
            <ThemeBtn
              current={theme}
              value="system"
              label="System"
              icon={Monitor}
              onClick={setTheme}
            />
            <ThemeBtn current={theme} value="light" label="Light" icon={Sun} onClick={setTheme} />
            <ThemeBtn current={theme} value="dark" label="Dark" icon={Moon} onClick={setTheme} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function ThemeBtn({
  current,
  value,
  label,
  icon: Icon,
  onClick,
}: {
  current: Theme;
  value: Theme;
  label: string;
  icon: typeof Monitor;
  onClick: (v: Theme) => void;
}) {
  return (
    <Button
      variant={current === value ? "default" : "outline"}
      className="h-20 flex-col gap-2"
      onClick={() => onClick(value)}
    >
      <Icon className="h-5 w-5" />
      <span className="text-xs">{label}</span>
    </Button>
  );
}
