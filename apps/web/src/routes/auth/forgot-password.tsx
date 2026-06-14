/**
 * forgot-password.tsx - Password reset request page.
 *
 * Uses the wolf background layout and ForgotPasswordForm component.
 * After successful request, redirects to waitVerify page with type=reset.
 */

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { type ReactNode, useState } from "react";
import { toast } from "sonner";

import wolfImage from "@/assets/images/black_wolf_evening_1781264516831.jpg";
import ForgotPasswordForm from "@/components/auth/ForgotPasswordForm";
import { requestPasswordReset } from "@/lib/clients/passwordClient";

const ForgotPasswordPage = (): ReactNode => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleResetSubmit = async (email: string): Promise<void> => {
    setIsSubmitting(true);
    try {
      await requestPasswordReset(email);
      toast.success("Password reset instructions sent to your email.");
      // Redirect to waitVerify page with email, empty token, and type=reset
      navigate({ to: "/auth/waitVerify", search: { email, token: "", type: "reset" } });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to process request";
      toast.error(msg);
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
          <ForgotPasswordForm
            onResetSubmit={handleResetSubmit}
            onBackToLogin={handleBackToLogin}
            isSubmitting={isSubmitting}
          />
        </div>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/auth/forgot-password")({
  head: () => ({ meta: [{ title: "Forgot Password - Vyzorix" }] }),
  component: ForgotPasswordPage,
});
