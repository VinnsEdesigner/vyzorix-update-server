/**
 * /auth/callback — Handles the JWT token returned by the Go server's Google OAuth callback.
 *
 * Flow:
 *   1. User clicks "Sign in with Google" on the login page
 *   2. Browser is redirected to /v1/auth/google on the Go server
 *   3. Go server redirects to Google's consent screen
 *   4. Google redirects to /v1/auth/google/callback on the Go server
 *   5. Go server validates the code, creates/retrieves the operator, issues a JWT
 *   6. Go server redirects browser to: FRONTEND_URL/auth/callback?token=<jwt>&isNew=<bool>
 *   7. This route parses the token, stores it in localStorage, redirects to /dashboard
 */

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useEffect, type ReactElement } from "react";
import { toast } from "sonner";

import { handleOAuthCallback } from "@/lib/vyzorix-auth";

export const Route = createFileRoute("/auth/callback")({
  ssr: false,
  head: () => ({ meta: [{ title: "Signing in — Vyzorix" }] }),
  component: OAuthCallbackPage,
});

const OAuthCallbackPage = (): ReactElement => {
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const token = params.get("token");
    const isNew = params.get("isNew") === "true";

    if (!token) {
      toast.error("OAuth callback missing token — sign-in failed");
      navigate({ to: "/login", replace: true });
      return;
    }

    const auth = handleOAuthCallback(token, isNew ? "true" : "false");
    if (!auth) {
      toast.error("Failed to process authentication token");
      navigate({ to: "/login", replace: true });
      return;
    }

    if (isNew) {
      toast.success("Welcome! Your operator account has been created.");
    }

    // Clean up the URL (remove token param) before navigating
    const cleanUrl = window.location.pathname;
    window.history.replaceState({}, "", cleanUrl);

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
