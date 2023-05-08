<script>

	import { writable } from 'svelte/store';	
	
	const messages = writable([]);
	const evtSource = new EventSource("https://localhost:8655/events?stream=metrics");
	evtSource.onmessage = function(event) {
		console.log(event);
		var dataobj = JSON.parse(event.data);
		messages.update(arr => arr.concat(dataobj));
	}
	
	
</script>

<header class="bg-white shadow">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <h1 class="text-3xl font-bold tracking-tight text-gray-900">Dashboard</h1>
    </div>
</header>
<main>
    <div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
        <div class="grid grid-cols-4">
            <div>
                <select placeholder="Placeholder" class="px-3 py-3 placeholder-slate-300 text-slate-600 relative bg-white bg-white rounded text-sm border-0 shadow outline-none focus:outline-none focus:ring w-full">
                    <option value="volvo">Volvo</option>
                    <option value="saab">Saab</option>
                    <option value="mercedes">Mercedes</option>
                    <option value="audi">Audi</option>
                </select>
            </div>
        </div>
    </div>
</main>
