<script lang="ts">
	import { onDestroy } from "svelte";
	import { endpointAddress } from "../../EndpointStore";
	import { FIELD_MATCH_PRIORITY, MATCH_TYPE_LPM, MATCH_TYPE_TERNARY, type SelectorSchema } from "$lib/models/SelectorSchema";
	import { filter, selector } from "d3";
	import { goto } from "$app/navigation";
	import type { SelectorEntry } from "$lib/models/SelectorEntry";

    let schemaPromise = Promise.resolve<SelectorSchema[]>([]);
    let localEndpointAddress: URL;
    let newEntry: SelectorEntry = {sessionId: 0, keys: []} as SelectorEntry;
    const hexRegex = /[0-9A-Fa-f]+/;

    const endpointAddrSub = endpointAddress.subscribe(val => {
        let url: URL;
        try {
            url = new URL(val);
        } catch(error) {
            goto(`/config`);
            return;
        }
        localEndpointAddress = url;
        url.pathname = '/api/v1/schema';
        schemaPromise = fetch(`${url.href}`).then(response => response.json().then((data: SelectorSchema[]) => {
            data = data.filter(elem => elem.name !== FIELD_MATCH_PRIORITY);
            newEntry.keys = [];
            data.forEach(item => {
                newEntry.keys.push({fieldId: item.id, matchType: item.matchType, value: ""})
            });
            return Promise.resolve<SelectorSchema[]>(data);
        }));
    });

    function handleSubmit(e: Event) {
        console.log(newEntry);
        for (const item of newEntry.keys) {
            if (item.matchType === MATCH_TYPE_TERNARY) {
                if (!item.valueMask?.match(hexRegex)) {
                    return
                }    
            }
            if (!item.value.match(hexRegex)) {
                return
            }
        }
        localEndpointAddress.pathname = '/api/v1/selectors'
        fetch(localEndpointAddress.href, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(newEntry)
        }).then(data => console.log(data)).catch(error => console.log(error));
    }

    onDestroy(endpointAddrSub);

</script>

<header class="bg-white shadow">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <h1 class="text-3xl font-bold tracking-tight text-gray-900">Create new rule</h1>
    </div>
</header>
<main>
    <div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
        <div class="bg-white rounded-lg px-8 py-8 shadow-lg">
            <p class="mb-6 text-lg text-gray-500 md:text-xl dark:text-gray-400">Enter all the values as Hexadecimal strings if not specified. (e.g. 0x00CAFE00)</p>
            {#await schemaPromise }
                waiting.default.
            {:then data } 
            <form on:submit|preventDefault={handleSubmit}>
                    {#each data as selectorKey, i (selectorKey.id) }
                    {#if selectorKey.matchType === MATCH_TYPE_LPM}
                    <div class="grid gap-6 mb-6 md:grid-cols-2">
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">{selectorKey.name} - ({selectorKey.matchType})</label>
                            <div class="flex">
                                <span class="inline-flex items-center px-3 text-sm text-gray-900 bg-gray-200 border border-r-0 border-gray-300 rounded-l-md dark:bg-gray-600 dark:text-gray-400 dark:border-gray-600">
                                    0x
                                </span>
                                <input type="text" bind:value={newEntry.keys[i].value} class="rounded-none rounded-r-lg bg-gray-50 border text-gray-900 focus:ring-indigo-500 focus:border-indigo-500 block flex-1 min-w-0 w-full text-sm border-gray-300 p-2.5  dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-500 dark:focus:border-indigo-500" placeholder="0a000101" required>
                            </div>
                        </div>
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Prefix Length <span class="italic">(numerical)</span></label>
                            <input type="number" bind:value={newEntry.keys[i].prefixLength} class="rounded-none rounded-r-lg bg-gray-50 border text-gray-900 focus:ring-indigo-500 focus:border-indigo-500 block flex-1 min-w-0 w-full text-sm border-gray-300 p-2.5  dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-500 dark:focus:border-indigo-500" placeholder="24" required>
                        </div>
                    </div>
                    {:else if selectorKey.matchType === MATCH_TYPE_TERNARY}
                    <div class="grid gap-6 mb-6 md:grid-cols-2">
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">{selectorKey.name} ({selectorKey.matchType})</label>
                            <div class="flex">
                                <span class="inline-flex items-center px-3 text-sm text-gray-900 bg-gray-200 border border-r-0 border-gray-300 rounded-l-md dark:bg-gray-600 dark:text-gray-400 dark:border-gray-600">
                                    0x
                                </span>
                                <input type="text" bind:value={newEntry.keys[i].value} class="rounded-none rounded-r-lg bg-gray-50 border text-gray-900 focus:ring-indigo-500 focus:border-indigo-500 block flex-1 min-w-0 w-full text-sm border-gray-300 p-2.5  dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-500 dark:focus:border-indigo-500" placeholder="0a000101" required>
                            </div>
                        </div>
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Mask</label>
                            <div class="flex">
                                <span class="inline-flex items-center px-3 text-sm text-gray-900 bg-gray-200 border border-r-0 border-gray-300 rounded-l-md dark:bg-gray-600 dark:text-gray-400 dark:border-gray-600">
                                    0x
                                </span>
                                <input type="text" bind:value={newEntry.keys[i].valueMask} class="rounded-none rounded-r-lg bg-gray-50 border text-gray-900 focus:ring-indigo-500 focus:border-indigo-500 block flex-1 min-w-0 w-full text-sm border-gray-300 p-2.5  dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-500 dark:focus:border-indigo-500" placeholder="ffffffff" required>
                            </div>
                        </div>
                    </div>
                    {:else}
                    <div class="mb-6">
                        <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">{selectorKey.name} ({selectorKey.matchType})</label>
                        <div class="flex">
                            <span class="inline-flex items-center px-3 text-sm text-gray-900 bg-gray-200 border border-r-0 border-gray-300 rounded-l-md dark:bg-gray-600 dark:text-gray-400 dark:border-gray-600">
                                0x
                            </span>
                            <input type="text" bind:value={newEntry.keys[i].value} class="rounded-none rounded-r-lg bg-gray-50 border text-gray-900 focus:ring-indigo-500 focus:border-indigo-500 block flex-1 min-w-0 w-full text-sm border-gray-300 p-2.5  dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-500 dark:focus:border-indigo-500" placeholder="0a000101" required>
                        </div>
                    </div>
                    {/if}
                    {/each}
                <button type="submit" class="text-white bg-indigo-700 hover:bg-indigo-800 focus:ring-4 focus:outline-none focus:ring-indigo-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-indigo-600 dark:hover:bg-indigo-700 dark:focus:ring-indigo-800">Create</button>
            </form>
            {/await}
            
        </div>
    </div>
</main>