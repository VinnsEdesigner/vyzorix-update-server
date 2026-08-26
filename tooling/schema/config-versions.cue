package vyzorix

// ConfigVersion — generated from openapi/schemas.go.
#ConfigVersion: {
        created_at: string
        snapshot: {[string]: _}
        id: string
        org_id: string
        resource_type: string
        changed_by: string
        version: int
}

// ConfigVersionListResult — generated from openapi/schemas.go.
#ConfigVersionListResult: {
        versions: [...#ConfigVersion]
}

// ConfigVersionRestoreResult — generated from openapi/schemas.go.
#ConfigVersionRestoreResult: {
        settings: {[string]: _}
        restored_to_version: int
}
