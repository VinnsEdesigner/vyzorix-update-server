package vyzorix

// ConfigVersion.
#ConfigVersion: {
        created_at: int64
        snapshot: string
        id: string
        org_id: string
        resource_type: string
        changed_by: string
        version: int
}

// ConfigVersionListResult.
#ConfigVersionListResult: {
        versions: [...#ConfigVersion]
}

// ConfigVersionRestoreResult.
#ConfigVersionRestoreResult: {
        settings: {[string]: _}
        restored_to_version: int
}
