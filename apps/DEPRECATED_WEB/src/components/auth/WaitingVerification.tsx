import { ShieldCheck } from "lucide-react";
import type { ReactElement } from "react";

import SpinningBlocksLoader from "./SpinningBlocksLoader";

interface WaitingVerificationProps {
  email: string;
  timeLeft: number;
  formatTime: (sec: number) => string;
  onResend: () => void;
  onCancel: () => void;
  statusText?: string;
}

export default function WaitingVerification({
  email,
  timeLeft,
  formatTime,
  onResend,
  onCancel,
  statusText,
}: WaitingVerificationProps): ReactElement {
  return (
    <div className="space-y-6 py-1 text-slate-100" id="waiting-verification-panel">
      <div className="text-left">
        <button
          onClick={onCancel}
          className="text-xs text-rose-500 hover:text-rose-400 font-semibold underline underline-offset-4 cursor-pointer transition-colors"
        >
          Return to Login
        </button>
      </div>

      <div className="text-center space-y-4 pt-1">
        <SpinningBlocksLoader />

        <div className="space-y-2 font-sans">
          <h2 className="text-2xl font-semibold text-white tracking-tight">Verify your account</h2>
          <p className="text-slate-350 text-xs md:text-sm max-w-sm mx-auto leading-relaxed">
            A verification link was sent to:
            <span className="block text-rose-400 font-semibold mt-1 break-all font-sans">
              {email || "sso-github@vyzorix.com"}
            </span>
          </p>
        </div>

        <div className="bg-slate-900/40 border border-white/5 rounded-xl py-3 px-5 max-w-xs mx-auto flex items-center justify-between gap-4 shadow-md text-xs leading-none">
          <span className="text-slate-400 font-semibold tracking-wide">Verification Window:</span>
          <span className="text-sm text-rose-500 font-bold font-sans tracking-wider">
            {formatTime(timeLeft)}
          </span>
        </div>
      </div>

      <div className="border border-white/10 rounded-xl p-5 bg-slate-950/40 space-y-3.5 text-xs text-left">
        <div className="flex items-center gap-2 border-b border-white/5 pb-2">
          <ShieldCheck className="w-4 h-4 text-rose-500 shrink-0" />
          <span className="font-semibold text-slate-200 uppercase tracking-wider text-[10px]">
            Verification Service
          </span>
        </div>

        <p className="text-slate-350 text-xs leading-relaxed font-normal">
          Please open your email inbox and confirm by clicking the link.
        </p>

        <div className="p-3 bg-rose-950/10 border border-rose-500/10 rounded-lg text-xs flex items-center gap-2.5">
          <div className="relative w-4 h-4 shrink-0">
            <style>{`
              @keyframes inline-micro-spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
              }
              .anim-inline-micro {
                animation: inline-micro-spin 2s linear infinite;
              }
            `}</style>
            <div className="absolute inset-0 border-2 border-white rounded-[3px] anim-inline-micro"></div>
            <div className="absolute inset-1 bg-rose-600 rounded-[1px] animate-pulse"></div>
          </div>
          <span className="font-semibold text-slate-300">
            {statusText ?? "Setting up profile..."}
          </span>
        </div>
      </div>

      <div className="space-y-3 pt-1">
        <button
          onClick={onResend}
          className="w-full bg-white/5 hover:bg-white/10 border border-white/10 text-gray-300 font-semibold py-3 rounded-xl transition-all text-xs text-center block cursor-pointer active:scale-[0.99]"
        >
          Resend Verification Email
        </button>

        <button
          onClick={onCancel}
          className="text-slate-400 hover:text-white transition-colors text-xs text-center block mx-auto underline cursor-pointer"
        >
          Change email address
        </button>
      </div>
    </div>
  );
}
