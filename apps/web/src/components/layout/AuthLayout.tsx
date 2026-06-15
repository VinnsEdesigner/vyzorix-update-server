import { type ReactElement, type ReactNode } from "react";

import wolfImage from "@/assets/images/black_wolf_evening_1781264516831.jpg";

interface AuthLayoutProps {
  children: ReactNode;
}

/**
 * AuthLayout - Shared layout component for authentication pages
 * 
 * Features:
 * - Wolf background image
 * - Centered content
 * - Consistent auth page styling
 */
export const AuthLayout = ({ children }: AuthLayoutProps): ReactElement => {
  return (
    <div
      className="relative min-h-screen w-full overflow-hidden"
      style={{
        backgroundImage: `url(${wolfImage})`,
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundRepeat: "no-repeat",
      }}
    >
      {/* Dark overlay for readability */}
      <div className="absolute inset-0 bg-black/60" />
      
      {/* Content */}
      <div className="relative z-10 flex min-h-screen items-center justify-center px-4">
        {children}
      </div>
    </div>
  );
};

export default AuthLayout;
