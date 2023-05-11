<script>

	import Dashboard from './Dashboard.svelte';
    const endpointPromise = fetch('/api/v1/endpoints').then(response => response.json());
	
</script>

<header class="bg-white shadow">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <h1 class="text-3xl font-bold tracking-tight text-gray-900">Dashboard</h1>
    </div>
</header>
<main>
    <div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
        <div class="bg-white rounded-lg px-8 py-8 shadow-lg">
            {#await endpointPromise }
                <p>Loading endpoints...</p>
            {:then data }
                <Dashboard endpoints={data}></Dashboard>
            {:catch error}
                <p>Loading endpoints failed! Retry later. {error}</p>                
            {/await }
        </div>
    </div>
</main>
