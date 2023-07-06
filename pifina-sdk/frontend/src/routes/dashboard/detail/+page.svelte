<script lang="ts">
    import * as Plot from "@observablehq/plot";
	import Chart from "../../../lib/components/Chart.svelte";
	import { page } from '$app/stores';
	import type { DTOPifinaMetricItem, MetricItem } from "$lib/models/MetricItem";
	import { PIFINA_PROBE_CHART_CFG, Y_AXIS_NAME_BYTE_RATE } from "$lib/models/metricNames";

	let selectedMetric = $page.url.searchParams.get('selectedMetric') || "";
	let metricData: MetricItem[] = [];
	let sessionIds = new Set<number>();
	let selectedSessionIds: number[] = [];
	let sessionIdFilterIsDirty = false;

	let worker = new SharedWorker(new URL('$lib/sharedworker/sharedworker.ts', import.meta.url), {type: 'module'});

	worker.port.postMessage({status: "CONNECT", endpoint: "tofino-dev"});
	worker.port.onmessage = (event) => {
		let dataobj: DTOPifinaMetricItem[] = event.data;

		dataobj.forEach(item => {
			let key = `${item.metricName}${item.type}`;
			// Check if it's a metric from a default probe
			if (!item.metricName.startsWith("PF_")) {
				key = item.metricName;
			}
			if (item.metricName.startsWith("PF_TM_")) {
				key = item.metricName;
			}
			if (item.metricName.startsWith("PF_EXTRA")) {
				key = item.metricName;
			}
			if (selectedMetric === key) {
				if (!sessionIds.has(item.sessionId) && !item.metricName.startsWith("PF_TM_")) {
					sessionIds.add(item.sessionId);
				}
				metricData.push({timestamp: new Date(item.timestamp), value: item.value, sessionId: item.sessionId, type: item.type});
			}
		})

		// Default initialization of filters
		if (!sessionIdFilterIsDirty && selectedSessionIds.length == 0) {
			for (const sessionId of sessionIds.values()) {
				selectedSessionIds.push(sessionId);
			}
		}

		// Limit series length
		let mapKeySessioId = new Set<number>();
			metricData.forEach(item => {
				mapKeySessioId.add(item.sessionId);
		});
		const metricSizeLimit = (mapKeySessioId.size + 1) * 20;
		if (metricData.length > metricSizeLimit) {
			metricData.splice(0, metricData.length - metricSizeLimit);
		}
		// Force rerender
		sessionIds = sessionIds;
		metricData = metricData;
	}


    let cliendScreenWidth;

	let yAxisName = "current";
	let title = selectedMetric in PIFINA_PROBE_CHART_CFG ? PIFINA_PROBE_CHART_CFG[selectedMetric]['title'] : selectedMetric;
	if (selectedMetric in PIFINA_PROBE_CHART_CFG) {
		yAxisName = PIFINA_PROBE_CHART_CFG[selectedMetric]['yAxisName'];			
	}
	if (selectedMetric.startsWith("PF_EXTRA")) {
		yAxisName = Y_AXIS_NAME_BYTE_RATE;
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

<main>
    <div class="mx-auto max-w-screen-2xl py-6 sm:px-6 lg:px-8">
        <div class="bg-white rounded-lg px-8 py-8 shadow-lg">
			<div class="divide-y divide-solid">
				{#if sessionIds.size > 0}
				<div class="mt-2">
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
				{/if}
				<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4 w-full">
					<h2>{title}</h2>
					{#if sessionIds.size > 0}
					<Chart options={{
						x: xScaleOptions,
						y: {
							label: yAxisName,
							grid: true
						},
						width: cliendScreenWidth,
						color: {legend: true, type: "categorical"},
						marks: [
							Plot.line(metricData, {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: "sessionId", marker: "dot"}),
							Plot.tip(metricData, Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
						]
					}} />
					{:else}
					<Chart options={{
						x: xScaleOptions,
						y: {
							label: yAxisName,
							grid: true
						},
						width: cliendScreenWidth,
						color: {legend: true, type: "categorical"},
						marks: [
							Plot.line(metricData, {x: "timestamp", y: "value", marker: "dot"}),
							Plot.tip(metricData, Plot.pointerX({x: "timestamp", y: "value"})),
						]
					}} />
					{/if}
				</div>
			</div>
        </div>
    </div>
</main>