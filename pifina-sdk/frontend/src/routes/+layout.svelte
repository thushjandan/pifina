<script lang="ts">
    import "../app.css";
    import { fade } from 'svelte/transition';
    import { page } from '$app/stores';
    let path: string;
    let mobileNavOpen = false;

    $: path = $page.url.pathname;

    const isHome = () => path === '/';
    const isConfig = () => path ==='/config';
    const isAbout = () => path === '/about';

</script>
<nav class="bg-indigo-500">
  <div class="mx-auto max-w-7xl px-2 sm:px-6 lg:px-8">
    <div class="relative flex h-16 items-center justify-between">
      <div class="absolute inset-y-0 left-0 flex items-center sm:hidden">
        <!-- Mobile menu button-->
        <button type="button" on:click={() => mobileNavOpen = !mobileNavOpen} class="inline-flex items-center justify-center rounded-md p-2 text-indigo-400 hover:bg-indigo-700 hover:text-white focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white" aria-controls="mobile-menu" aria-expanded="false">
          <span class="sr-only">Open main menu</span>
          <!--
            Icon when menu is closed.

            Menu open: "hidden", Menu closed: "block"
          -->
          <svg class="block h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
          </svg>
          <!--
            Icon when menu is open.

            Menu open: "block", Menu closed: "hidden"
          -->
          <svg class="hidden h-6 w-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <a href="/" class="ml-12 md:ml-0 flex items-center">
          <span class="self-center text-2xl font-semibold whitespace-nowrap text-white">PIFINA</span>
      </a>
      <div class="flex flex-grow items-center justify-end sm:items-stretch">
        <div class="hidden sm:ml-6 sm:block">
          <div class="flex space-x-4">
            {#key path}
            <a href="/" class:bg-indigo-700={isHome()} class:text-white={isHome()} class:text-indigo-300={!isHome()} class:hover:bg-indigo-700={!isHome()} class:hover:text-white={!isHome()} class="rounded-md px-3 py-2 text-sm font-medium" aria-current="page">Dashboard</a>
            <a href="/config" class:bg-indigo-700={isConfig()} class:text-white={isConfig()} class:text-indigo-300={!isConfig()} class:hover:bg-indigo-700={!isConfig()} class:hover:text-white={!isConfig()} class="rounded-md px-3 py-2 text-sm font-medium" aria-current="page">Configuration</a>
            <a href="/about" class:bg-indigo-700={isAbout()} class:text-white={isAbout()} class:text-indigo-300={!isAbout()} class:hover:bg-indigo-700={!isAbout()} class:hover:text-white={!isAbout()} class="rounded-md px-3 py-2 text-sm font-medium" aria-current="page">About</a>
            {/key}
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Mobile menu, show/hide based on menu state. -->
  {#if mobileNavOpen}
  <div transition:fade class="sm:hidden" id="mobile-menu">
    <div class="space-y-1 px-2 pb-3 pt-2">
      <!-- Current: "bg-indigo-900 text-white", Default: "text-indigo-300 hover:bg-indigo-700 hover:text-white" -->
      {#key path}
      <a href="/" class:bg-indigo-900={isHome()} class:text-white={isHome()} class:text-indigo-300={!isHome()} class:hover:bg-indigo-700={!isHome()} class:hover:text-white={!isHome()} class="bg-indigo-900 text-white block rounded-md px-3 py-2 text-base font-medium" aria-current="page">Dashboard</a>
      <a href="/config" class:bg-indigo-700={isConfig()} class:text-white={isConfig()} class:text-indigo-300={!isConfig()} class:hover:bg-indigo-700={!isConfig()} class:hover:text-white={!isConfig()} class="text-white block rounded-md px-3 py-2 text-base font-medium" aria-current="page">Configuration</a>
      <a href="/about" class:bg-indigo-700={isAbout()} class:text-white={isAbout()} class:text-indigo-300={!isAbout()} class:hover:bg-indigo-700={!isAbout()} class:hover:text-white={!isAbout()} class="text-white block rounded-md px-3 py-2 text-base font-medium" aria-current="page">About</a>
      {/key}
    </div>
  </div>
  {/if}
</nav>

<slot />