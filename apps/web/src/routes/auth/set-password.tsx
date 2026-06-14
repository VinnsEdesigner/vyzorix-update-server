/**
 * set-password.tsx - Set new password after email verification.
 *
 * User clicks the reset link in their email which goes to /auth/set-password?token=xxx
 * This page allows them to set a new password.
 */

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { type ReactNode, useState } from "react";
import { toast } from "sonner";

import wolfImage from "@/assets/images/black_wolf_evening_1781264516831.jpg";
import SetPasswordForm from "@/components/auth/SetPasswordForm";

const SetPasswordPage = (): ReactNode => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { token = "" } = Route.useSearch();

  const handleSubmit = async (password: string): Promise<void> => {
    setIsSubmitting(true);
    try {
      if (!token) {
        toast.error("Invalid reset link. Please request a new one.");
        navigate({ to: "/auth/forgot-password" });
        return;
      }

      // Call the API to set the new password
      const response = await fetch("/v1/auth/reset-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ token, newPassword: password }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message ?? "Failed to set password");
      }

      toast.success("Password set successfully!");
      navigate({ to: "/auth/login" });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to set password";
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleBackToLogin = (): void => {
    navigate({ to: "/auth/login" });
  };

  return (
    <div className="relative min-h-screen w-full overflow-hidden">
      {/* Wolf background image - fixed, covers viewport */}
      <div
        className="fixed inset-0 bg-cover bg-center bg-no-repeat"
        style={{ backgroundImage: `url(${wolfImage})` }}
        aria-hidden="true"
      />

      {/* Dark gradient overlay for readability */}
      <div
        className="fixed inset-0 bg-gradient-to-br from-slate-950/80 via-slate-950/70 to-slate-950/85"
        aria-hidden="true"
      />

      {/* Content container - centered */}
      <div className="relative z-10 flex min-h-screen items-center justify-center px-4 py-8">
        <div className="w-full max-w-md">
          <SetPasswordForm
            onSubmit={handleSubmit}
            onBackToLogin={handleBackToLogin}
            isSubmitting={isSubmitting}
          />
        </div>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/auth/set-password")({
  head: () => ({ meta: [{ title: "Set Password - Vyzorix" }] }),
  validateSearch: (search: Record<string, unknown>) => {
    return {
      token: (search.token as string) ?? "",
    };
  },
  component: SetPasswordPage,
});
