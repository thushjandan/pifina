<script>
	import { endpointAddress } from "./EndpointStore";

    let endpointAddressForm = "";
    function saveEndpoint() {
        endpointAddress.set(endpointAddressForm);
    }
</script>

<div class="grid grid-cols-3">
    <div class="sm:col-span-1">
        <legend class="text-sm font-semibold leading-6 text-gray-900">Which monitoring target do you want to reconfigure?</legend>
        <p class="mt-1 text-sm leading-6 text-gray-600">Define a FQDN or IP address.</p>
        <div class="mt-2 mb-4">
            <input type="text" bind:value={endpointAddressForm} name="endpoint" id="endpoint" class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6">        
        </div>
        <button type="button" on:click={saveEndpoint} class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600">Retrieve configuration</button>
    </div>
</div>