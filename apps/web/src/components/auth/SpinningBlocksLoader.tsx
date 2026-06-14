import type { ReactElement } from "react";

export default function SpinningBlocksLoader(): ReactElement {
  return (
    <div className="flex flex-col items-center justify-center py-6" id="spinning-blocks-loader">
      <style>{`
        @keyframes block-large-spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
        @keyframes block-pulse {
          0%, 100% { transform: scale(0.95); opacity: 0.9; }
          50% { transform: scale(1.05); opacity: 1; }
        }
        .anim-block-large-spin {
          animation: block-large-spin 3s cubic-bezier(0.4, 0, 0.2, 1) infinite;
        }
        .anim-block-pulse {
          animation: block-pulse 2s ease-in-out infinite;
        }
      `}</style>

      <div className="relative w-16 h-16 flex items-center justify-center">
        {/* Outer rotating square block */}
        <div className="absolute inset-0 border-3 border-white rounded-[10px] anim-block-large-spin"></div>

        {/* Inner pulsing solid red core block */}
        <div className="absolute inset-4 bg-rose-600 rounded-[5px] anim-block-pulse"></div>
      </div>
    </div>
  );
}
