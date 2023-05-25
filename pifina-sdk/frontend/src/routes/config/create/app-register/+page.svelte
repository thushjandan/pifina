<script lang="ts">
	import type { AppRegisterModel } from "$lib/models/AppRegister";
	import { onDestroy } from "svelte";
	import { endpointAddress } from "../../EndpointStore";
	import { goto } from "$app/navigation";
	import Modal from "$lib/components/Modal.svelte";
	import { fade } from "svelte/transition";

    let localEndpointAddress: URL;
    let newEntry: AppRegisterModel = {index: 0, name: ""} as AppRegisterModel;
    let availableRegPromise: Promise<string[]>;
    let loading= false;
    let createDone = false;
    let createErrorMsg = "";
    let showModal = false;
    let closeModal: (() => void);


    const endpointAddrSub = endpointAddress.subscribe(val => {
        let url: URL;
        try {
            url = new URL(val);
        } catch(error) {
            goto(`/config`);
            return;
        }
        localEndpointAddress = url;
        url.pathname = '/api/v1/app-registers/available';
        availableRegPromise = fetch(`${url.href}`).then(response => response.json());
    });

    function handleSubmit() {
        loading = true;
        createErrorMsg = "";
        localEndpointAddress.pathname = '/api/v1/app-registers'
        fetch(localEndpointAddress.href, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(newEntry)
        }).then(data => {
            loading = false;
            if (data.ok) {
                createDone = true;
                setInterval(() => goto('/config'), 500);
            } else {
                data.json().then(data => createErrorMsg = data.message);
            }
        }).catch(error => { 
            loading = false;
            createErrorMsg = error;
        });

    }

    onDestroy(endpointAddrSub);
</script>

<header class="bg-white shadow">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <h1 class="text-3xl font-bold tracking-tight text-gray-900">Add new probe</h1>
    </div>
</header>
<main>
    <div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
        <div class="bg-white rounded-lg px-8 py-8 shadow-lg">
            {#if createErrorMsg !== "" }
            <div class="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400" role="alert">
                <span class="font-medium">Error occured</span> {createErrorMsg}
              </div>
            {/if}
            <p class="mb-2 text-lg text-gray-500 md:text-xl dark:text-gray-400">Enter the exact full name of the your register and its index, which you would like to monitor with PIFINA.</p>
            <form on:submit|preventDefault={handleSubmit}>
                    <div class="grid gap-6 mb-6 md:grid-cols-2">
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Register's full name</label>
                            <div class="flex">
                                <input type="text" bind:value={newEntry.name} class="rounded-lg bg-gray-50 border text-gray-900 focus:ring-indigo-500 focus:border-indigo-500 block flex-1 min-w-0 w-full text-sm border-gray-300 p-2.5  dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-500 dark:focus:border-indigo-500" placeholder="pipe.SwitchIngress.MyApp.MyRegister" required>
                            </div>
                        </div>
                        <div>
                            <label for="first_name" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Register index</label>
                            <input type="number" bind:value={newEntry.index} class="rounded-lg bg-gray-50 border text-gray-900 focus:ring-indigo-500 focus:border-indigo-500 block flex-1 min-w-0 w-full text-sm border-gray-300 p-2.5  dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-indigo-500 dark:focus:border-indigo-500" placeholder="2" required>
                        </div>
                    </div>
                <button type="submit" class:bg-indigo-600="{!createDone}" class:bg-green-600="{createDone}" class="text-white text-center hover:bg-indigo-800 font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 text-center disabled:bg-indigo-300 mr-2" disabled={loading}>
                    {#if loading && !createDone }
                    <svg aria-hidden="true" role="status" class="inline w-4 h-3 mr-2 text-gray-200 animate-spin dark:text-gray-600" viewBox="0 0 100 101" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z" fill="currentColor"/>
                        <path d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z" fill="#1C64F2"/>
                    </svg>
                    Creating...
                    {:else if !loading && createDone }
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-5 h-5 mr-1">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                    </svg>                      
                    Created!
                    {:else }
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-5 h-5 mr-1">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                    </svg>
                    Create
                    {/if}
                </button>
                <button on:click={() => showModal = !showModal} type="button" class="py-2.5 px-5 mr-2 text-sm font-medium text-gray-900 bg-white rounded-lg border border-gray-200 hover:bg-gray-100 hover:text-indigo-700">Show all available registers</button>
                <button type="button" on:click={() => goto('/config')} class="py-2.5 px-5 mr-2 mb-2 text-sm font-medium text-gray-900 bg-white rounded-lg border border-gray-200 hover:bg-gray-100 hover:text-indigo-700">
                    Cancel
                </button>
            </form>
            {#if showModal}
            <div transition:fade class="mt-6">
            {#await availableRegPromise then data }
                <h2 class="mb-4 text-3xl font-bold dark:text-white">Name of all available registers</h2>
                <ul class="list-disc list-inside text-gray-500 dark:text-gray-400">
                {#each data as entry}
                    <li><a href={undefined} on:click|preventDefault={() => newEntry.name = entry}>{entry}</a></li>
                {/each}
                </ul>
            {/await}     
            </div>
            {/if}
        </div>
    </div>
</main>