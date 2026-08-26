package vyzorix

// AnnotationRequest — generated from openapi/schemas.go.
#AnnotationRequest: {
        title: string
        text: string
        source: string
        start_time: string
        end_time: string
        tags: [...string]
}

// Annotation — generated from openapi/schemas.go.
#Annotation: {
        start_time: string
        created_at: string
        updated_at: string
        end_time?: string
        id: string
        org_id: string
        title: string
        text: string
        source: string
        tags: [...string]
}

// AnnotationListResult — generated from openapi/schemas.go.
#AnnotationListResult: {
        annotations: [...#Annotation]
}
