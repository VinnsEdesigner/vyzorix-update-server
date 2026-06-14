import { Github, Chrome, ArrowRight, Eye, EyeOff } from "lucide-react";
import { useState, type FormEvent, type ReactElement } from "react";

interface SignUpFormProps {
  onSignUp: (data: { fullName: string; email: string; username: string; password: string }) => void;
  onSSO: (provider: "GitHub" | "Google") => void;
  isSubmitting: boolean;
  triggerToast: (msg: string, type?: "success" | "alert") => void;
}

export default function SignUpForm({
  onSignUp,
  onSSO,
  isSubmitting,
  triggerToast,
}: SignUpFormProps): ReactElement {
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [agreeTerms, setAgreeTerms] = useState(false);

  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: FormEvent): void => {
    e.preventDefault();
    const newErrors: Record<string, string> = {};

    if (!fullName.trim()) {
      newErrors.fullName = "Full Name is required.";
    } else if (fullName.trim().length < 3) {
      newErrors.fullName = "Must be at least 3 characters.";
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!email.trim()) {
      newErrors.email = "Email address is required.";
    } else if (!emailRegex.test(email)) {
      newErrors.email = "Invalid email format.";
    }

    if (!username.trim()) {
      newErrors.username = "Username is required.";
    } else if (username.trim().length < 3) {
      newErrors.username = "Must be at least 3 characters.";
    } else if (!/^[a-zA-Z0-9_]+$/.test(username)) {
      newErrors.username = "Alphanumerics and underscores only.";
    }

    if (!password) {
      newErrors.password = "Password is required.";
    } else if (password.length < 8) {
      newErrors.password = "Must be at least 8 characters.";
    }

    if (password !== confirmPassword) {
      newErrors.confirmPassword = "Passwords do not match.";
    }

    if (!agreeTerms) {
      newErrors.terms = "Terms agreement must be checked.";
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      triggerToast("Please solve form validation requirements.", "alert");
      return;
    }

    setErrors({});
    onSignUp({ fullName, email, username, password });
  };

  return (
    <div id="signup-form-container">
      <div className="mb-8 text-center md:text-left">
        <h1 className="text-2xl md:text-3xl font-semibold text-white tracking-tight">
          Get an operator account
        </h1>
      </div>

      <form className="space-y-6" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label className="text-xs font-semibold text-slate-300 tracking-wide">
              Your Full Name
            </label>
            {errors.fullName && (
              <span className="text-xs text-rose-400 font-medium">{errors.fullName}</span>
            )}
          </div>
          <input
            className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500 focus:ring-1 focus:ring-rose-500/20"
            placeholder="e.g. Alexis Thorne"
            type="text"
            value={fullName}
            onChange={(e) => {
              setFullName(e.target.value);
              if (errors.fullName) setErrors((prev) => ({ ...prev, fullName: "" }));
            }}
          />
        </div>

        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label className="text-xs font-semibold text-slate-300 tracking-wide">
              Your Email Address
            </label>
            {errors.email && (
              <span className="text-xs text-rose-400 font-medium">{errors.email}</span>
            )}
          </div>
          <input
            className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500 focus:ring-1 focus:ring-rose-500/20"
            placeholder="alexis@organization.com"
            type="email"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
              if (errors.email) setErrors((prev) => ({ ...prev, email: "" }));
            }}
          />
        </div>

        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label className="text-xs font-semibold text-slate-300 tracking-wide">
              Enter Your Username
            </label>
            {errors.username && (
              <span className="text-xs text-rose-400 font-medium">{errors.username}</span>
            )}
          </div>
          <input
            className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500 focus:ring-1 focus:ring-rose-500/20"
            placeholder="alexis_handle"
            type="text"
            value={username}
            onChange={(e) => {
              setUsername(e.target.value.replace(/\s+/g, ""));
              if (errors.username) setErrors((prev) => ({ ...prev, username: "" }));
            }}
          />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <label className="text-xs font-semibold text-slate-300 tracking-wide">Password</label>
              {errors.password && (
                <span className="text-xs text-rose-400 font-medium">{errors.password}</span>
              )}
            </div>
            <div className="relative">
              <input
                className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 pr-11 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500 focus:ring-1 focus:ring-rose-500/20"
                placeholder="••••••••"
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  if (errors.password) setErrors((prev) => ({ ...prev, password: "" }));
                }}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white transition-colors"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <label className="text-xs font-semibold text-slate-300 tracking-wide">Confirm</label>
              {errors.confirmPassword && (
                <span className="text-xs text-rose-400 font-medium">{errors.confirmPassword}</span>
              )}
            </div>
            <div className="relative">
              <input
                className="w-full bg-slate-900/65 border border-white/10 text-white placeholder-slate-500 px-4 pr-11 py-3.5 rounded-lg text-sm transition-all outline-none focus:border-rose-500"
                placeholder="••••••••"
                type={showConfirmPassword ? "text" : "password"}
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value);
                  if (errors.confirmPassword)
                    setErrors((prev) => ({ ...prev, confirmPassword: "" }));
                }}
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white transition-colors"
              >
                {showConfirmPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>
        </div>

        <div className="flex items-start gap-3 pt-2">
          <input
            className="mt-1 rounded border-white/20 bg-slate-900 text-rose-600 focus:ring-rose-500 text-sm cursor-pointer h-4 w-4"
            id="terms"
            type="checkbox"
            checked={agreeTerms}
            onChange={(e) => {
              setAgreeTerms(e.target.checked);
              if (errors.terms) setErrors((prev) => ({ ...prev, terms: "" }));
            }}
          />
          <div className="flex flex-col">
            <label
              className="text-slate-350 text-xs cursor-pointer select-none leading-relaxed"
              htmlFor="terms"
            >
              I agree to the{" "}
              <a
                className="text-rose-455 hover:text-rose-300 font-semibold underline underline-offset-2"
                href="#terms"
                onClick={(e) => {
                  e.preventDefault();
                  triggerToast("Agreement recorded successfully.", "success");
                }}
              >
                Terms of Service
              </a>{" "}
              and{" "}
              <a
                className="text-rose-455 hover:text-rose-300 font-semibold underline underline-offset-2"
                href="#privacy"
                onClick={(e) => {
                  e.preventDefault();
                  triggerToast("No tracking cookies saved.", "success");
                }}
              >
                Privacy Policies
              </a>
              .
            </label>
            {errors.terms && (
              <span className="text-xs text-rose-400 font-medium mt-1">{errors.terms}</span>
            )}
          </div>
        </div>

        <button
          className="w-full bg-rose-600 hover:bg-rose-500 text-white font-semibold py-4 rounded-xl transition-all duration-300 cursor-pointer flex items-center justify-center gap-2 text-sm uppercase tracking-wider shadow-lg shadow-rose-955/30"
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
              <span>Registering Profile...</span>
            </>
          ) : (
            <>
              <span>Create Account</span>
              <ArrowRight className="w-4 h-4" />
            </>
          )}
        </button>

        <div className="pt-6 border-t border-white/10 text-center">
          <p className="text-xs font-semibold text-slate-400 tracking-wider mb-4">
            Or register instantly via Single Sign-On
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
