import { KeyRound, ArrowLeft } from "lucide-react";
import { useState, type FormEvent, type ReactElement } from "react";

interface ForgotPasswordFormProps {
  onResetSubmit: (email: string) => void;
  onBackToLogin: () => void;
  isSubmitting: boolean;
}

export default function ForgotPasswordForm({
  onResetSubmit,
  onBackToLogin,
  isSubmitting,
}: ForgotPasswordFormProps): ReactElement {
  const [resetEmail, setResetEmail] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = (e: FormEvent): void => {
    e.preventDefault();
    if (!resetEmail.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(resetEmail)) {
      setError("Provide a valid corporate email address.");
      return;
    }
    setError("");
    onResetSubmit(resetEmail);
  };

  return (
    <div id="forgot-password-container">
      <div className="mb-6 text-center md:text-left">
        <h1 className="text-2xl font-semibold text-white tracking-tight flex items-center justify-center md:justify-start gap-2.5">
          <KeyRound className="w-6 h-6 text-rose-500" />
          Reset your password
        </h1>
        <p className="text-slate-400 text-sm mt-3 leading-relaxed">
          Enter your email address to carry out a secure directory password restoration link.
        </p>
      </div>

      <form className="space-y-6" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label
              className="text-xs font-semibold text-slate-300 tracking-wide"
              htmlFor="reset-email"
            >
              Email Address
            </label>
            {error && <span className="text-xs text-rose-455 font-medium">{error}</span>}
          </div>
          <input
            id="reset-email"
            className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500"
            placeholder="alexis@organization.com"
            type="email"
            value={resetEmail}
            onChange={(e) => {
              setResetEmail(e.target.value);
              if (error) setError("");
            }}
            disabled={isSubmitting}
          />
        </div>

        <button
          className="w-full bg-rose-600 hover:bg-rose-500 text-white font-semibold py-4 rounded-xl transition-all duration-300 cursor-pointer flex items-center justify-center gap-2.5 text-sm uppercase tracking-wider shadow-lg shadow-rose-955/35 disabled:opacity-50 disabled:cursor-not-allowed"
          type="submit"
          disabled={isSubmitting}
        >
          {isSubmitting ? (
            <>
              <div className="relative w-4 h-4 mr-2">
                <style>{`
                  @keyframes inline-mini-spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                  }
                  .anim-inline-mini {
                    animation: inline-mini-spin 2s linear infinite;
                  }
                `}</style>
                <div className="absolute inset-0 border-2 border-white rounded-[3px] anim-inline-mini"></div>
                <div className="absolute inset-1 bg-rose-600 rounded-[1px] animate-pulse"></div>
              </div>
              <span>Sending...</span>
            </>
          ) : (
            <span>Send Recovery Link</span>
          )}
        </button>
      </form>

      <div className="mt-8 pt-6 border-t border-white/10 text-center">
        <button
          type="button"
          className="text-slate-400 hover:text-white font-medium text-xs flex items-center justify-center gap-2 mx-auto transition-colors"
          onClick={onBackToLogin}
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to workspace login</span>
        </button>
      </div>
    </div>
  );
}
