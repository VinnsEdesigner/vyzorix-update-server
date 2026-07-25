

export interface VerifyEmailResponse {
  verified: boolean;
  email?: string;
}

export interface ResendVerificationResponse {
  message: string;
}

export interface CancelVerificationResponse {
  success: boolean;
}

export interface PollVerificationResponse {
  verified: boolean;
  email?: string;
}
