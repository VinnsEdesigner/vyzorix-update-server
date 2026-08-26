package vyzorix

// ClientCredential — generated from openapi/schemas.go.
#ClientCredential: {
        id: string
        name: string
        clientId: string
        secret?: string
        createdAt?: string
}

// ClientCredentialListResult — generated from openapi/schemas.go.
#ClientCredentialListResult: {
        clients: [...#ClientCredential]
}

// CreateClientCredentialRequest — generated from openapi/schemas.go.
#CreateClientCredentialRequest: {
        name: string
}

// UpdateClientCredentialRequest — generated from openapi/schemas.go.
#UpdateClientCredentialRequest: {
        name?: string
}
