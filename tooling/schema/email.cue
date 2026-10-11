package vyzorix

// EmailVerifyRequest.
#EmailVerifyRequest: {
        token: string
}

// EmailVerifyResult.
#EmailVerifyResult: {
        email?: string
        verified: bool
}

// EmailNotifications.
#EmailNotifications: {
        thresholdBreach: bool
        deviceOffline: bool
        deviceOnline: bool
        updateAvailable: bool
        commandFailed: bool
        registrationRequest: bool
}
