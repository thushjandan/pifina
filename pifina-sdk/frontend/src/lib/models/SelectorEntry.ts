// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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