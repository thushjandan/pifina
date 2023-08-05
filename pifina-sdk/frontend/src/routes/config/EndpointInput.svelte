<script lang="ts">
	import { EndpointType, type EndpointModel } from "$lib/models/EndpointModel";
	import { endpointAddress } from "./EndpointStore";

    const fetchEndpoints = () => {
        return fetch('/api/v1/endpoints').then(
            response => response.json().then(
                endpoints => new Promise<EndpointModel[]>(resolve => resolve(endpoints.filter((item: EndpointModel) => item.type == EndpointType.HOSTTYPE_TOFINO)))
            )
        )
    }

    let endpointPromise = fetchEndpoints();
    let editEndpointEntity: EndpointModel;
    let editModeEnabled: boolean = false;
    let loading: boolean = false;
    let editErrorMsg = "";

    function saveEndpoint(endpoint: string) {
        endpointAddress.set(endpoint);
    }

    function editEndpoint(endpoint: EndpointModel) {
        editEndpointEntity = {
            name: endpoint.name,
            type: endpoint.type,
            groupId: endpoint.groupId,
            address: endpoint.address,
            port: endpoint.port
        }
        editModeEnabled = true;
    }

    function submitEditEndpoint() {
        editModeEnabled = false;
        loading = true;

        fetch(`/api/v1/endpoints`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(editEndpointEntity)
        }).then(data => {
            loading = false;
            if (data.ok) {
                endpointPromise = fetchEndpoints();
            } else {
                data.json().then(data => editErrorMsg = data.message);
            }
        }).catch(error => { 
            loading = false;
            editErrorMsg = error;
        });
    }
</script>


<div class="grid grid-cols-3">
    {#await endpointPromise }
        <p>Loading endpoints...</p>
    {:then data}
    {#if !editModeEnabled}
    {#each data as endpoint}
    <div class="sm:col-span-1 mx-4 my-4">
        <div class="max-w-sm p-6 bg-white border border-gray-200 rounded-lg shadow dark:bg-gray-800 dark:border-gray-700">
            <a href={undefined} on:click={() => saveEndpoint(endpoint.name)}>
                <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">{endpoint.name}</h5>
            </a>
            <p class="mb-3 font-normal text-gray-700 dark:text-gray-400">Connection: {endpoint.address}:{endpoint.port}<button on:click={() => editEndpoint(endpoint)} class="text-white bg-orange-700 hover:bg-orange-800 focus:ring-4 focus:outline-none focus:ring-orange-300 font-medium rounded-lg text-sm p-2.5 text-center inline-flex items-center ml-2">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                </svg>
            </button></p>
            <button type="button" on:click={() => saveEndpoint(endpoint.name)} class="inline-flex items-center px-4 py-3 text-sm font-medium text-center text-white bg-blue-700 rounded-lg hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
                Select
                <svg class="w-3.5 h-3.5 ml-2" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
                    <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M1 5h12m0 0L9 1m4 4L9 9"/>
                </svg>
            </button>
        </div>
    </div>
    {/each}
    {/if}
    {:catch error}
        <p>Loading endpoints failed! Retry later. {error}</p>                
    {/await }
    {#if editModeEnabled}
    <div class="max-w-sm p-6 bg-white border border-gray-200 rounded-lg shadow dark:bg-gray-800 dark:border-gray-700">
        {#if editErrorMsg !== ""}
        <div class="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400" role="alert">
            <span class="font-medium">Update failed</span> {editErrorMsg}
        </div>
        {/if}
        <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Edit {editEndpointEntity?.name}</h5>
        <div class="grid gap-6 mb-6 md:grid-cols-2">
            <div>
                <label for="address" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">FQDN or IP</label>
                <input type="text" id="address" bind:value={editEndpointEntity.address} class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="Doe" required>
            </div>
            <div>
                <label for="port" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Controller API Port</label>
                <input type="text" id="port" bind:value={editEndpointEntity.port} class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="Flowbite" required>
            </div>  
        </div>
        <button type="button" on:click={submitEditEndpoint} disabled={loading} class="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
            {#if loading}
            Updating...
            {:else}
            Update
            {/if}
        </button>
        <button type="button" on:click={() => editModeEnabled = false} class="py-2.5 px-5 mr-2 mb-2 text-sm font-medium text-gray-900 focus:outline-none bg-white rounded-lg border border-gray-200 hover:bg-gray-100 hover:text-blue-700 focus:z-10 focus:ring-4 focus:ring-gray-200 dark:focus:ring-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 dark:hover:text-white dark:hover:bg-gray-700">Cancel</button>
    </div>
    {/if}
</div>