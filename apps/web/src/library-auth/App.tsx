import { ArrowLeft } from "lucide-react";
import { useState, useEffect } from "react";

import wolfImage from "./assets/images/black_wolf_evening_1781264516831.jpg";
import ForgotPasswordForm from "./components/ForgotPasswordForm";
import LoginForm from "./components/LoginForm";
import OperatorDashboard from "./components/OperatorDashboard";
import SignUpForm from "./components/SignUpForm";
import SuccessView from "./components/SuccessView";
import WaitingVerification from "./components/WaitingVerification";
import {
  registerOperator,
  loginOperator,
  initiateSSO,
  requestPasswordReset,
  pollVerificationStatus,
  triggerTokenResend,
  cancelVerificationSession,
  getCurrentSession,
  handleSSOCallback,
  logoutOperator,
} from "./lib/api";
import { getHydratedState, ARCHITECTURE_CONFIG } from "./lib/config";
import { ViewMode, SuccessReport } from "./types";

export default function App() {
  // Navigation View State with SSR hydration capability
  const [view, setView] = useState<ViewMode>(() => {
    return getHydratedState<ViewMode>("view", "signup");
  });

  // Unified global values for verified profile report with SSR hydration capability
  const [profileData, setProfileData] = useState<{
    fullName: string;
    email: string;
    username: string;
  }>(() => {
    const hydrated = getHydratedState<{ fullName: string; email: string; username: string } | null>(
      "profileData",
      null,
    );
    if (hydrated) return hydrated;
    return { fullName: "", email: "", username: "" };
  });

  // Loading/Authenticating Polling Simulation
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [verificationStatus, setVerificationStatus] = useState(
    "Connecting to secure verification service...",
  );
  const [successReport, setSuccessReport] = useState<SuccessReport | null>(() => {
    return getHydratedState<SuccessReport | null>("successReport", null);
  });
  const [verificationToken, setVerificationToken] = useState<string | null>(() => {
    return getHydratedState<string | null>("verificationToken", null);
  });

  // Initial mount: capture handles, tokens, SSO callbacks, or fetch current active server session
  useEffect(() => {
    if (typeof window !== "undefined") {
      const params = new URLSearchParams(window.location.search);

      // A. Query parameters: Email Verification Link (?token=)
      const tokenVal = params.get("token");
      if (tokenVal) {
        setVerificationToken(tokenVal);
        setView("waiting_verification");
        window.history.replaceState({}, document.title, window.location.pathname);
        triggerToast("Security token detected from verification link.", "success");
        return;
      }

      // B. Query parameters: OAuth/SSO callback values (?code= & state= & provider=)
      const ssoCode = params.get("code");
      const ssoState = params.get("state");
      const ssoProvider =
        params.get("provider") || (window.location.search.includes("github") ? "GitHub" : "Google");

      if (ssoCode && ssoState) {
        setIsSubmitting(true);
        triggerToast(
          `Completing secure Single Sign-On handshake with ${ssoProvider}...`,
          "success",
        );

        handleSSOCallback(ssoProvider as "Google" | "GitHub", ssoCode, ssoState)
          .then((report) => {
            setIsSubmitting(false);
            setProfileData({
              fullName: report.fullName,
              email: report.email,
              username: report.username,
            });
            setSuccessReport(report);
            setView("success");
            window.history.replaceState({}, document.title, window.location.pathname);
            triggerToast("Successfully signed in via Single Sign-On!", "success");
          })
          .catch((err) => {
            setIsSubmitting(false);
            window.history.replaceState({}, document.title, window.location.pathname);
            triggerToast(err.message || `${ssoProvider} sign-in failed.`, "alert");
          });
        return;
      }

      // C. Safe Cookies Session Verification Lookup:
      // Query server if they are already logged in (when there are no redirection query params in URL)
      if (!ARCHITECTURE_CONFIG.IS_SIMULATED) {
        getCurrentSession()
          .then((session) => {
            if (session) {
              setProfileData({
                fullName: session.fullName,
                email: session.email,
                username: session.username,
              });
              setSuccessReport(session);
              setView("success");
              triggerToast("Restored active secure operator session.", "success");
            }
          })
          .catch(() => {
            // Unauthenticated lookup is expected, ignore errors silently
          });
      }
    }
  }, []);

  // Time remaining countdown clock (default 15 minutes, 900 seconds)
  const [timeLeft, setTimeLeft] = useState(900);

  // Unified Toast State with rose-600, white and black colors
  const [toast, setToast] = useState<{ message: string; type: "success" | "alert" } | null>(null);

  // Trigger modular non-green notification toasts
  const triggerToast = (message: string, type: "success" | "alert" = "success") => {
    setToast({ message, type });
    setTimeout(() => {
      setToast(null);
    }, 4500);
  };

  // Convert timer seconds to mm:ss format
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  // Automated background linkage detection & continuous database polling
  useEffect(() => {
    if (view === "waiting_verification") {
      if (ARCHITECTURE_CONFIG.IS_SIMULATED) {
        setVerificationStatus("Awaiting secure database link interaction...");

        const step1 = setTimeout(() => {
          setVerificationStatus("Incoming token signature detected from email client...");
        }, 1500);

        const step2 = setTimeout(() => {
          setVerificationStatus("Acoustic link handshake established. Setting up profile...");
        }, 3000);

        const finalRedirection = setTimeout(() => {
          const fallbackEmail = profileData.email || "sso-google@vyzorix.com";
          const fallbackUsername = profileData.username || "google_member";
          const fallbackFullName = profileData.fullName || "Enterprise Workspace User";

          setSuccessReport({
            fullName: fallbackFullName,
            email: fallbackEmail,
            username: fallbackUsername,
            memberId: "VXZ-64981",
            operatorRole: "Operator",
            region: "Paris, France",
            createdAt: new Date().toISOString().replace("T", " ").substring(0, 19) + " UTC",
            method: "Standard Email",
          });

          setView("success");
          triggerToast("Email verification authenticated successfully!", "success");
        }, 4500);

        return () => {
          clearTimeout(step1);
          clearTimeout(step2);
          clearTimeout(finalRedirection);
        };
      } else {
        // REAL REST POLLING MODE: Query SQLite DB continuously
        setVerificationStatus("Initiating live database polling sequence...");
        const tokenToPoll = verificationToken || "token_placeholder";
        let attempts = 0;

        const pollInterval = setInterval(() => {
          attempts++;
          setVerificationStatus(`Polling verification server (Ping #${attempts})...`);

          pollVerificationStatus(tokenToPoll)
            .then((res) => {
              if (res.status === "success" && res.report) {
                clearInterval(pollInterval);
                setSuccessReport(res.report);
                setView("success");
                triggerToast("Account verification authenticated successfully!", "success");
              }
            })
            .catch((err) => {
              setVerificationStatus(`Polling error: ${err.message || "connection failed"}`);
            });
        }, 2500);

        return () => {
          clearInterval(pollInterval);
        };
      }
    }
    return undefined;
  }, [view, profileData, verificationToken]);

  // Handle waiting countdown
  useEffect(() => {
    let timer: any;
    if (view === "waiting_verification" && timeLeft > 0) {
      timer = setInterval(() => {
        setTimeLeft((prev) => prev - 1);
      }, 1000);
    } else if (timeLeft === 0 && view === "waiting_verification") {
      triggerToast("Verification window expired. Please register again.", "alert");
      setView("signup");
    }
    return () => clearInterval(timer);
  }, [view, timeLeft]);

  // Create workspace action
  const handleSignUpAction = (data: { fullName: string; email: string; username: string }) => {
    setProfileData(data);
    setIsSubmitting(true);

    registerOperator(data)
      .then((res: any) => {
        setIsSubmitting(false);
        setTimeLeft(900);
        if (res?.token) {
          setVerificationToken(res.token);
        }
        setView("waiting_verification");
        triggerToast(res.message || "Verification link dispatched.", "success");
      })
      .catch((err) => {
        setIsSubmitting(false);
        triggerToast(err.message || "Registration failed", "alert");
      });
  };

  // Sign In action
  const handleLoginAction = (ident: string, pass: string) => {
    setIsSubmitting(true);

    loginOperator(ident, pass)
      .then((report) => {
        setIsSubmitting(false);
        setSuccessReport(report);
        setView("success");
        triggerToast("Welcome back! Authentication approved.", "success");
      })
      .catch((err) => {
        setIsSubmitting(false);
        triggerToast(err.message || "Authentication failed", "alert");
      });
  };

  // SSO connector
  const handleSSOAction = (provider: "GitHub" | "Google") => {
    triggerToast(`Initializing communication with ${provider}...`, "success");
    setIsSubmitting(true);

    initiateSSO(provider)
      .then((report) => {
        setIsSubmitting(false);
        setProfileData({
          fullName: report.fullName,
          email: report.email,
          username: report.username,
        });

        if (ARCHITECTURE_CONFIG.IS_SIMULATED) {
          setSuccessReport(report);
          setView("success");
          triggerToast(`Successfully signed in via ${provider} Single Sign-On!`, "success");
        } else {
          setTimeLeft(900);
          setView("waiting_verification");
          triggerToast(`${provider} credentials loaded. Pending secure validation...`, "success");
        }
      })
      .catch((err) => {
        setIsSubmitting(false);
        if (err.message && !err.message.includes("Redirecting")) {
          triggerToast(err.message || "SSO link failed", "alert");
        }
      });
  };

  // Recover password action
  const handleForgotPasswordAction = (email: string) => {
    setIsSubmitting(true);
    requestPasswordReset(email)
      .then((res) => {
        setIsSubmitting(false);
        triggerToast(res.message, "success");
      })
      .catch((err) => {
        setIsSubmitting(false);
        triggerToast(err.message || "Reset request failed", "alert");
      });
  };

  // Hard Reset Application session
  const clearSession = () => {
    logoutOperator()
      .then(() => {
        setProfileData({ fullName: "", email: "", username: "" });
        setSuccessReport(null);
        setView("signup");
        setTimeLeft(900);
        triggerToast("Workspace registration session recycled.", "success");
      })
      .catch(() => {
        // Fallback reset in case of connection failure
        setProfileData({ fullName: "", email: "", username: "" });
        setSuccessReport(null);
        setView("signup");
        setTimeLeft(900);
        triggerToast("Workspace offline session cleared.", "success");
      });
  };

  return (
    <div
      className="relative min-h-screen text-slate-100 font-sans overflow-x-hidden flex flex-col justify-between selection:bg-rose-600 selection:text-white"
      style={{
        backgroundImage: `linear-gradient(rgba(10, 10, 15, 0.75), rgba(10, 10, 15, 0.92)), url(${wolfImage})`,
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundAttachment: "fixed",
      }}
    >
      {/* Toast Notifier in strict black, white, and rose color scheme */}
      {toast && (
        <div className="fixed top-8 left-1/2 -translate-x-1/2 z-50 px-6 py-4 rounded-xl border-t-2 bg-slate-950/95 border-rose-600/40 text-slate-100 shadow-2xl flex items-center gap-4 text-xs max-w-sm w-[90%] transition-all animate-fade-in">
          <span className="shrink-0 h-2.5 w-2.5 rounded-full bg-rose-600 animate-pulse"></span>
          <p className="flex-grow font-semibold leading-relaxed">{toast.message}</p>
        </div>
      )}

      {/* Header element */}
      <header className="w-full top-0 sticky z-40 bg-slate-950/40 backdrop-blur-md border-b border-white/5">
        <div className="flex justify-between items-center px-8 py-5 max-w-7xl mx-auto w-full">
          <div
            onClick={clearSession}
            className="flex items-center gap-3 cursor-pointer hover:opacity-90 transition-all select-none"
          >
            <div className="h-3 w-3 bg-rose-600 rounded-full shadow-[0_0_12px_#f43f5e]"></div>
            <span className="font-extrabold text-xl tracking-wider uppercase text-white font-sans">
              Vyzorix
            </span>
          </div>

          <div className="flex items-center gap-5">
            {view === "signup" && (
              <>
                <span className="text-slate-400 text-xs hidden sm:inline font-normal">
                  Already have an account?
                </span>
                <button
                  className="bg-white/5 hover:bg-white/10 border border-white/10 text-white font-medium text-xs py-2 px-5 rounded-lg cursor-pointer transition-colors"
                  onClick={() => setView("login")}
                >
                  Log In
                </button>
              </>
            )}
            {view === "login" && (
              <>
                <span className="text-slate-400 text-xs hidden sm:inline font-normal">
                  New to Vyzorix?
                </span>
                <button
                  className="bg-rose-600 hover:bg-rose-500 text-white font-semibold text-xs py-2 px-5 rounded-lg cursor-pointer transition-colors shadow-lg shadow-rose-950/20"
                  onClick={() => setView("signup")}
                >
                  Sign Up
                </button>
              </>
            )}
            {(view === "forgot_password" || view === "waiting_verification") && (
              <button
                className="text-slate-400 hover:text-white font-medium text-xs flex items-center gap-2 transition-colors"
                onClick={() => setView("login")}
              >
                <ArrowLeft className="w-4 h-4" />
                <span>Return to Login</span>
              </button>
            )}
            {view === "success" && (
              <button
                className="bg-white/5 hover:bg-white/10 border border-white/10 text-rose-455 font-semibold text-xs py-2 px-5 rounded-lg cursor-pointer transition-colors"
                onClick={clearSession}
              >
                Reset Session
              </button>
            )}
          </div>
        </div>
      </header>

      {/* Main Form Center Layout */}
      <main className="flex-grow flex items-center justify-center p-8 z-10 my-12">
        <div className={`w-full ${view === "dashboard" ? "max-w-6xl" : "max-w-lg"}`}>
          {/* Main Card */}
          <div className="bg-slate-950/75 backdrop-blur-2xl border border-white/10 rounded-2xl p-8 md:p-10 shadow-2xl transition-all duration-305 relative before:absolute before:inset-0 before:bg-gradient-to-b before:from-white/5 before:to-transparent before:rounded-2xl before:pointer-events-none">
            {/* 1. SIGN UP SECTION */}
            {view === "signup" && (
              <SignUpForm
                onSignUp={handleSignUpAction}
                onSSO={handleSSOAction}
                isSubmitting={isSubmitting}
                triggerToast={triggerToast}
              />
            )}

            {/* 2. LOGIN SECTION */}
            {view === "login" && (
              <LoginForm
                onLogin={handleLoginAction}
                onSSO={handleSSOAction}
                onForgotPassword={() => setView("forgot_password")}
                isSubmitting={isSubmitting}
              />
            )}

            {/* 3. FORGOT PASSWORD SECTION */}
            {view === "forgot_password" && (
              <ForgotPasswordForm
                onResetSubmit={handleForgotPasswordAction}
                onBackToLogin={() => setView("login")}
                isSubmitting={isSubmitting}
                resetSent={false}
                setResetSent={() => {}}
              />
            )}

            {/* 4. WAITING VERIFICATION */}
            {view === "waiting_verification" && (
              <WaitingVerification
                email={profileData.email}
                timeLeft={timeLeft}
                formatTime={formatTime}
                statusText={verificationStatus}
                onResend={() => {
                  setTimeLeft(900);
                  if (ARCHITECTURE_CONFIG.IS_SIMULATED) {
                    triggerToast(
                      `New verification token transmitted to ${profileData.email}`,
                      "success",
                    );
                  } else {
                    setIsSubmitting(true);
                    triggerTokenResend(profileData.email)
                      .then((res: any) => {
                        setIsSubmitting(false);
                        if (res?.token) {
                          setVerificationToken(res.token);
                        }
                        triggerToast(
                          res.message ||
                            `New verification token transmitted to ${profileData.email}`,
                          "success",
                        );
                      })
                      .catch((err) => {
                        setIsSubmitting(false);
                        triggerToast(
                          err.message || "Failed to resend verification token.",
                          "alert",
                        );
                      });
                  }
                }}
                onCancel={() => {
                  if (ARCHITECTURE_CONFIG.IS_SIMULATED) {
                    setView("signup");
                    triggerToast("Verification cancelled. Registration unlocked.", "alert");
                  } else {
                    setIsSubmitting(true);
                    cancelVerificationSession(profileData.email)
                      .then(() => {
                        setIsSubmitting(false);
                        setView("signup");
                        triggerToast("Verification cancelled. Registration unlocked.", "alert");
                      })
                      .catch((err) => {
                        setIsSubmitting(false);
                        setView("signup");
                        triggerToast(err.message || "Verification session aborted.", "alert");
                      });
                  }
                }}
              />
            )}

            {/* 5. SUCCESS CARD */}
            {view === "success" && successReport && (
              <SuccessView
                successReport={successReport}
                onProceed={() => {
                  setView("dashboard" as ViewMode);
                  triggerToast("Entering secure core operator dashboard...", "success");
                }}
              />
            )}

            {/* 6. OPERATOR DASHBOARD */}
            {view === "dashboard" && <OperatorDashboard />}
          </div>
        </div>
      </main>

      {/* Footer Nav Bar */}
      <footer className="w-full bg-slate-950/45 backdrop-blur-md border-t border-white/10 py-6">
        <div className="max-w-7xl mx-auto px-8 flex flex-col md:flex-row justify-between items-center gap-4 text-slate-400 text-xs font-normal">
          <div>™ 2026 Vyzorix All rights reserved. v0.0.11</div>
          <div className="flex gap-4">
            <a
              className="hover:text-rose-450 transition-colors font-medium cursor-pointer"
              href="#privacy"
              onClick={(e) => {
                e.preventDefault();
                triggerToast("Zero tracking logs cached.", "success");
              }}
            >
              Privacy Policy
            </a>
            <a
              className="hover:text-rose-455 transition-colors font-medium cursor-pointer"
              href="#terms"
              onClick={(e) => {
                e.preventDefault();
                triggerToast("Standard corporate policies apply.", "success");
              }}
            >
              Terms of Use
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
