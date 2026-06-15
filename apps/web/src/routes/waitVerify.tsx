/**
 * waitVerify.tsx - Email verification and password reset waiting page.
 *
 * Two flows with different behaviors:
 *
 * REGISTRATION FLOW (type=verify):
 *   - After sign-up, user lands here with email param (no token yet)
 *   - Shows "check your email" UI with polling
 *   - When email link is clicked: /auth/waitVerify?token=xxx&type=verify
 *   - Verifies token → success → dashboard
 *
 * PASSWORD RESET FLOW (type=reset):
 *   - After forgot-password submit, user lands here with email param
 *   - Shows "Reset link sent! Check your email..." with resend countdown
 *   - "Didn't receive link?" option to go back to forgot-password
 *   - Resend button with progressive rate limiting (30s, 60s, 90s...)
 *   - When email link is clicked: /auth/waitVerify?token=xxx&type=reset
 *   - Verifies token → success → redirect to /auth/set-password?token=xxx
 */

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { CheckCircle, Loader2, Mail, ArrowLeft } from "lucide-react";
import { useEffect, useState, type ReactElement } from "react";
import { toast } from "sonner";

import wolfImage from "@/assets/images/black_wolf_evening_1781264516831.jpg";
import AuthLayout from "@/components/layout/AuthLayout";
import { resendPasswordReset } from "@/lib/clients/passwordClient";
import { triggerTokenResend } from "@/lib/clients/verificationClient";

const MAX_WAIT_SECONDS = 900; // 15 minutes

type FlowType = "verify" | "reset";

interface ResendButtonProps {
  isResending: boolean;
  resendCooldown: number;
  onResend: () => void;
  formatTime: (seconds: number) => string;
}

function ResendButton({
  isResending,
  resendCooldown,
  onResend,
  formatTime,
}: ResendButtonProps): ReactElement {
  const isDisabled = resendCooldown > 0 || isResending;

  const getButtonText = (): string => {
    if (isResending) {
      return "Sending...";
    }
    if (resendCooldown > 0) {
      return `Resend in ${formatTime(resendCooldown)}`;
    }
    return "Resend reset link";
  };

  return (
    <button
      type="button"
      className="w-full bg-rose-600 hover:bg-rose-500 disabled:bg-slate-700 disabled:text-slate-500 text-white font-semibold py-3 rounded-xl transition-all text-sm flex items-center justify-center gap-2 disabled:cursor-not-allowed"
      onClick={onResend}
      disabled={isDisabled}
    >
      {isResending && <Loader2 className="w-4 h-4 animate-spin" />}
      <span>{getButtonText()}</span>
    </button>
  );
}

