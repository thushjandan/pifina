<script lang="ts">
	import { onDestroy } from "svelte";
	import { endpointAddress } from "../../EndpointStore";
	import { FIELD_MATCH_PRIORITY, MATCH_TYPE_LPM, MATCH_TYPE_TERNARY, type SelectorSchema } from "$lib/models/SelectorSchema";
	import { filter, selector } from "d3";
	import { goto } from "$app/navigation";

    let schemaPromise = Promise.resolve<SelectorSchema[]>([]);

    const endpointAddrSub = endpointAddress.subscribe(val => {
        let url: URL;
        try {
            url = new URL(val);
        } catch(error) {
            goto(`/config`);
            return;
        }
        url.pathname = '/api/v1/schema';
        schemaPromise = fetch(`${url.href}`).then(response => response.json());
    });

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
            <form>
                    {#each data.filter(elem => elem.name !== FIELD_MATCH_PRIORITY) as selectorKey (selectorKey.id) }
                    {#if selectorKey.matchType === MATCH_TYPE_LPM}
                    <div class="grid gap-6 mb-6 md:grid-cols-2">
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">{selectorKey.name} - ({selectorKey.matchType})</label>
                            <input type="text" id={`key${selectorKey.id}`} class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="0x0a000101" required>
                        </div>
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Prefix Length <span class="italic">(numerical)</span></label>
                            <input type="text" id="first_name" class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="24" required>
                        </div>
                    </div>
                    {:else if selectorKey.matchType === MATCH_TYPE_TERNARY}
                    <div class="grid gap-6 mb-6 md:grid-cols-2">
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">{selectorKey.name} ({selectorKey.matchType})</label>
                            <input type="text" id={`key${selectorKey.id}`} class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="0x0a000101" required>
                        </div>
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Mask</label>
                            <input type="text" id="first_name" class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="0xffffffff" required>
                        </div>
                    </div>
                    {:else}
                    <div class="mb-6">
                        <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">{selectorKey.name} ({selectorKey.matchType})</label>
                        <input type="text" id={`key${selectorKey.id}`} class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="0x0a000101" required>
                    </div>
                    {/if}
                    {/each}
                    
                
                <button type="button" class="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">Create</button>
            </form>
            {/await}
            
        </div>
    </div>
</main>