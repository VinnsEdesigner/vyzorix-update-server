package vyzorix

// ForgotPasswordRequest.
#ForgotPasswordRequest: {
        email: string
}

// ResetPasswordRequest.
#ResetPasswordRequest: {
        token: string
        newPassword: string
}
