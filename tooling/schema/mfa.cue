package vyzorix

// MFAStatusResult — generated from openapi/schemas.go.
#MFAStatusResult: {
        mfa_enabled: bool
}

// MFAEnrollResult — generated from openapi/schemas.go.
#MFAEnrollResult: {
        secret: string
        uri: string
}

// MFAVerifySetupRequest — generated from openapi/schemas.go.
#MFAVerifySetupRequest: {
        code: string
        token: string
}

// MFAEnableRequest — generated from openapi/schemas.go.
#MFAEnableRequest: {
        code: string
        token: string
}

// MFAEnableResult — generated from openapi/schemas.go.
#MFAEnableResult: {
        backup_codes: [...string]
        success: bool
}

// MFADisableRequest — generated from openapi/schemas.go.
#MFADisableRequest: {
        code: string
}

// MFABackupCodeRequest — generated from openapi/schemas.go.
#MFABackupCodeRequest: {
        code: string
}

// MFABackupCodeResult — generated from openapi/schemas.go.
#MFABackupCodeResult: {
        valid: bool
}

// MFARegenerateResult — generated from openapi/schemas.go.
#MFARegenerateResult: {
        backup_codes: [...string]
}

// MFAVerifyRequest — generated from openapi/schemas.go.
#MFAVerifyRequest: {
        operator_id: string
        code: string
}

// MFAVerifyResult — generated from openapi/schemas.go.
#MFAVerifyResult: {
        session_id: string
        access_token: string
        refresh_token: string
        signing_key: string
        operator: #MFAOperator
        expires_at: int64
        success: bool
}

// MFAOperator — generated from openapi/schemas.go.
#MFAOperator: {
        id: string
        email: string
        name: string
        role: string
        mfa_enabled: bool
}
