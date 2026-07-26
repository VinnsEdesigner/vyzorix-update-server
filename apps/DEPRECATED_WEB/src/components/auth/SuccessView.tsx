import { CheckCircle2, ArrowRight } from "lucide-react";
import type { ReactElement } from "react";

interface SuccessViewProps {
  successReport: {
    fullName: string;
    email: string;
    username: string;
    memberId: string;
    operatorRole: string;
    region: string;
    createdAt: string;
    method: string;
  };
  onProceed: () => void;
}

export default function SuccessView({ successReport, onProceed }: SuccessViewProps): ReactElement {
  return (
    <div className="space-y-8 text-slate-100" id="success-view">
      <div className="text-center space-y-3">
        <div className="h-16 w-16 rounded-full bg-rose-500/10 border border-rose-500/30 text-rose-500 mx-auto flex items-center justify-center shadow-xl shadow-rose-955/20 animate-fade-in">
          <CheckCircle2 className="w-8 h-8" />
        </div>
        <div className="space-y-1">
          <span className="text-[10px] tracking-wider font-bold uppercase text-white bg-rose-600/20 px-3 py-1 rounded-full border border-rose-600/35">
            Verified
          </span>
          <h2 className="text-2xl font-bold text-white tracking-tight mt-3">Welcome</h2>
          <p className="text-slate-400 text-xs">Login Successful</p>
        </div>
      </div>

      <div className="border border-white/10 rounded-2xl p-6 bg-slate-950/40 space-y-5 shadow-2xl relative before:absolute before:inset-0 before:bg-gradient-to-b before:from-white/5 before:to-transparent before:rounded-2xl before:pointer-events-none">
        <div className="space-y-4 text-xs font-sans">
          <div className="flex justify-between items-center py-1">
            <span className="text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              Vyzorix member
            </span>
            <span className="text-white font-semibold text-sm">
              @{successReport.username || "google_member"}
            </span>
          </div>

          <div className="flex justify-between items-center py-1 border-t border-white/5 pt-3">
            <span className="text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              Your email
            </span>
            <span className="text-white font-medium select-all break-all">
              {successReport.email || "sso-google@vyzorix.com"}
            </span>
          </div>

          <div className="flex justify-between items-center py-1 border-t border-white/5 pt-3">
            <span className="text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              ID
            </span>
            <span className="text-slate-200 font-mono font-bold">
              {successReport.memberId || "VXZ-64981"}
            </span>
          </div>

          <div className="flex justify-between items-center py-1 border-t border-white/5 pt-3">
            <span className="text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              Role
            </span>
            <span className="text-rose-400 font-semibold">
              {successReport.operatorRole || "Operator"}
            </span>
          </div>

          <div className="flex justify-between items-center py-1 border-t border-white/5 pt-3">
            <span className="text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              Region
            </span>
            <span className="text-slate-300 font-medium">
              {successReport.region || "Paris, Île-de-France, France"}
            </span>
          </div>

          <div className="flex justify-between items-center py-1 border-t border-white/5 pt-3">
            <span className="text-slate-400 font-semibold uppercase tracking-wider text-[10px]">
              Session Date
            </span>
            <span className="text-slate-350 font-medium">2026-06-12 12:29:27 UTC</span>
          </div>
        </div>
      </div>

      <div className="pt-2">
        <button
          onClick={onProceed}
          className="w-full bg-rose-600 hover:bg-rose-500 text-white font-semibold py-4 rounded-xl active:scale-[0.98] transition-all flex items-center justify-center gap-2.5 text-xs uppercase tracking-widest cursor-pointer shadow-lg shadow-rose-955/35 font-sans"
        >
          <span>Proceed</span>
          <ArrowRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
