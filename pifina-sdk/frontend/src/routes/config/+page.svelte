<!--
 Copyright (c) 2023 Thushjandan Ponnudurai
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import EndpointInput from "./EndpointInput.svelte";
    import { endpointAddress } from "./EndpointStore";
	import { onDestroy } from 'svelte';
	import TrafficSelector from "./TrafficSelector.svelte";
	import AppRegister from "./AppRegister.svelte";
	import DevPortList from "./DevPortList.svelte";

    let localEndpointAddress: string = "";

    const unsubscribeEndpointAddr = endpointAddress.subscribe(value => {
        localEndpointAddress = value;
    });

    onDestroy(unsubscribeEndpointAddr);
</script>
<header class="bg-white shadow">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <h1 class="text-3xl font-bold tracking-tight text-gray-900">Configuration {localEndpointAddress}</h1>
    </div>
</header>
<main>
    <div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
        <div class="bg-white rounded-lg px-8 py-8 shadow-lg">
            {#if localEndpointAddress == "" }
                <EndpointInput />
            {:else }
                <TrafficSelector />
                <AppRegister />
                <DevPortList />
            {/if}
        </div>
    </div>
</main>
