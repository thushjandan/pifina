<script lang="ts">
	import { onDestroy } from "svelte";
	import { endpointAddress } from "./EndpointStore";
	import { FIELD_MATCH_PRIORITY, MATCH_TYPE_EXACT, MATCH_TYPE_LPM, MATCH_TYPE_TERNARY, type SelectorSchema } from "$lib/models/SelectorSchema";
	import type { SelectorEntry, SelectorKey } from "$lib/models/SelectorEntry";
	import { each } from "svelte/internal";
	import { select } from "d3";
	import { goto } from "$app/navigation";

    let localEndpointAddress: string = "";
    let endpointPromise = Promise.resolve<SelectorEntry[]>([]);
    let matchSelectorSchema: SelectorSchema[] = [];

    const endpointAddrSub = endpointAddress.subscribe(val => {
        localEndpointAddress = val;
        let url: URL;
        try {
            url = new URL(val);
        } catch(error) {
            endpointAddress.set("");
            return;
        }
        url.pathname = '/api/v1/schema';
        fetch(`${url.href}`).then(response => response.json().then(data => matchSelectorSchema = data.filter((elem: SelectorSchema) => elem.name !== FIELD_MATCH_PRIORITY)));
        url.pathname = '/api/v1/selectors';
        endpointPromise = fetch(`${url.href}`).then(response => response.json());
    });

    onDestroy(endpointAddrSub);
</script>

<button type="button" on:click={() => (endpointAddress.set(""))} class="px-3 py-2 text-xs font-medium text-center text-indigo-700 border border-indigo-700 hover:bg-indigo-700 hover:text-white focus:ring-4 focus:outline-none focus:ring-indigo-300 font-medium rounded-lg inline-flex items-center dark:border-indigo-500 dark:text-indigo-500 dark:hover:text-white dark:focus:ring-indigo-800 dark:hover:bg-indigo-500">
    Go Back
</button>
{#await endpointPromise }
    wait
{:then data }
<div class="relative overflow-x-auto mt-4">
    <h2 class="mb-4 text-3xl font-bold dark:text-white">Active match selector entries</h2>
    <button type="button" on:click={() => goto(`/config/create/selector`)} class="mb-4 text-white bg-indigo-500 hover:bg-indigo-800 focus:ring-4 focus:outline-none focus:ring-indigo-300 font-medium rounded-lg text-sm p-2.5 text-center inline-flex items-center mr-2 dark:bg-indigo-600 dark:hover:bg-indigo-500 dark:focus:ring-indigo-800">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-1">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        Create new rule
    <span class="sr-only">Create new rule</span>
    </button>
    <table class="w-full text-sm text-left text-gray-500 dark:text-gray-400">
        <thead class="text-xs text-gray-700 bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
            <tr>
                {#each matchSelectorSchema as entry (entry.id)}
                {#if entry.matchType == MATCH_TYPE_LPM }
                <th scope="col" class="px-6 py-3 border-l dark:border-gray-700">{entry.name} - ({entry.matchType})</th>
                <th scope="col" class="px-6 py-3 border-r dark:border-gray-700">Prefix Length</th>
                {:else if entry.matchType == MATCH_TYPE_TERNARY }
                <th scope="col" class="px-6 py-3 border-l dark:border-gray-700">{entry.name} - ({entry.matchType})</th>
                <th scope="col" class="px-6 py-3 border-r dark:border-gray-700">Mask</th>
                {:else}
                <th scope="col" class="px-6 py-3">{entry.name}- ({entry.matchType})</th>
                {/if}
                {/each}
                <th scope="col" class="px-6 py-3">Actions</th>
            </tr>
        </thead>
        <tbody>
            {#each data as entry }
            <tr class="bg-white border-b dark:bg-gray-800 dark:border-gray-700"> 
                {#each entry.keys.filter(elem => matchSelectorSchema.find(key => key.id == elem.fieldId)) as selectorKey (selectorKey.fieldId) }
                {@const schemaItem = matchSelectorSchema.find(elem => selectorKey.fieldId == elem.id)}
                {#if schemaItem?.matchType == MATCH_TYPE_LPM}
                <td class="px-6 py-4 border-l dark:border-gray-700">0x{selectorKey.value}</td>
                <td class="px-6 py-4 border-r dark:border-gray-700">{selectorKey.prefixLength}</td>
                {:else if schemaItem?.matchType == MATCH_TYPE_TERNARY}
                <td class="px-6 py-4 border-l dark:border-gray-700">0x{selectorKey.value}</td>
                <td class="px-6 py-4 border-r dark:border-gray-700">0x{selectorKey.valueMask}</td>
                {:else}
                <td class="px-6 py-4">0x{selectorKey.value}</td>
                {/if}
                {/each}
                <td class="px-6 py-4">Delete</td>
            </tr>
            {/each}
        </tbody>
    </table>
</div>
{/await}
