package vyzorix

// ContactPointRequest.
#ContactPointRequest: {
        name: string
        channel: string
        secret: string
        config: {[string]: string}
        template_id: string
        enabled: bool
}

// ContactPoint.
#ContactPoint: {
        created_at: int64
        updated_at: int64
        config: {[string]: string}
        id: string
        org_id: string
        name: string
        channel: string
        template_id: string
        secret: bool
        enabled: bool
}

// ContactPointListResult.
#ContactPointListResult: {
        contact_points: [...#ContactPoint]
}

// ContactPointTestResult.
#ContactPointTestResult: {
        tested_at: int64
        sent: bool
        deliveries?: _
}
