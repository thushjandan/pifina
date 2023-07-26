<script lang="ts">
	import Chart from '../lib/components/Chart.svelte';
	import * as Plot from "@observablehq/plot";
	import type { DTOPifinaMetricItem, MetricData } from '../lib/models/MetricItem';
	import { PIFINA_DEFAULT_PROBE_CHART_ORDER, PIFINA_PROBE_CHART_CFG, PROBE_INGRESS_JITTER } from '$lib/models/metricNames';
	import { ChartMenuCategoryModel } from '$lib/models/ChartMenuCategory';
	import TmMetricCharts from '$lib/components/TMMetricCharts.svelte';
	import type { EndpointModel } from '$lib/models/EndpointModel';
	import { PifinaMetricName } from '$lib/models/metricTypes';

    export let endpoints: EndpointModel[];

	let clientFullScreenWidth;
	let clientHalfScreenWidth;
    let selectedEndpoint: string = endpoints[0]?.name || "";
	let sessionIds = new Set<number>();
	let selectedSessionIds: number[] = [];
	let sessionIdFilterIsDirty = false;
	let metricData: MetricData = {};
	let appRegister = new Set<string>();
	let extraProbeNames = new Set<string>();
	let tmMetrics = new Set<string>();
	let selectedChartCategory = ChartMenuCategoryModel.MAIN_CHARTS;
	let isEnabled = true;

	let worker = new SharedWorker(new URL('$lib/sharedworker/sharedworker.ts', import.meta.url), {type: 'module'});

	worker.port.postMessage({status: "CONNECT", endpoint: selectedEndpoint});
	worker.port.onmessage = (event: MessageEvent) => {
		let dataobj: DTOPifinaMetricItem[] = event.data;

		dataobj.forEach(item => {
			let key = `${item.metricName}${item.type}`;
			// Check if it's a metric from a default probe
			if (!item.metricName.startsWith("PF_")) {
				appRegister.add(item.metricName);
				key = item.metricName;
			}
			if (item.metricName.startsWith("PF_TM_")) {
				tmMetrics.add(item.metricName)
				key = item.metricName;
			}
			if (item.metricName.startsWith("PF_EXTRA")) {
				extraProbeNames.add(item.metricName);
				key = item.metricName;
			}
			// check if key exists. If not, create a new list.
			if (!(key in metricData)) {
				metricData[key] = [];
			}
			if (!sessionIds.has(item.sessionId) && !item.metricName.startsWith("PF_TM_")) {
				sessionIds.add(item.sessionId);
			}
			// Convert nano seconds to miliseconds
			if (item.metricName == PifinaMetricName.INGRESS_JITTER_AVG) {
				if (item.value > 0) {
					item.value = item.value / 1000;
				}
			}
			metricData[key].push({timestamp: new Date(item.timestamp), value: item.value, sessionId: item.sessionId, type: item.type});
		})

		// Default initialization of filters
		if (!sessionIdFilterIsDirty && selectedSessionIds.length == 0) {
			for (const sessionId of sessionIds.values()) {
				selectedSessionIds.push(sessionId);
			}
			sessionIds = sessionIds;
		}

		// Limit series length
		for (const mapkey in metricData) {
			let mapKeySessioId = new Set<number>();
			const groupBySessionId = metricData[mapkey].forEach(item => {
				mapKeySessioId.add(item.sessionId);
			});
			const metricSizeLimit = (mapKeySessioId.size + 1) * 20;
			if (metricData[mapkey].length > metricSizeLimit) {
				metricData[mapkey].splice(0, metricData[mapkey].length - metricSizeLimit);
			}
		}

		// Force rerender
		metricData = metricData;
	}


	const onEndpointChange = () => {
		worker.port.postMessage({status: "CONNECT", endpoint: selectedEndpoint});
	}

	const toggleEventStreaming = () => {
		if (isEnabled) {
			// Close previous event source
			worker.port.postMessage({status: "CLOSE", endpoint: selectedEndpoint});
		}
		if (!isEnabled) {
			worker.port.postMessage({status: "CONNECT", endpoint: selectedEndpoint});
		}
		isEnabled = !isEnabled;
	}

	const isMainChartSelected = () => selectedChartCategory == ChartMenuCategoryModel.MAIN_CHARTS;
	const isTMChartSelected = () => selectedChartCategory == ChartMenuCategoryModel.TM_CHARTS;
	const isAppRegChartSelected = () => selectedChartCategory == ChartMenuCategoryModel.APP_REG_CHARTS;
	const isExtraProbesChartSelected = () => selectedChartCategory == ChartMenuCategoryModel.EXTRA_PROBES_CHARTS;

	const openDetailView = (metricName: string) => {
		window.open(`/dashboard/detail?endpoint=${selectedEndpoint}&selectedMetric=${metricName}`, "_blank");
	}

	const xScaleOptions: Plot.ScaleOptions = {
		label: "Timestamp",
		tickSpacing: 150,
		tickFormat: (value, _) => (value.toLocaleString(undefined, {
			hour: 'numeric',
			minute: 'numeric',
			second: 'numeric'
		})),
	};

