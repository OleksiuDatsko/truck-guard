<script lang="ts">
  import "./layout.css";
  import favicon from "$lib/assets/favicon.svg";
  import AppSidebar from "$lib/components/app-sidebar.svelte";
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import { Separator } from "$lib/components/ui/separator/index.js";
  import { ModeWatcher } from "mode-watcher";
  import { Toaster } from "$lib/components/ui/sonner";

  let { children, data } = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>TruckGuard</title>
</svelte:head>
<ModeWatcher defaultMode="light" />
<Toaster />

{#if data.user}
  <Sidebar.Provider>
    <AppSidebar user={data.user} />
    <Sidebar.Inset>
      <header class="flex h-14 shrink-0 items-center gap-2 border-b px-4">
        <Sidebar.Trigger class="-ms-1" />
        <Separator orientation="vertical" class="me-2 h-4" />
      </header>
      <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
      {@render children()}
      </div>
    </Sidebar.Inset>
  </Sidebar.Provider>
{:else}
  {@render children()}
{/if}
