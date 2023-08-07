<!--
 Copyright (c) 2023 Thushjandan Ponnudurai
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script>

	import Dashboard from './Dashboard.svelte';
    const endpointPromise = fetch('/api/v1/endpoints').then(response => response.json());
	
</script>

<header class="bg-white shadow">
    <div class="mx-auto max-w-screen-2xl px-4 py-6 sm:px-6 lg:px-8">
      <h1 class="text-3xl font-bold tracking-tight text-gray-900">Dashboard</h1>
    </div>
</header>
<main>
    <div class="mx-auto max-w-screen-2xl py-6 sm:px-6 lg:px-8">
        <div class="bg-white rounded-lg px-8 py-8 shadow-lg">
            {#await endpointPromise }
                <p class="mb-2 text-lg text-gray-500 md:text-xl dark:text-gray-400">Loading endpoints...</p>
            {:then data }
                {#if data.length > 0}
                <Dashboard endpoints={data}></Dashboard>
                {:else}
                <p class="mb-2 text-lg text-gray-500 md:text-xl dark:text-gray-400">Currently no probes are sending any metrics. Please refresh page.</p>
                {/if}
            {:catch error}
                <p>Loading endpoints failed! Retry later. {error}</p>                
            {/await }
        </div>
    </div>
</main>
