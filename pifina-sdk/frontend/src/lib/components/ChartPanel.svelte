<script lang="ts">
    import Chart from './Chart.svelte';
    import * as Plot from "@observablehq/plot";
    import type { MetricItem } from '../models/MetricItem';
	import { sessionFilterStore } from '$lib/stores/sessionFilterStore';
	import { endpointFilterStore } from '$lib/stores/endpointFilterStore';
	import { groupIdFilterStore } from '$lib/stores/groupIdFilterStore';
	import { onDestroy } from 'svelte';

    export let chartTitle: string;
    export let screenWidth: number;
    export let xScaleOptions: Plot.ScaleOptions = {
		label: "Timestamp",
		tickSpacing: 150,
		tickFormat: (value, _) => (value.toLocaleString(undefined, {
			hour: 'numeric',
			minute: 'numeric',
			second: 'numeric'
		})),
	};
    export let yAxisLabel: string = "unknown";
    export let metricData: MetricItem[];
    export let metricAttributeName: string;
    export let yAxisTickFormat: string = "s";
    export let disableSeriesFilter = false;

    let selectedSessionIds: number[] = [];
    let selectedEndpoint: string = "";
    let selectedGroupId: number = 1;

    const sessionFilterStoreSub = sessionFilterStore.subscribe(val => selectedSessionIds = val);
    const endpointFilterStoreSub = endpointFilterStore.subscribe(val => selectedEndpoint = val);
    const groupIdFilterStoreSub = groupIdFilterStore.subscribe(val => selectedGroupId = val);

    const openDetailView = () => {
		window.open(`/dashboard/detail?groupId=${selectedGroupId}&endpoint=${selectedEndpoint}&selectedMetric=${metricAttributeName}`, "_blank");
	}

    onDestroy(sessionFilterStoreSub);
    onDestroy(endpointFilterStoreSub);
    onDestroy(groupIdFilterStoreSub);
</script>

<div bind:clientWidth={screenWidth} class="mt-8 pt-4">
    <div class="grid grid-cols-4">
        <div class="sm:col-span-3">
            <h2>{chartTitle}</h2>
        </div>
        <div class="sm:col-span-1 justify-self-end pr-4">
            <button on:click={() => openDetailView()} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
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
            label: yAxisLabel,
            grid: true,
            tickFormat: yAxisTickFormat
        },
        width: screenWidth,
        color: {legend: true, type: "categorical"},
        marks: [
            Plot.line(metricData, {filter: (d) => (disableSeriesFilter || selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", z: "sessionId", stroke: "sessionId", marker: "dot"}),
            Plot.tip(metricData, Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (disableSeriesFilter || selectedSessionIds.includes(d.sessionId))})),
        ]
    }} />
</div>