const WaitVerifyPage = (): ReactElement => {
  const navigate = useNavigate();
  const search = Route.useSearch();
  const { email = "", token = "", type = "verify" } = search;

  const [timeLeft, setTimeLeft] = useState(MAX_WAIT_SECONDS);
  const [isVerifying, setIsVerifying] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);
  const [isResending, setIsResending] = useState(false);

  const flowType = (type as FlowType) || "verify";
  const hasToken = Boolean(token);
  const isResetFlow = flowType === "reset";

  // Countdown timer for waiting state (only when no token and not reset flow)
  useEffect(() => {
    if (hasToken || showSuccess || isResetFlow) {
      return; // Don't count down when token present, showing success, or reset flow
    }

    const interval = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [hasToken, showSuccess, isResetFlow]);

  // Resend cooldown countdown
  useEffect(() => {
    if (resendCooldown <= 0) {
      return;
    }

    const interval = setInterval(() => {
      setResendCooldown((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [resendCooldown]);

  // Handle token verification when present in URL
  useEffect(() => {
    if (!hasToken) {
      return;
    }

    setIsVerifying(true);

    const verifyAndRedirect = async (): Promise<void> => {
      try {
        const response = await fetch("/v1/auth/verify-email", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ token }),
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(errorData.message ?? "Verification failed");
        }

        setShowSuccess(true);
        setIsVerifying(false);

        // Short delay before redirect for user to see success state
        setTimeout(() => {
          if (flowType === "reset") {
            // Password reset: redirect to set-password page with token
            navigate({ to: "/set-password", search: { token }, replace: true });
          } else {
            // Registration: redirect to dashboard
            navigate({ to: "/dashboard", replace: true });
          }
        }, 1500);
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Verification failed";
        toast.error(msg);
        setIsVerifying(false);
      }
    };

    verifyAndRedirect();
  }, [token, flowType, hasToken, navigate]);

  const handleResend = async (): Promise<void> => {
    if (resendCooldown > 0 || isResending) {
      return;
    }

    setIsResending(true);
    try {
      if (isResetFlow) {
        // Password reset flow - use resendPasswordReset
        const result = await resendPasswordReset(email);
        if (result.success) {
          toast.success("Reset link resent!");
          setResendCooldown(30); // Minimum cooldown after first resend
        }
      } else {
        // Email verification flow - use triggerTokenResend
        const result = await triggerTokenResend(email);
        if (result.success) {
          toast.success("Verification email resent!");
          setResendCooldown(30); // Minimum cooldown after first resend
        }
      }
    } catch (err) {
      const error = err as Error & { retry_after?: number; locked_until?: number };
      if (error.retry_after) {
        setResendCooldown(error.retry_after);
        toast.error(`Please wait ${error.retry_after}s before resending.`);
      } else if (error.locked_until) {
        const lockSeconds = Math.ceil((error.locked_until - Date.now()) / 1000);
        setResendCooldown(lockSeconds);
        toast.error(
          `Too many attempts. Please try again in ${Math.ceil(lockSeconds / 60)} minutes.`,
        );
      } else {
        toast.error(err instanceof Error ? err.message : "Resend failed");
      }
    } finally {
      setIsResending(false);
    }
  };

  const handleBackToLogin = (): void => {
    navigate({ to: "/login" });
  };

  const handleBackToForgotPassword = (): void => {
    navigate({ to: "/forgot-password" });
  };

  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  // Show success state
  if (showSuccess) {
    return (
      <div className="relative min-h-screen w-full overflow-hidden">
        <div
          className="fixed inset-0 bg-cover bg-center bg-no-repeat"
          style={{ backgroundImage: `url(${wolfImage})` }}
          aria-hidden="true"
        />
        <div
          className="fixed inset-0 bg-gradient-to-br from-slate-950/80 via-slate-950/70 to-slate-950/85"
          aria-hidden="true"
        />
        <div className="relative z-10 flex min-h-screen items-center justify-center px-4 py-8">
          <div className="w-full max-w-md">
            <div className="bg-slate-900/80 backdrop-blur-xl border border-white/10 rounded-2xl p-8 text-center shadow-2xl">
              <div className="flex justify-center mb-4">
                <div className="relative w-16 h-16">
                  <style>{`
                    @keyframes success-pulse {
                      0%, 100% { transform: scale(1); opacity: 1; }
                      50% { transform: scale(1.1); opacity: 0.8; }
                    }
                    .success-pulse {
                      animation: success-pulse 1.5s ease-in-out infinite;
                    }
                  `}</style>
                  <CheckCircle className="w-16 h-16 text-emerald-500 success-pulse" />
                </div>
              </div>
              <h2 className="text-2xl font-semibold text-white mb-2">
                {flowType === "reset" ? "Reset Link Verified!" : "Email Verified!"}
              </h2>
              <p className="text-slate-400 text-sm">
                {flowType === "reset"
                  ? "Redirecting to set your new password..."
                  : "Redirecting to your dashboard..."}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Show verification loading state when token is present
  if (isVerifying) {
    return (
      <div className="relative min-h-screen w-full overflow-hidden">
        <div
          className="fixed inset-0 bg-cover bg-center bg-no-repeat"
          style={{ backgroundImage: `url(${wolfImage})` }}
          aria-hidden="true"
        />
        <div
          className="fixed inset-0 bg-gradient-to-br from-slate-950/80 via-slate-950/70 to-slate-950/85"
          aria-hidden="true"
        />
        <div className="relative z-10 flex min-h-screen items-center justify-center px-4 py-8">
          <div className="w-full max-w-md">
            <div className="bg-slate-900/80 backdrop-blur-xl border border-white/10 rounded-2xl p-8 text-center shadow-2xl">
              <div className="flex justify-center mb-4">
                <div className="relative w-16 h-16">
                  <style>{`
                    @keyframes verify-spin {
                      0% { transform: rotate(0deg); }
                      100% { transform: rotate(360deg); }
                    }
                    .verify-spin {
                      animation: verify-spin 1s linear infinite;
                    }
                  `}</style>
                  <Loader2 className="w-16 h-16 text-rose-500 verify-spin" />
                </div>
              </div>
              <h2 className="text-2xl font-semibold text-white mb-2">Verifying...</h2>
              <p className="text-slate-400 text-sm">
                {flowType === "reset"
                  ? "Confirming your password reset link..."
                  : "Confirming your email address..."}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // For reset flow: show specialized UI with resend option
  if (isResetFlow) {
    return (
      <div className="relative min-h-screen w-full overflow-hidden">
        <div
          className="fixed inset-0 bg-cover bg-center bg-no-repeat"
          style={{ backgroundImage: `url(${wolfImage})` }}
          aria-hidden="true"
        />
        <div
          className="fixed inset-0 bg-gradient-to-br from-slate-950/80 via-slate-950/70 to-slate-950/85"
          aria-hidden="true"
        />
        <div className="relative z-10 flex min-h-screen items-center justify-center px-4 py-8">
          <div className="w-full max-w-md">
            <div className="bg-slate-900/80 backdrop-blur-xl border border-white/10 rounded-2xl p-8 shadow-2xl">
              {/* Header */}
              <div className="text-center mb-6">
                <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-rose-500/10">
                  <Mail className="h-7 w-7 text-rose-500" />
                </div>
                <h2 className="text-2xl font-semibold text-white mb-2">Check your email</h2>
                <p className="text-slate-400 text-sm">We sent a password reset link to:</p>
                <p className="text-rose-400 font-medium text-sm mt-1 break-all">{email}</p>
              </div>

              {/* Status indicator */}
              <div className="bg-slate-800/50 border border-white/5 rounded-xl p-4 mb-6">
                <p className="text-emerald-400 text-sm font-medium text-center">
                  ✓ Reset link sent successfully
                </p>
                <p className="text-slate-400 text-xs text-center mt-2">
                  Click the link in your email to reset your password.
                </p>
              </div>

              {/* Actions */}
              <div className="space-y-3">
                {/* Resend button with cooldown */}
                <ResendButton
                  isResending={isResending}
                  resendCooldown={resendCooldown}
                  onResend={handleResend}
                  formatTime={formatTime}
                />

                {/* Didn't receive link? */}
                <div className="text-center space-y-2">
                  <p className="text-slate-400 text-xs">Did not receive the link?</p>
                  <button
                    type="button"
                    className="text-rose-400 hover:text-rose-300 text-xs font-medium underline underline-offset-4 transition-colors"
                    onClick={handleBackToForgotPassword}
                  >
                    Try different email address
                  </button>
                </div>

                {/* Back to login */}
                <button
                  type="button"
                  className="w-full text-slate-400 hover:text-white text-xs transition-colors flex items-center justify-center gap-1 mx-auto"
                  onClick={handleBackToLogin}
                >
                  <ArrowLeft className="w-3 h-3" />
                  <span>Back to login</span>
                </button>
              </div>

              {/* Info note */}
              <div className="mt-6 pt-4 border-t border-white/5">
                <p className="text-slate-500 text-xs text-center">
                  Reset links expire after 15 minutes. If you do not see the email, check your spam
                  folder.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // For verify flow: show the existing waiting UI
  const getResendButtonText = (): string => {
    if (isResending) return "Resending...";
    if (resendCooldown > 0) return `Resend in ${formatTime(resendCooldown)}`;
    return "Resend verification email";
  };
  const resendButtonText = getResendButtonText();

  return (
    <div className="relative min-h-screen w-full overflow-hidden">
      <div
        className="fixed inset-0 bg-cover bg-center bg-no-repeat"
        style={{ backgroundImage: `url(${wolfImage})` }}
        aria-hidden="true"
      />
      <div
        className="fixed inset-0 bg-gradient-to-br from-slate-950/80 via-slate-950/70 to-slate-950/85"
        aria-hidden="true"
      />
      <div className="relative z-10 flex min-h-screen items-center justify-center px-4 py-8">
        <div className="w-full max-w-md">
          <div className="bg-slate-900/80 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-2xl">
            <div className="text-center mb-4">
              <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-rose-500/10">
                <Mail className="h-6 w-6 text-rose-500" />
              </div>
              <h2 className="text-xl font-semibold text-white mb-2">Verify your email</h2>
              <p className="text-slate-400 text-sm">We sent a verification link to:</p>
              <p className="text-rose-400 font-medium text-sm mt-1 break-all">{email}</p>
            </div>

            <div className="bg-slate-800/50 border border-white/5 rounded-xl p-4 mb-4">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Window:</span>
                <span className="text-rose-400 font-mono">{formatTime(timeLeft)}</span>
              </div>
            </div>

            <p className="text-slate-400 text-xs text-center mb-4">
              Click the link in your email to verify your account.
            </p>

            <button
              type="button"
              className="w-full bg-white/5 hover:bg-white/10 border border-white/10 text-gray-300 font-semibold py-3 rounded-xl transition-all text-xs text-center block mb-3"
              onClick={handleResend}
              disabled={resendCooldown > 0 || isResending}
            >
              {resendButtonText}
            </button>

            <button
              type="button"
              className="w-full bg-white/5 hover:bg-white/10 border border-white/10 text-gray-300 font-semibold py-3 rounded-xl transition-all text-xs text-center block"
              onClick={handleBackToLogin}
            >
              Change email address
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/waitVerify")({
  head: () => ({ meta: [{ title: "Verify - Vyzorix" }] }),
  validateSearch: (search: Record<string, unknown>) => {
    return {
      email: (search.email as string) ?? "",
      token: (search.token as string) ?? "",
      type: (search.type as string) ?? "verify",
    };
  },
  component: () => (
    <AuthLayout>
      <WaitVerifyPage />
    </AuthLayout>
  ),
});
