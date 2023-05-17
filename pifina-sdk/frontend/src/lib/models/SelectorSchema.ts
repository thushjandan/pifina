export interface SelectorSchema {
    id: number
    name: string
    matchType: string
    type: string
    width: number
}

export const FIELD_MATCH_PRIORITY = "$MATCH_PRIORITY"
export const MATCH_TYPE_EXACT   = "Exact"
export const MATCH_TYPE_TERNARY = "Ternary"
export const MATCH_TYPE_LPM     = "LPM"