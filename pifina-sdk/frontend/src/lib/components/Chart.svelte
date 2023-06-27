<script lang="ts">
    import * as Plot from "@observablehq/plot";

    export let options: Plot.PlotOptions;

    function myPlot(node: Node) {
        node.appendChild(
            Plot.plot(options)
        );
    }
</script>

{#key options}
    <div use:myPlot {...$$restProps}></div>
{/key}