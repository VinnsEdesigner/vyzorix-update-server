package vyzorix

// EmailVerifyRequest — generated from openapi/schemas.go.
#EmailVerifyRequest: {
        token: string
}

// EmailVerifyResult — generated from openapi/schemas.go.
#EmailVerifyResult: {
        email?: string
        verified: bool
}

// EmailNotifications — generated from openapi/schemas.go.
#EmailNotifications: {
        thresholdBreach: bool
        deviceOffline: bool
        deviceOnline: bool
        updateAvailable: bool
        commandFailed: bool
        registrationRequest: bool
}
