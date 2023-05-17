export interface SelectorEntry {
    sessionId: number
    keys: SelectorKey[]
}

export interface SelectorKey {
    fieldId: number
    value: string
    matchType: string
    valueMask?: string
    prefixLength?: number
}