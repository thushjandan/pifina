<script lang="ts">
	import ChartPanel from "$lib/components/ChartPanel.svelte";
	import { getPifinaChartConfigByMetricName } from "$lib/config/chartConfig";
	import { PIFINA_DASHBOARD_CONF } from "$lib/config/dashboardConfig";
	import type { MetricData } from "$lib/models/MetricItem";

    export let metricData: MetricData = {};

    // Select default view
	let selectedChartCategory: string = PIFINA_DASHBOARD_CONF.HOSTTYPE_NIC[0].key;

    let clientFullScreenWidth: number;
	let clientHalfScreenWidth: number;

</script>

{#key selectedChartCategory }
<div class="mt-6 text-sm font-medium text-center text-gray-500 border-b border-gray-200 dark:text-gray-400 dark:border-gray-700">
    <ul class="flex flex-wrap -mb-px">
        {#each PIFINA_DASHBOARD_CONF.HOSTTYPE_NIC as menuItem }
        <li class="mr-2">
            <a href={null} on:click={() => selectedChartCategory = menuItem.key} class="inline-block p-4 border-b-2 rounded-t-lg hover:text-gray-600 hover:border-gray-300" 
                class:border-transparent={selectedChartCategory != menuItem.key} class:text-indigo-600={selectedChartCategory == menuItem.key} 
				class:border-blue-600={selectedChartCategory == menuItem.key} class:hover:text-gray-600={selectedChartCategory != menuItem.key} 
                class:hover:border-gray-300={selectedChartCategory != menuItem.key}>
				{menuItem.title}
			</a>
        </li>
        {/each}
    </ul>
</div>
{#each PIFINA_DASHBOARD_CONF.HOSTTYPE_NIC as confItem, i}
    {#if confItem.key == selectedChartCategory }
        <div class="divide-y divide-solid">
        {#if confItem.type == "static" && confItem.charts !== undefined}
            {#each confItem.charts as probeItem}
                {#if Array.isArray(probeItem)}
                <div class="mt-8 grid md:grid-cols-2">
                    {#each probeItem as subProbeItem }
                        {#if subProbeItem in metricData }
                            <ChartPanel chartTitle={getPifinaChartConfigByMetricName(subProbeItem).title} metricAttributeName={subProbeItem} 
                            metricData={metricData[subProbeItem]} yAxisLabel={getPifinaChartConfigByMetricName(subProbeItem).yAxisName} 
                            screenWidth={clientHalfScreenWidth} disableSeriesFilter={confItem.disableSessionFilter} />
                        {/if}
                    {/each}
                </div>
                {:else}
                    {#if probeItem in metricData }
                        <ChartPanel chartTitle={getPifinaChartConfigByMetricName(probeItem).title} metricAttributeName={probeItem} 
                        metricData={metricData[probeItem]} yAxisLabel={getPifinaChartConfigByMetricName(probeItem).yAxisName} 
                        screenWidth={clientFullScreenWidth} disableSeriesFilter={confItem.disableSessionFilter} />
                    {/if}
                {/if}
            {/each}
        {:else}
        Unknown chart type
        {/if}
        </div>
    {/if}
{/each}
{/key}