</script>

<div class="grid grid-cols-4">
	<div class="sm:col-span-1">
		<label for="target" class="block text-sm font-medium leading-6 text-gray-900">Choose a monitoring target:</label>
		<div class="mt-2">
			<select bind:value={selectedEndpoint} on:change={onEndpointChange} name="target" class="px-3 py-3 placeholder-slate-300 text-slate-600 relative bg-white bg-white rounded text-sm border-0 shadow outline-none focus:outline-none focus:ring w-full">
				{#each endpoints as endpoint }
				<option value={endpoint.name}>{endpoint.name}</option>
				{/each}
			</select>
		</div>
	</div>
	<div class="sm:col-span-3 justify-self-end">
		<div class="mt-2 mr-4">
			<button on:click={toggleEventStreaming} class:bg-orange-600={isEnabled} class:hover:bg-orange-800={isEnabled} class:bg-green-600={!isEnabled} class:hover:bg-green-600={!isEnabled} class="text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
				{#if isEnabled}
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5" />
				</svg>				  
				Pause
				{:else}
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z" />
				</svg>
				Start
				{/if}
			</button>
		</div>
	</div>
</div>
{#key selectedChartCategory }
<div class="mt-6 text-sm font-medium text-center text-gray-500 border-b border-gray-200 dark:text-gray-400 dark:border-gray-700">
    <ul class="flex flex-wrap -mb-px">
        <li class="mr-2">
            <a href={null} on:click={() => selectedChartCategory = ChartMenuCategoryModel.MAIN_CHARTS} class="inline-block p-4 border-b-2 rounded-t-lg hover:text-gray-600 hover:border-gray-300" class:border-transparent={!isMainChartSelected()} class:text-indigo-600={isMainChartSelected()} 
				class:border-blue-600={isMainChartSelected()} class:hover:text-gray-600={!isMainChartSelected()} class:hover:border-gray-300={!isMainChartSelected()}>
				Default Probes
			</a>
        </li>
		<li class="mr-2">
            <a href={null} on:click={() => selectedChartCategory = ChartMenuCategoryModel.APP_REG_CHARTS} class="inline-block p-4 border-b-2 rounded-t-lg hover:text-gray-600 hover:border-gray-300" class:border-transparent={!isAppRegChartSelected()} class:text-indigo-600={isAppRegChartSelected()} 
				class:border-blue-600={isAppRegChartSelected()} class:hover:text-gray-600={!isAppRegChartSelected()} class:hover:border-gray-300={!isAppRegChartSelected()}>
				Application owned registers</a>
        </li>
		<li class="mr-2">
            <a href={null} on:click={() => selectedChartCategory = ChartMenuCategoryModel.EXTRA_PROBES_CHARTS} class="inline-block p-4 border-b-2 rounded-t-lg hover:text-gray-600 hover:border-gray-300" class:border-transparent={!isExtraProbesChartSelected()} class:text-indigo-600={isExtraProbesChartSelected()} 
				class:border-blue-600={isExtraProbesChartSelected()} class:hover:text-gray-600={!isExtraProbesChartSelected()} class:hover:border-gray-300={!isExtraProbesChartSelected()}>
				Extra probes</a>
        </li>
        <li class="mr-2">
            <a href={null} on:click={() => selectedChartCategory = ChartMenuCategoryModel.TM_CHARTS} class="inline-block p-4 border-b-2 rounded-t-lg hover:text-gray-600 hover:border-gray-300" class:border-transparent={!isTMChartSelected()} class:text-indigo-600={isTMChartSelected()} 
				class:border-blue-600={isTMChartSelected()} class:hover:text-gray-600={!isTMChartSelected()} class:hover:border-gray-300={!isTMChartSelected()}>
				Traffic Manager</a>
        </li>
    </ul>
</div>
{#if isMainChartSelected() == true }
<div class="divide-y divide-solid">
	<div class="mt-8">
		<div class="sm:col-span-1">
			<label for="sessionIds" class="block text-sm font-medium leading-6 text-gray-900">Filter by session ID:</label>
			<div class="mt-2 flex flex-row">
				{#each [...sessionIds.values()] as sessionId}
						<div class="items-center">
							<input type=checkbox bind:group={selectedSessionIds} on:change={() => sessionIdFilterIsDirty = true} name="sessionIds" value={sessionId} class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600" />
							<label for="comments" class="ml-1 mr-4 font-medium text-gray-900">{sessionId}</label>
						</div>
				{/each}
			</div>
		</div>
	</div>
	{#each PIFINA_DEFAULT_PROBE_CHART_ORDER as probeItem}
	{#if Array.isArray(probeItem)}
	<div class="mt-8 grid md:grid-cols-2">
		{#each probeItem as subProbeItem }
		{#if subProbeItem in metricData }
		<div bind:clientWidth={clientHalfScreenWidth} class="mt-8 pt-4">
			<div class="grid grid-cols-4">
				<div class="sm:col-span-3">
					<h2>{PIFINA_PROBE_CHART_CFG[subProbeItem]['title']}</h2>
				</div>
				<div class="sm:col-span-1 justify-self-end pr-4">
					<button on:click={() => openDetailView(subProbeItem)} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
						</svg>					  
						Open
					</button>
				</div>
			</div>
			<Chart options={{
				x: xScaleOptions,
				y: {
					label: PIFINA_PROBE_CHART_CFG[subProbeItem]['yAxisName'],
					grid: true
				},
				width: clientHalfScreenWidth,
				color: {legend: true, type: "categorical"},
				marks: [
					Plot.line(metricData[subProbeItem], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: 'sessionId', marker: "dot"}),
					Plot.tip(metricData[subProbeItem], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
				]
			}} />
		</div>
		{/if}
		{/each}
	</div>
	{:else}
		{#if probeItem in metricData }
			<div bind:clientWidth={clientFullScreenWidth} class="mt-8 pt-4 w-full">
				<div class="grid grid-cols-4">
					<div class="sm:col-span-1">
						<h2>{PIFINA_PROBE_CHART_CFG[probeItem]['title']}</h2>
					</div>
					<div class="sm:col-span-3 justify-self-end">
						<button on:click={() => openDetailView(String(probeItem))} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
							<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
							</svg>					  
							Open
						</button>
					</div>
				</div>
				<Chart options={{
					x: xScaleOptions,
					y: {
						label: PIFINA_PROBE_CHART_CFG[probeItem]['yAxisName'],
						grid: true
					},
					width: clientFullScreenWidth,
					color: {legend: true, type: "categorical"},
					marks: [
						Plot.line(metricData[probeItem], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: "sessionId", marker: "dot"}),
						Plot.tip(metricData[probeItem], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
					]
				}} />
			</div>
		{/if}
	{/if}		
	{/each}
</div>
{/if}
{#if isExtraProbesChartSelected() == true }
<div class="divide-y divide-solid">
	{#each [...extraProbeNames.values()] as entry }
	<div bind:clientWidth={clientFullScreenWidth} class="mt-8 pt-4">
		<div class="grid grid-cols-4">
			<div class="sm:col-span-1">
				<h2>{entry}</h2>
			</div>
			<div class="sm:col-span-3 justify-self-end">
				<button on:click={() => openDetailView(entry)} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
					</svg>					  
					Open
				</button>
			</div>
		</div>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "bytes/sec",
				grid: true
			},
			width: clientFullScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[entry], {x: "timestamp", y: "value", stroke: (d) => `${d.sessionId}`, marker: "dot"}),
				Plot.tip(metricData[entry], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
			]
		}} />
	</div>
	{/each}
</div>
{/if}
{#if isAppRegChartSelected() == true }
<div class="divide-y divide-solid">
	{#each [...appRegister.values()] as entry }
	<div bind:clientWidth={clientFullScreenWidth} class="mt-8 pt-4">
		<div class="grid grid-cols-4">
			<div class="sm:col-span-1">
				<h2>{entry} register (app-owned)</h2>
			</div>
			<div class="sm:col-span-3 justify-self-end">
				<button on:click={() => openDetailView(entry)} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
					</svg>					  
					Open
				</button>
			</div>
		</div>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "current",
				grid: true
			},
			width: clientFullScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[entry], {x: "timestamp", y: "value", stroke: (d) => `${d.sessionId}`, marker: "dot"}),
				Plot.tip(metricData[entry], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}})),
			]
		}} />
	</div>		
	{/each}
</div>
{/if}
{#if isTMChartSelected() == true }
<div class="divide-y divide-solid">
	<TmMetricCharts metricData={metricData} openDetailView={openDetailView} />
</div>
{/if}
{/key}