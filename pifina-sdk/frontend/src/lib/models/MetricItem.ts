export interface MetricData {
    [metricName: string]: MetricItem[]
}

export interface MetricItem {
    sessionId: number
    timestamp: Date
    value: number
    type: string
}

export interface DTOPifinaMetricItem {
    sessionId: number
    metricName: string
    type: string
    value: number
    timestamp: string
}

export interface MetricNameGroup {
    [key: string]: Set<string>
}