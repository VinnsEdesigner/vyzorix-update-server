import { Github, Chrome, ArrowRight, Eye, EyeOff } from "lucide-react";
import { useState, type FormEvent, type ReactElement } from "react";

interface LoginFormProps {
  onLogin: (ident: string, pass: string) => void;
  onSSO: (provider: "GitHub" | "Google") => void;
  onForgotPassword: () => void;
  isSubmitting: boolean;
}

export default function LoginForm({
  onLogin,
  onSSO,
  onForgotPassword,
  isSubmitting,
}: LoginFormProps): ReactElement {
  const [loginIdent, setLoginIdent] = useState("");
  const [loginPass, setLoginPass] = useState("");
  const [showLoginPassword, setShowLoginPassword] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: FormEvent): void => {
    e.preventDefault();
    const newErrors: Record<string, string> = {};

    if (!loginIdent.trim()) {
      newErrors.loginIdent = "Username or Email address is required.";
    }
    if (!loginPass) {
      newErrors.loginPass = "Password is required.";
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    setErrors({});
    onLogin(loginIdent, loginPass);
  };

  return (
    <div id="login-form-container">
      <div className="mb-8 text-center md:text-left">
        <h1 className="text-2xl md:text-3xl font-semibold text-white tracking-tight">
          Welcome back
        </h1>
        <p className="text-slate-400 text-sm mt-2">Access your authorized database workspace.</p>
      </div>

      <form className="space-y-6" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label className="text-xs font-semibold text-slate-300 tracking-wide">
              Username or Email address
            </label>
            {errors.loginIdent && (
              <span className="text-xs text-rose-400 font-medium">{errors.loginIdent}</span>
            )}
          </div>
          <input
            className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500 focus:ring-1 focus:ring-rose-500/20"
            placeholder="alexis_handle"
            type="text"
            value={loginIdent}
            onChange={(e) => {
              setLoginIdent(e.target.value);
              if (errors.loginIdent) setErrors((prev) => ({ ...prev, loginIdent: "" }));
            }}
          />
        </div>

        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label className="text-xs font-semibold text-slate-300 tracking-wide">Password</label>
            {errors.loginPass && (
              <span className="text-xs text-rose-400 font-medium">{errors.loginPass}</span>
            )}
          </div>
          <div className="relative">
            <input
              className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 pr-11 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500"
              placeholder="••••••••"
              type={showLoginPassword ? "text" : "password"}
              value={loginPass}
              onChange={(e) => {
                setLoginPass(e.target.value);
                if (errors.loginPass) setErrors((prev) => ({ ...prev, loginPass: "" }));
              }}
            />
            <button
              type="button"
              onClick={() => setShowLoginPassword(!showLoginPassword)}
              className="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white transition-colors"
            >
              {showLoginPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
        </div>

        <div className="flex justify-end pt-1">
          <button
            type="button"
            className="text-xs text-muted-foreground hover:text-primary font-semibold underline underline-offset-4 cursor-pointer transition-colors"
            onClick={onForgotPassword}
          >
            Forgot your password?
          </button>
        </div>

        <button
          className="w-full bg-rose-600 hover:bg-rose-505 text-white font-semibold py-4 rounded-xl transition-all duration-300 cursor-pointer flex items-center justify-center gap-2 text-sm uppercase tracking-wider shadow-lg shadow-rose-955/25"
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
              <span>Authenticating Profile...</span>
            </>
          ) : (
            <>
              <span>Access Dashboard</span>
              <ArrowRight className="w-4 h-4" />
            </>
          )}
        </button>

        <div className="pt-6 border-t border-white/10 text-center">
          <p className="text-xs font-semibold text-slate-400 tracking-wider mb-4">
            Or login securely with Single Sign-On
          </p>
          <div className="flex gap-4 justify-center">
            <button
              className="flex items-center gap-3 px-6 py-3 border border-white/10 rounded-xl bg-white/5 hover:bg-white/10 transition-colors cursor-pointer text-xs font-semibold text-white"
              type="button"
              onClick={() => onSSO("GitHub")}
              disabled={isSubmitting}
            >
              <Github className="w-4.5 h-4.5 shrink-0 text-rose-500" />
              <span>GitHub</span>
            </button>
            <button
              className="flex items-center gap-3 px-6 py-3 border border-white/10 rounded-xl bg-white/5 hover:bg-white/10 transition-colors cursor-pointer text-xs font-semibold text-white"
              type="button"
              onClick={() => onSSO("Google")}
              disabled={isSubmitting}
            >
              <Chrome className="w-4.5 h-4.5 shrink-0 text-rose-500" />
              <span>Google</span>
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
