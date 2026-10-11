package vyzorix

// AnnotationRequest.
#AnnotationRequest: {
        title: string
        text: string
        source: string
        start_time: string
        end_time: string
        tags: [...string]
}

// Annotation.
#Annotation: {
        start_time: int64
        created_at: int64
        updated_at: int64
        end_time?: int64
        id: string
        org_id: string
        title: string
        text: string
        source: string
        tags: [...string]
}

// AnnotationListResult.
#AnnotationListResult: {
        annotations: [...#Annotation]
}
