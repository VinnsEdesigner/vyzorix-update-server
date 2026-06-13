import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Shield, Loader2 } from "lucide-react";
import { useEffect, useState, type ReactElement } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Toaster } from "@/components/ui/sonner";
import { logger } from "@/lib/logger";
import { getFullHydratedState } from "@/lib/server/state-injector";
import { login, register, redirectToGoogleOAuth } from "@/lib/vyzorix-auth";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

type Mode = "signin" | "signup";

const LoginPage = (): ReactElement => {
  const navigate = useNavigate();
  const { serverUrl } = useVyzorixConfig();
  const [mode, setMode] = useState<Mode>("signin");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [pw, setPw] = useState("");
  const [loading, setLoading] = useState(false);
  const [oauthLoading, setOauthLoading] = useState(false);

  /**
   * SSR-aware authentication check
   *
   * This runs on mount and checks:
   * 1. Server-injected state (SSR hydration) - fastest, no flash
   * 2. API fallback - for client-side navigation without SSR state
   *
   * Matches Library's pattern for auth state checking
   */
  useEffect(() => {
    const checkAuth = async (): Promise<void> => {
      // 1. Check server-injected state first (SSR hydration)
      if (typeof window !== "undefined") {
        const globalState = getFullHydratedState();
        if (globalState?.isAuthenticated) {
          // Server already validated - redirect immediately
          navigate({ to: "/dashboard", replace: true });
          return;
        }
      }

      // 2. Fallback: Fetch from API (client-side navigation without SSR state)
      try {
        const res = await fetch("/v1/auth/me", { credentials: "include" });
        if (res.ok) {
          navigate({ to: "/dashboard", replace: true });
        }
      } catch {
        // Not authenticated - stay on login page
      }
    };

    checkAuth();
  }, [navigate]);

  const submit = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();
    setLoading(true);
    try {
      if (mode === "signin") {
        await login(serverUrl, email.trim(), pw);
        logger.info("auth", "Password sign-in OK", { email: email.trim() });
      } else {
        await register(serverUrl, email.trim(), pw, name.trim());
        logger.info("auth", "Operator registration OK", { email: email.trim(), name: name.trim() });
        toast.success("Operator account created.");
      }
      navigate({ to: "/dashboard", replace: true });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Authentication failed";
      toast.error(msg);
      logger.warn("auth", `Auth failed: ${msg}`);
    } finally {
      setLoading(false);
    }
  };

  const google = (): void => {
    setOauthLoading(true);
    try {
      // The Go server redirects to Google's OAuth consent screen.
      // After approval, Google redirects back to the Go server's callback,
      // which then redirects to the frontend /auth/callback page with the JWT.
      redirectToGoogleOAuth(serverUrl, "/dashboard");
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Google sign-in failed";
      toast.error(msg);
      setOauthLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Shield className="h-5 w-5" />
          </div>
          <CardTitle>Vyzorix Console</CardTitle>
          <CardDescription>
            {mode === "signin"
              ? "Sign in to the update server control plane"
              : "Create the first operator account (becomes super_admin)"}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <Button
            type="button"
            variant="outline"
            className="w-full"
            onClick={google}
            disabled={oauthLoading || loading}
          >
            {oauthLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            Continue with Google
          </Button>
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Separator className="flex-1" />
            or
            <Separator className="flex-1" />
          </div>
          <form onSubmit={submit} className="space-y-3">
            {mode === "signup" && (
              <div className="space-y-1.5">
                <Label htmlFor="name">Full name</Label>
                <Input
                  id="name"
                  type="text"
                  autoComplete="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Vinns Designer"
                  required
                />
              </div>
            )}
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="pw">Password</Label>
              <Input
                id="pw"
                type="password"
                autoComplete={mode === "signin" ? "current-password" : "new-password"}
                minLength={8}
                value={pw}
                onChange={(e) => setPw(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading || oauthLoading}>
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              {mode === "signin" ? "Sign in" : "Create account"}
            </Button>
          </form>
          <p className="text-center text-xs text-muted-foreground">
            {mode === "signin" ? (
              <>
                First time?{" "}
                <button
                  type="button"
                  className="font-medium underline"
                  onClick={() => setMode("signup")}
                >
                  Create the operator account
                </button>
              </>
            ) : (
              <button
                type="button"
                className="font-medium underline"
                onClick={() => setMode("signin")}
              >
                Back to sign in
              </button>
            )}
          </p>
          {mode === "signin" && (
            <p className="text-center text-xs text-muted-foreground">
              Forgot your password?{" "}
              <button
                type="button"
                className="font-medium underline"
                onClick={() => navigate({ to: "/forgot-password" })}
              >
                Reset it
              </button>
            </p>
          )}
          <p className="text-center text-[10px] text-muted-foreground">
            Access is restricted to allowlisted operators only.
          </p>
        </CardContent>
      </Card>
      <Toaster />
    </div>
  );
};

export const Route = createFileRoute("/login")({
  head: () => ({ meta: [{ title: "Sign in — Vyzorix" }] }),
  component: LoginPage,
});
