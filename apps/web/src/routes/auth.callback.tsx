/**
 * /auth/callback — Handles OAuth callback with cookie-based flow.
 *
 * With cookie-based auth, the OAuth flow is:
 *   1. User clicks "Sign in with Google" on the login page
 *   2. Browser is redirected to /v1/auth/google on the Go server
 *   3. Go server redirects to Google's consent screen
 *   4. Google redirects to /v1/auth/google/callback on the Go server
 *   5. Go server validates the code, creates/retrieves the operator, sets HttpOnly cookie
 *   6. Go server redirects browser to: FRONTEND_URL/dashboard?oauth=success&new=<bool>
 *
 * This route is kept for backwards compatibility but OAuth now redirects to dashboard
 * directly with the cookie already set.
 */

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, type ReactElement } from "react";
import { toast } from "sonner";

const OAuthCallbackPage = (): ReactElement => {
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const oauth = params.get("oauth");
    const isNew = params.get("new") === "true";

    // OAuth flow with cookies: Go server redirects to /dashboard?oauth=success&new=true
    // The cookie is already set, we just need to show a toast and redirect
    if (oauth === "success") {
      if (isNew) {
        toast.success("Welcome! Your operator account has been created.");
      } else {
        toast.success("Welcome back!");
      }
      navigate({ to: "/dashboard", replace: true });
      return;
    }

    // Fallback: if no oauth param, just redirect to dashboard
    // The useAuth hook will handle checking the session cookie
    navigate({ to: "/dashboard", replace: true });
  }, [navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="flex flex-col items-center gap-3">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        <p className="text-sm text-muted-foreground">Signing you in…</p>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/auth/callback")({
  head: () => ({ meta: [{ title: "Signing in — Vyzorix" }] }),
  component: OAuthCallbackPage,
});
