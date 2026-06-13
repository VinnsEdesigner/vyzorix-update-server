export type ViewMode = "signup" | "login" | "forgot_password" | "waiting_verification" | "success" | "dashboard";

export interface SuccessReport {
  fullName: string;
  email: string;
  username: string;
  memberId: string;
  operatorRole: string;
  region: string;
  createdAt: string;
  method: "SSO" | "Standard Email";
}
