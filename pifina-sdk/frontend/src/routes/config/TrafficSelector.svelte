<!--
 Copyright (c) 2023 Thushjandan Ponnudurai
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import { onDestroy } from "svelte";
	import { endpointConfigAddressStore } from "../../lib/stores/endpointConfigStore";
	import { FIELD_MATCH_PRIORITY, MATCH_TYPE_LPM, MATCH_TYPE_TERNARY, type SelectorSchema } from "$lib/models/selectorSchema";
	import type { SelectorEntry } from "$lib/models/selectorEntry";
	import { goto } from "$app/navigation";
	import Modal from "$lib/components/Modal.svelte";

    let localEndpointAddress: string;
    let endpointPromise = Promise.resolve<SelectorEntry[]>([]);
    let matchSelectorSchema: SelectorSchema[] = [];
    let loading = false;
    let showModal = false;
    let closeModal: (() => void);
    let targetRuleToDelete: SelectorEntry;

    const endpointAddrSub = endpointConfigAddressStore.subscribe(val => {
        if (val === "") {
            return
        }
        localEndpointAddress = val;
        fetch(`/api/v1/schema?endpoint=${localEndpointAddress}`).then(response => response.json().then(data => matchSelectorSchema = data.filter((elem: SelectorSchema) => elem.name !== FIELD_MATCH_PRIORITY)));
        fetchEntries();
    });

    function fetchEntries() {
        endpointPromise = fetch(`/api/v1/selectors?endpoint=${localEndpointAddress}`).then(response => response.json());
    }

    function showConfirmModal(entry: SelectorEntry) {
        targetRuleToDelete = entry;
        showModal = true;
    }

    function deleteRule() {
        if (typeof targetRuleToDelete !== 'undefined') {
            loading = true;
            fetch(`/api/v1/selectors?endpoint=${localEndpointAddress}`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(targetRuleToDelete)
            }).then(data => {
                loading = false;
                showModal = false;
                closeModal();
                fetchEntries();
            }).catch(error => { 
                loading = false;
            });
        }
    }

    onDestroy(endpointAddrSub);
</script>

<button type="button" on:click={() => (endpointConfigAddressStore.set(""))} class="px-3 py-2 text-xs font-medium text-center text-indigo-700 border border-indigo-700 hover:bg-indigo-700 hover:text-white focus:ring-4 focus:outline-none focus:ring-indigo-300 font-medium rounded-lg inline-flex items-center dark:border-indigo-500 dark:text-indigo-500 dark:hover:text-white dark:focus:ring-indigo-800 dark:hover:bg-indigo-500">
    Go Back
</button>
{#await endpointPromise }
    wait
{:then data }
<div class="relative overflow-x-auto mt-4">
    <h2 class="mb-4 text-3xl font-bold dark:text-white">Active match selector entries</h2>
    <button type="button" on:click={() => goto(`/config/create/selector`)} class="mb-4 text-white bg-indigo-500 hover:bg-indigo-800 font-medium rounded-lg text-sm p-2.5 text-center inline-flex items-center">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-1">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        Create new rule
    <span class="sr-only">Create new rule</span>
    </button>
    <div class="grid lg:grid-cols-3 md:grid-cols-2 sm:grid-cols-1">
        {#each data as entry }
        <div class="sm:col-span-1 mx-4 my-4">
            <div class="max-w-sm p-6 bg-white border border-gray-200 rounded-lg shadow dark:bg-gray-800 dark:border-gray-700">
                <div class="flex items-center justify-between mb-4">
                    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Session {entry.sessionId}</h5>
                    <button type="button" title="Delete rule" on:click={() => showConfirmModal(entry)} class="inline-flex items-center px-4 py-3 text-sm font-medium text-center text-white bg-red-600 rounded-lg hover:bg-red-800 focus:ring-4 focus:outline-none focus:ring-red-300">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                        </svg>
                    </button>
                </div>
                <div class="grid grid-cols-2">
                {#each entry.keys.filter(elem => matchSelectorSchema.find(key => key.id == elem.fieldId)) as selectorKey (selectorKey.fieldId) }
                {@const schemaItem = matchSelectorSchema.find(elem => selectorKey.fieldId == elem.id)}
                {#if schemaItem?.matchType == MATCH_TYPE_LPM}
                <div class="sm:col-span-2 my-1">
                    <p class="text-lg font-normal text-gray-700 dark:text-gray-400"></p><h6>{schemaItem.name} ({schemaItem.matchType})</h6>
                </div>
                <div class="sm:col-span-1 mb-1">
                    <p class="font-normal text-gray-700 dark:text-gray-400">Value:</p>
                </div>
                <div class="sm:col-span-1 mb-1">
                    <p class="tracking-wider">0x{selectorKey.value}</p>
                </div>
                <div class="sm:col-span-1 mb-2">
                    <p class="font-normal text-gray-700 dark:text-gray-400">Prefix length:</p>
                </div>
                <div class="sm:col-span-1 mb-2">
                    <p class="tracking-wider">{selectorKey.prefixLength}</p>
                </div>
                {:else if schemaItem?.matchType == MATCH_TYPE_TERNARY}
                <div class="sm:col-span-2 mb-1">
                    <p class="text-lg font-normal text-gray-700 dark:text-gray-400">{schemaItem.name} ({schemaItem.matchType})</p>
                </div>
                <div class="sm:col-span-1 mb-1">
                    <p class="font-normal text-gray-700 dark:text-gray-400">Value:</p>
                </div>
                <div class="sm:col-span-1 mb-1">
                    <p class="tracking-wider">0x{selectorKey.value}</p>
                </div>
                <div class="sm:col-span-1 mb-2">
                    <p class="font-normal text-gray-700 dark:text-gray-400">Mask:</p>
                </div>
                <div class="sm:col-span-1 mb-2">
                    <p class="tracking-wider">0x{selectorKey.valueMask}</p>
                </div>
                {:else}
                <div class="sm:col-span-2 mb-1">
                    <p class="text-lg font-normal text-gray-700 dark:text-gray-400">{schemaItem?.name} ({schemaItem?.matchType})</p>
                </div>
                <div class="sm:col-span-1 mb-2">
                    <p class="font-normal text-gray-700 dark:text-gray-400">Value: </p>
                </div>
                <div class="sm:col-span-1 mb-2">
                    <p class="tracking-wider">0x{selectorKey.value}</p>
                </div>
                {/if}
                {/each}
                </div>
            </div>
        </div>
        {/each}
    </div>
</div>

<Modal bind:showModal={showModal} bind:closeModal={closeModal}>
    <svg aria-hidden="true" class="mx-auto mb-4 text-gray-400 w-14 h-14 dark:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">Are you sure you want to delete this rule?</h3>
    <button type="button" on:click={() => deleteRule()} class="text-white bg-red-600 hover:bg-red-800 focus:ring-4 focus:outline-none focus:ring-red-300 dark:focus:ring-red-800 font-medium rounded-lg text-sm inline-flex items-center px-5 py-2.5 text-center mr-2" disabled={loading}>
        {#if loading }
        <svg aria-hidden="true" role="status" class="inline w-4 h-3 mr-2 text-gray-200 animate-spin dark:text-gray-600" viewBox="0 0 100 101" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z" fill="currentColor"/>
            <path d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z" fill="#1C64F2"/>
        </svg>
        Deleting...
        {:else}
        Delete
        {/if}
    </button>
</Modal>
{:catch error}
<div class="relative overflow-x-auto mt-8">
    <div class="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400" role="alert">
        <span class="font-medium">Controller unreachable</span> Cannot retrieve traffic selector information from controller.
    </div>
</div>
{/await}
