export interface MetricData {
    [metricName: string]: MetricItem[]
}

export interface MetricItem {
    sessionId: number
    timestamp: Date
    value: number
}

export interface DTOPifinaMetricItem {
    sessionId: number
    metricName: string
    type: string
    value: number
    timestamp: string
}