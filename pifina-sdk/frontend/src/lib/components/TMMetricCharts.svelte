<script lang="ts">
    import * as Plot from "@observablehq/plot";
	import type { MetricData } from "$lib/models/MetricItem";
    import { PROBE_INGRESS_MATCH_CNT_BYTE, PROBE_TM_EGRESS_DROP_PKT, PROBE_TM_ERESS_USAGE_CELLS, PROBE_TM_INGRESS_DROP_PKT, PROBE_TM_INRESS_USAGE_CELLS, PROBE_TM_PIPE_EG_DROP_PKT, PROBE_TM_PIPE_IG_FULL_BUF, PROBE_TM_PIPE_TOTAL_BUF_DROP } from '$lib/models/metricNames';
	import Chart from "./Chart.svelte";

    export let metricData: MetricData = {};
    export let openDetailView: Function;

    let clientHalfScreenWidth;

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


<div class="mt-8 grid md:grid-cols-2">
    {#if PROBE_TM_INGRESS_DROP_PKT in metricData }
    <div bind:clientWidth={clientHalfScreenWidth} class="mt-8 pt-4">
        <h2>Ingress & egress packet drops from TM perspective</h2>
        <Chart options={{
            x: xScaleOptions,
            y: {
                label: "pkts",
                grid: true
            },
            width: clientHalfScreenWidth,
            color: {legend: true, type: "categorical"},
            marks: [
                Plot.line(metricData[PROBE_TM_INGRESS_DROP_PKT], {x: "timestamp", y: "value", stroke: (d) => `ingress: ${d.sessionId}`, marker: "dot"}),
                Plot.line(metricData[PROBE_TM_EGRESS_DROP_PKT], {x: "timestamp", y: "value", stroke: (d) => `egress: ${d.sessionId}`, marker: "dot"}),
                Plot.tickY(metricData[PROBE_TM_INGRESS_DROP_PKT], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `ingress: ${d.sessionId}`}),
                Plot.tickY(metricData[PROBE_TM_EGRESS_DROP_PKT], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `egress: ${d.sessionId}`})
            ]
        }} />
    </div>
    <div bind:clientWidth={clientHalfScreenWidth} class="mt-8 pt-4">
        <h2>Port usage count in terms of number of memory cells usage from TM ingress & egress perspective</h2>
        <Chart options={{
            x: xScaleOptions,
            y: {
                label: "pkts",
                grid: true
            },
            width: clientHalfScreenWidth,
            color: {legend: true, type: "categorical"},
            marks: [
                Plot.line(metricData[PROBE_TM_INRESS_USAGE_CELLS], {x: "timestamp", y: "value", stroke: (d) => `ingress: ${d.sessionId}`, marker: "dot"}),
                Plot.line(metricData[PROBE_TM_ERESS_USAGE_CELLS], {x: "timestamp", y: "value", stroke: (d) => `egress: ${d.sessionId}`, marker: "dot"}),
                Plot.tickY(metricData[PROBE_TM_INRESS_USAGE_CELLS], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `ingress: ${d.sessionId}`}),
                Plot.tickY(metricData[PROBE_TM_ERESS_USAGE_CELLS], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `egress: ${d.sessionId}`})
            ]
        }} />
    </div>
    {/if}
    {#if PROBE_TM_PIPE_TOTAL_BUF_DROP in metricData }
    <div bind:clientWidth={clientHalfScreenWidth} class="mt-8 pt-4">
        <div class="grid grid-cols-4">
            <div class="sm:col-span-3">
                <h2>Number of packets which were dropped because of buffer full condition</h2>
            </div>
            <div class="sm:col-span-1 justify-self-end pr-4">
                <button on:click={() => openDetailView(PROBE_TM_PIPE_TOTAL_BUF_DROP)} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
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
                label: "pkts",
                grid: true
            },
            width: clientHalfScreenWidth,
            color: {legend: true, type: "categorical"},
            marks: [
                Plot.line(metricData[PROBE_TM_PIPE_TOTAL_BUF_DROP], {x: "timestamp", y: "value", stroke: (d) => `pipeline: ${d.sessionId}`,  marker: "dot"}),
                Plot.tip(metricData[PROBE_TM_PIPE_TOTAL_BUF_DROP], Plot.pointerX({x: "timestamp", y: "value"})),
            ]
        }} />
    </div>
    <div bind:clientWidth={clientHalfScreenWidth} class="mt-8 pt-4">
        <div class="grid grid-cols-4">
            <div class="sm:col-span-3">
                <h2>The number of packets which were dropped because of buffer full condition on ingress side</h2>
            </div>
            <div class="sm:col-span-1 justify-self-end pr-4">
                <button on:click={() => openDetailView(PROBE_TM_PIPE_IG_FULL_BUF)} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
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
                label: "pkts",
                grid: true
            },
            width: clientHalfScreenWidth,
            color: {legend: true, type: "categorical"},
            marks: [
                Plot.line(metricData[PROBE_TM_PIPE_IG_FULL_BUF], {x: "timestamp", y: "value", stroke: (d) => `pipeline: ${d.sessionId}`, marker: "dot"}),
                Plot.tip(metricData[PROBE_TM_PIPE_IG_FULL_BUF], Plot.pointerX({x: "timestamp", y: "value"})),
            ]
        }} />
    </div>
    <div bind:clientWidth={clientHalfScreenWidth} class="mt-8 pt-4">
        <div class="grid grid-cols-4">
            <div class="sm:col-span-3">
                <h2>The total number of packets which were dropped on egress side.</h2>
            </div>
            <div class="sm:col-span-1 justify-self-end pr-4">
                <button on:click={() => openDetailView(PROBE_TM_PIPE_EG_DROP_PKT)} class="bg-indigo-600 hover:bg-indigo-800 text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
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
                label: "pkts",
                grid: true
            },
            width: clientHalfScreenWidth,
            color: {legend: true, type: "categorical"},
            marks: [
                Plot.line(metricData[PROBE_TM_PIPE_EG_DROP_PKT], {x: "timestamp", y: "value", stroke: (d) => `pipeline: ${d.sessionId}`, marker: "dot"}),
                Plot.tip(metricData[PROBE_TM_PIPE_EG_DROP_PKT], Plot.pointerX({x: "timestamp", y: "value"}))
            ]
        }} />
    </div>
    {/if}
</div>