package vyzorix

// ContactPointRequest — generated from openapi/schemas.go.
#ContactPointRequest: {
        name: string
        channel: string
        secret: string
        config: {[string]: _}
        template_id: string
        enabled: bool
}

// ContactPoint — generated from openapi/schemas.go.
#ContactPoint: {
        created_at: string
        updated_at: string
        config: {[string]: _}
        id: string
        org_id: string
        name: string
        channel: string
        template_id: string
        secret: bool
        enabled: bool
}

// ContactPointListResult — generated from openapi/schemas.go.
#ContactPointListResult: {
        contact_points: [...#ContactPoint]
}

// ContactPointTestResult — generated from openapi/schemas.go.
#ContactPointTestResult: {
        tested_at: string
        sent: bool
}
