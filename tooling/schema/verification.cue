package vyzorix

// PollVerificationResult — generated from openapi/schemas.go.
#PollVerificationResult: {
        status: string
        email?: string
        emailError?: string
}

// ResendVerificationRequest — generated from openapi/schemas.go.
#ResendVerificationRequest: {
        email: string
}

// CancelVerificationRequest — generated from openapi/schemas.go.
#CancelVerificationRequest: {
        email: string
}
