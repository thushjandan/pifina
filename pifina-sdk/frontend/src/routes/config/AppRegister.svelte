<!--
 Copyright (c) 2023 Thushjandan Ponnudurai
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import Modal from "$lib/components/Modal.svelte";
	import { onDestroy } from "svelte";
	import { endpointAddress } from "./EndpointStore";
	import { goto } from "$app/navigation";
	import type { AppRegisterModel } from "$lib/models/AppRegister";

    let localEndpointAddress: string;
    let appRegPromise: Promise<AppRegisterModel[]>;
    let showModal = false;
    let targetAppRegisterToDelete: AppRegisterModel;
    let loading = false;
    let closeModal: (() => void);

    const endpointAddrSub = endpointAddress.subscribe(val => {
        if (val === "") {
            return
        }
        localEndpointAddress = val;
        fetchEntries();
    });

    function fetchEntries() {
        appRegPromise = fetch(`/api/v1/app-registers?endpoint=${localEndpointAddress}`).then(response => response.json());
    }

    function showConfirmModal(entry: AppRegisterModel) {
        targetAppRegisterToDelete = entry;
        showModal = true;
    }

    function deleteProbe() {
        if (typeof targetAppRegisterToDelete !== 'undefined') {
            loading = true;
            let url = localEndpointAddress;
            fetch(`/api/v1/app-registers?endpoint=${localEndpointAddress}`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(targetAppRegisterToDelete)
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


{#await appRegPromise }
    wait
{:then data }
<div class="relative overflow-x-auto mt-8">
    <h2 class="mb-4 text-3xl font-bold dark:text-white">Probes for user application owned register</h2>
    <button type="button" on:click={() => goto(`/config/create/app-register`)} class="mb-4 text-white bg-indigo-500 hover:bg-indigo-800 font-medium rounded-lg text-sm p-2.5 text-center inline-flex items-center">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-1">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        Add new register
    <span class="sr-only">Add new register</span>
    </button>
    <table class="w-full text-sm text-left text-gray-500 dark:text-gray-400">
        <thead class="text-xs text-gray-700 bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
            <tr>
                <th scope="col" class="px-6 py-3">Register Name</th>
                <th scope="col" class="px-6 py-3">Register Index</th>
                <th scope="col" class="px-6 py-3">Actions</th>
            </tr>
        </thead>
        <tbody>
            {#each data as entry }
            <tr class="bg-white border-b dark:bg-gray-800 dark:border-gray-700"> 
                <td class="px-6 py-4">{entry.name}</td>
                <td class="px-6 py-4">{entry.index}</td>
                <td class="px-6 py-4">
                    <button type="button" on:click={() => showConfirmModal(entry)} class="text-white text-center bg-red-600 hover:bg-red-800 font-medium rounded-lg text-sm w-full sm:w-auto px-2 py-1.5 text-center">
                        Delete
                    </button>
                </td>
            </tr>
            {/each}
        </tbody>
    </table>
</div>

<Modal bind:showModal={showModal} bind:closeModal={closeModal}>
    <svg aria-hidden="true" class="mx-auto mb-4 text-gray-400 w-14 h-14 dark:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">Are you sure you want to delete this rule?</h3>
    <button type="button" on:click={() => deleteProbe()} class="text-white bg-red-600 hover:bg-red-800 focus:ring-4 focus:outline-none focus:ring-red-300 dark:focus:ring-red-800 font-medium rounded-lg text-sm inline-flex items-center px-5 py-2.5 text-center mr-2" disabled={loading}>
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
        <span class="font-medium">Controller unreachable</span> Cannot retrieve app register information from controller.
    </div>
</div>
{/await}