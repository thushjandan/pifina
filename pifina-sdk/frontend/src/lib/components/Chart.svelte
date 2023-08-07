<!--
 Copyright (c) 2023 Thushjandan Ponnudurai
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

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