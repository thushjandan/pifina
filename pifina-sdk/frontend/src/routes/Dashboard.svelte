<!--
 Copyright (c) 2023 Thushjandan Ponnudurai
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import type {  MetricData, MetricNameGroup } from '../lib/models/metricItem';
	import { EndpointType, type DTOTelemetryMessage, type EndpointModel } from '$lib/models/endpointModel';
	import { endpointFilterStore } from '$lib/stores/endpointFilterStore';
	import { sessionFilterStore } from '$lib/stores/sessionFilterStore';
	import TofinoDashboardType from '$lib/dashboardType/TofinoDashboardType.svelte';
	import NicDashboardType from '$lib/dashboardType/NICDashboardType.svelte';
	import { groupIdFilterStore } from '$lib/stores/groupIdFilterStore';

    export let endpoints: EndpointModel[];

	let groupIds: number[] = [...new Set(endpoints.map(item => item.groupId))];
	let selectedGroupid: number = endpoints[0]?.groupId || 1;
    let selectedEndpoint: EndpointModel = endpoints[0];
	let sessionIds = new Set<number>();
	let selectedSessionIds: number[] = [];
	let sessionIdFilterIsDirty = false;
	let metricData: MetricData = {};
	let metricNamesGroupedByType: MetricNameGroup = {
		"appRegister": new Set<string>(),
		"extraProbes": new Set<string>(),
		"tmMetrics": new Set<string>(),
	}
	let isEnabled: boolean = true;

	endpointFilterStore.set(endpoints[0]?.name || "");
	groupIdFilterStore.set(endpoints[0]?.groupId || 1);
	let worker = new SharedWorker(new URL('$lib/sharedworker/sharedworker.ts', import.meta.url), {type: 'module'});

	worker.port.postMessage({status: "CONNECT", groupId: selectedGroupid});
	worker.port.onmessage = (event: MessageEvent) => {
		let telemetryMessage: DTOTelemetryMessage = event.data;
		// Skip any message not related
		if (telemetryMessage.source != selectedEndpoint.name) {
			return
		}

		telemetryMessage.metrics.forEach(item => {
			let key = item.metricName;
			if (telemetryMessage.type === EndpointType.HOSTTYPE_TOFINO) {
				key = `${item.metricName}${item.type}`;
				// Check if it's a metric from a default probe
				if (!item.metricName.startsWith("PF_")) {
					metricNamesGroupedByType["appRegister"].add(item.metricName);
					key = item.metricName;
				}
				if (item.metricName.startsWith("PF_TM_")) {
					metricNamesGroupedByType["tmMetrics"].add(item.metricName)
					key = item.metricName;
				}
				if (item.metricName.startsWith("PF_EXTRA")) {
					metricNamesGroupedByType["extraProbes"].add(item.metricName);
					key = item.metricName;
				}
			}
			// check if key exists. If not, create a new list.
			if (!(key in metricData)) {
				metricData[key] = [];
			}
			if (!sessionIds.has(item.sessionId) && !item.metricName.startsWith("PF_TM_")) {
				sessionIds.add(item.sessionId);
			}
			metricData[key].push({timestamp: new Date(item.timestamp), value: item.value, sessionId: item.sessionId, type: item.type});
		})

		// Default initialization of filters
		if (!sessionIdFilterIsDirty && selectedSessionIds.length == 0) {
			for (const sessionId of sessionIds.values()) {
				selectedSessionIds.push(sessionId);
			}
			sessionIds = sessionIds;
			sessionFilterStore.set(selectedSessionIds);
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


	const onHostGroupChange = () => {
		groupIdFilterStore.set(selectedGroupid);
		worker.port.postMessage({status: "CONNECT", groupId: selectedGroupid});
	}

	const onEndpointChange = () => {
		// Delete all existing metrics from map
		// Will be refilled by event source.
		metricData = {};
		endpointFilterStore.set(selectedEndpoint.name);
	}

	const onSessionIdFilterChange = () => {
		sessionIdFilterIsDirty = true;
		sessionFilterStore.set(selectedSessionIds);
	}

	const toggleEventStreaming = () => {
		if (isEnabled) {
			// Close previous event source
			worker.port.postMessage({status: "CLOSE", groupId: selectedGroupid});
		}
		if (!isEnabled) {
			worker.port.postMessage({status: "CONNECT", groupId: selectedGroupid});
		}
		isEnabled = !isEnabled;
	}

</script>

<div class="grid grid-cols-4">
	<div class="sm:col-span-1 mx-2">
		<label for="target" class="block text-sm font-medium leading-6 text-gray-900">Choose a Group:</label>
		<div class="mt-2">
			<select bind:value={selectedGroupid} on:change={onHostGroupChange} name="target" class="px-3 py-3 placeholder-slate-300 text-slate-600 relative bg-white bg-white rounded text-sm border-0 shadow outline-none focus:outline-none focus:ring w-full">
				{#each groupIds as id }
				<option value={id}>Group {id}</option>
				{/each}
			</select>
		</div>
	</div>
	<div class="sm:col-span-1 mx-2">
		<label for="target" class="block text-sm font-medium leading-6 text-gray-900">Choose a monitoring target:</label>
		<div class="mt-2">
			<select bind:value={selectedEndpoint} on:change={onEndpointChange} name="target" class="px-3 py-3 placeholder-slate-300 text-slate-600 relative bg-white bg-white rounded text-sm border-0 shadow outline-none focus:outline-none focus:ring w-full">
				{#each endpoints as endpoint }
				<option value={endpoint}>{endpoint.name}</option>
				{/each}
			</select>
		</div>
	</div>
	<div class="sm:col-span-2 justify-self-end">
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
<div class="mt-8">
	<div class="sm:col-span-1">
		<label for="sessionIds" class="block text-sm font-medium leading-6 text-gray-900">Filter by session ID:</label>
		<div class="mt-2 flex flex-row">
			{#each [...sessionIds.values()] as sessionId}
					<div class="items-center">
						<input type=checkbox bind:group={selectedSessionIds} on:change={onSessionIdFilterChange} name="sessionIds" value={sessionId} class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600" />
						<label for="comments" class="ml-1 mr-4 font-medium text-gray-900">{sessionId}</label>
					</div>
			{/each}
		</div>
	</div>
</div>
{#if selectedEndpoint.type == EndpointType.HOSTTYPE_TOFINO }
<TofinoDashboardType metricData={metricData} metricNameGroup={metricNamesGroupedByType}></TofinoDashboardType>
{:else if selectedEndpoint.type == EndpointType.HOSTTYPE_NIC }
<NicDashboardType metricData={metricData}></NicDashboardType>
{:else}
<p class="mb-2 text-lg text-gray-500 md:text-xl dark:text-gray-400">Unknown Host type. Cannot visualize metrics for this type of MetricTypes.</p>
{/if}