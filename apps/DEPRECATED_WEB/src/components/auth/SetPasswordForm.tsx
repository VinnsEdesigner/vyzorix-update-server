import { Eye, EyeOff, ArrowLeft } from "lucide-react";
import { useState, type FormEvent, type ReactElement } from "react";

interface SetPasswordFormProps {
  onSubmit: (password: string) => Promise<void>;
  onBackToLogin: () => void;
  isSubmitting: boolean;
}

export default function SetPasswordForm({
  onSubmit,
  onBackToLogin,
  isSubmitting,
}: SetPasswordFormProps): ReactElement {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = (e: FormEvent): void => {
    e.preventDefault();
    setError("");

    if (!password) {
      setError("Password is required.");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    onSubmit(password);
  };

  return (
    <div id="set-password-container">
      <div className="mb-6 text-center md:text-left">
        <h1 className="text-2xl font-semibold text-white tracking-tight">Create New Password</h1>
        <p className="text-slate-400 text-sm mt-3 leading-relaxed">
          Enter a new secure password for your account.
        </p>
      </div>

      <form className="space-y-6" onSubmit={handleSubmit}>
        {/* Password */}
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label className="text-xs font-semibold text-slate-300 tracking-wide">
              New Password
            </label>
            {error && password && (
              <span className="text-xs text-rose-400 font-medium">{error}</span>
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
                if (error) setError("");
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

        {/* Confirm Password */}
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <label className="text-xs font-semibold text-slate-300 tracking-wide">
              Confirm Password
            </label>
            {error && confirmPassword && (
              <span className="text-xs text-rose-400 font-medium">{error}</span>
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
                if (error) setError("");
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

        {/* Password Requirements */}
        <div className="p-3 bg-white/5 rounded-xl text-xs text-slate-400 leading-relaxed border border-white/5">
          <b className="text-slate-300">Requirements:</b> At least 8 characters long.
        </div>

        {/* Submit */}
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
              <span>Setting Password...</span>
            </>
          ) : (
            <span>Set New Password</span>
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
          <span>Back to login</span>
        </button>
      </div>
    </div>
  );
}
