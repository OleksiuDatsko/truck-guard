<script lang="ts">
  import "./layout.css";
  import favicon from "$lib/assets/favicon.svg";

  let { children, data } = $props();
  $inspect(data);
</script>

<svelte:head
  ><link rel="icon" href={favicon} /><title>TruckGuard</title></svelte:head
>
{@render children()}
