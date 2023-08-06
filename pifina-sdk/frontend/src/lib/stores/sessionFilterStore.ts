import { writable } from "svelte/store";

export const sessionFilterStore = writable<number[]>([]);