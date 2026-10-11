package vyzorix

// PollVerificationResult.
#PollVerificationResult: {
        status: string
        email?: string
        emailError?: string
}

// ResendVerificationRequest.
#ResendVerificationRequest: {
        email: string
}

// CancelVerificationRequest.
#CancelVerificationRequest: {
        email: string
}
