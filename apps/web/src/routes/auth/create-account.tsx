/**
 * create-account.tsx - Operator registration page.
 *
 * Uses the wolf background layout and SignUpForm component.
 */

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { type ReactNode, useState } from "react";
import { toast } from "sonner";

import wolfImage from "@/assets/images/black_wolf_evening_1781264516831.jpg";
import SignUpForm from "@/components/auth/SignUpForm";
import { registerOperator } from "@/lib/clients/authClient";
import { initiateSSO, type SSOProvider } from "@/lib/clients/ssoClient";

const CreateAccountPage = (): ReactNode => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSignUp = async (data: {
    fullName: string;
    email: string;
    username: string;
    password: string;
  }): Promise<void> => {
    setIsSubmitting(true);
    try {
      // Note: username is collected but Go backend uses email for identity
      // The password is required for registration
      await registerOperator({
        email: data.email,
        password: data.password,
        fullName: data.fullName,
      });
      toast.success("Account created. Check your email to verify.");
      navigate({
        to: "/auth/waitVerify",
        search: { email: data.email, token: "", type: "verify" },
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Registration failed";
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSSO = (provider: SSOProvider): void => {
    initiateSSO(provider);
  };

  const triggerToast = (msg: string, type?: "success" | "alert"): void => {
    if (type === "alert" || type === undefined) {
      toast.error(msg);
    } else {
      toast.success(msg);
    }
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
          <SignUpForm
            onSignUp={handleSignUp}
            onSSO={handleSSO}
            isSubmitting={isSubmitting}
            triggerToast={triggerToast}
          />
        </div>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/auth/create-account")({
  head: () => ({ meta: [{ title: "Create Account - Vyzorix" }] }),
  component: CreateAccountPage,
});
