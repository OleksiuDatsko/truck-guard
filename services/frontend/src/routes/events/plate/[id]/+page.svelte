<script lang="ts">
  import { enhance } from "$app/forms";
  import { page } from "$app/state";
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Separator } from "$lib/components/ui/separator";
  import {
    ChevronLeft,
    Clock,
    Camera,
    CreditCard,
    Check,
    X,
    Eye,
    Pencil,
  } from "@lucide/svelte";
  import { toast } from "svelte-sonner";

  let { data } = $props();
  // Initialize as state for optimistic updates
  let event = $state(data.event);
  const user = $derived(data.user);

  const canEdit = $derived(
    user?.permissions?.includes("update:events") ?? false,
  );

  let isEditing = $state(false);
  let loading = $state(false);

  function formatDate(dateStr: string) {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleString("uk-UA");
  }

  function toggleEdit() {
    isEditing = !isEditing;
  }
</script>

<div class="container mx-auto py-6 space-y-6 max-w-5xl">
  <!-- Header -->
  <div class="flex items-center gap-4">
    <Button variant="outline" size="icon" href="/events" class="h-8 w-8">
      <ChevronLeft class="h-4 w-4" />
    </Button>
    <div>
      <h1 class="text-2xl font-bold tracking-tight text-foreground">
        Подія розпізнавання #{event.ID}
      </h1>
      <p class="text-muted-foreground text-sm">Детальна інформація про подію</p>
    </div>
    {#if event.is_manual}
      <div
        class="ml-auto bg-amber-50 dark:bg-amber-950/30 text-amber-700 dark:text-amber-400 px-3 py-1 rounded-full text-xs font-bold border border-amber-200 dark:border-amber-800"
      >
        ВРУЧНУ
      </div>
    {:else}
      <div
        class="ml-auto bg-blue-50 dark:bg-blue-950/30 text-blue-700 dark:text-blue-400 px-3 py-1 rounded-full text-xs font-bold border border-blue-200 dark:border-blue-800"
      >
        ANPR
      </div>
    {/if}
  </div>

  <div class="grid gap-6 md:grid-cols-2">
    <!-- Image Section -->
    <Card.Root class="overflow-hidden md:col-span-1 shadow-lg bg-card/50 gap-0 py-0">
      <Card.Content
        class="p-0 relative aspect-video bg-black/5 flex items-center justify-center h-full min-h-[300px]"
      >
        {#if event.image_key}
          <img
            src={`/api/images/${event.image_key}`}
            alt="Vehicle Plate"
            class="w-full h-full object-cover"
          />
          <div
            class="absolute inset-0 opacity-20 hover:opacity-100 transition-opacity flex items-end justify-center p-4"
          >
            <Button
              variant="secondary"
              size="sm"
              class="gap-2"
              href={`/api/images/${event.image_key}`}
              target="_blank"
            >
              <Eye class="h-4 w-4" /> Відкрити оригінал
            </Button>
          </div>
        {:else}
          <div class="flex flex-col items-center gap-2 text-muted-foreground">
            <Camera class="h-12 w-12 opacity-20" />
            <span class="text-sm">Зображення відсутнє</span>
          </div>
        {/if}
      </Card.Content>
    </Card.Root>

    <!-- Details Section -->
    <div class="md:col-span-1 space-y-6">
      <Card.Root class="shadow-md">
        <Card.Header class="pb-3">
          <Card.Title>Основні дані</Card.Title>
        </Card.Header>
        <Card.Content class="space-y-6">
          <div class="space-y-4">
            <!-- Plate Number Block -->
            <div class="bg-muted/30 p-4 rounded-xl border border-border/50">
              <div class="flex items-center justify-between mb-2">
                <Label
                  class="text-muted-foreground flex items-center gap-2 text-xs uppercase tracking-wider font-semibold"
                >
                  <CreditCard class="h-3 w-3" /> Номерний знак
                </Label>
                {#if !isEditing && canEdit}
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-6 w-6 p-0 hover:bg-background"
                    onclick={toggleEdit}
                  >
                    <Pencil class="h-3.5 w-3.5 text-muted-foreground" />
                  </Button>
                {/if}
              </div>

              {#if isEditing}
                <form
                  method="POST"
                  action="?/correct"
                  use:enhance={({ formData }) => {
                    loading = true;
                    const newPlate = formData.get("plate") as string;
                    const originalPlate = event.plate_corrected || event.plate;

                    // Optimistic update
                    event.plate_corrected = newPlate;
                    event.is_manual = true;
                    // Construct display name
                    const userName =
                      [user?.first_name, user?.last_name]
                        .filter(Boolean)
                        .join(" ") ||
                      user?.username ||
                      "You";
                    event.corrected_by_name = userName;

                    isEditing = false; // Close edit mode immediately

                    return async ({ result, update }) => {
                      loading = false;
                      if (result.type === "success") {
                        toast.success("Номер успішно змінено");
                        // Optionally apply server state if needed, but optimistic is usually enough
                        // await update();
                      } else {
                        // Revert on error
                        event.plate_corrected = originalPlate;
                        toast.error("Помилка при зміні номеру");
                      }
                    };
                  }}
                  class="flex gap-2 items-center"
                >
                  <Input
                    name="plate"
                    value={event.plate_corrected || event.plate}
                    class="font-mono text-lg font-bold uppercase h-10 bg-background"
                    autofocus
                  />
                  <Button
                    type="submit"
                    size="sm"
                    disabled={loading}
                    class="h-10 w-10 p-0"
                  >
                    {#if loading}
                      <span class="loading loading-spinner loading-xs"></span>
                    {:else}
                      <Check class="h-4 w-4" />
                    {/if}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    class="h-10 w-10 p-0"
                    onclick={toggleEdit}
                    disabled={loading}
                  >
                    <X class="h-4 w-4" />
                  </Button>
                </form>
              {:else}
                <div class="flex items-baseline gap-3">
                  <span
                    class="text-3xl font-black font-mono tracking-wider text-foreground"
                  >
                    {event.plate_corrected || event.plate}
                  </span>
                  {#if event.plate_corrected}
                    <span
                      class="text-xs text-muted-foreground line-through decoration-red-500/50"
                    >
                      {event.plate}
                    </span>
                  {/if}
                </div>
                {#if event.corrected_by_name || event.corrected_by}
                  <div
                    class="mt-2 text-xs text-muted-foreground flex gap-1 items-center"
                  >
                    <Clock class="h-3 w-3" />
                    Відкориговано користувачем:
                    <span class="font-medium text-foreground"
                      >{event.corrected_by_name || event.corrected_by}</span
                    >
                  </div>
                {/if}
              {/if}
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-1">
                <Label
                  class="text-xs text-muted-foreground uppercase tracking-wider"
                  >Час фіксації</Label
                >
                <div class="flex items-center gap-2 font-medium">
                  <Clock class="h-4 w-4 text-muted-foreground" />
                  {formatDate(event.timestamp)}
                </div>
              </div>
              <div class="space-y-1">
                <Label
                  class="text-xs text-muted-foreground uppercase tracking-wider"
                  >Камера</Label
                >
                <div class="flex items-center gap-2 font-medium">
                  <Camera class="h-4 w-4 text-muted-foreground" />
                  {event.camera_name || event.camera_id}
                </div>
              </div>
            </div>
          </div>
        </Card.Content>
      </Card.Root>

      <!-- Additional Info (JSON Payload or other) -->
      <Card.Root class="border-0 shadow-sm bg-muted/10">
        <Card.Header>
          <Card.Title class="text-sm font-medium text-muted-foreground"
            >Додаткова інформація</Card.Title
          >
        </Card.Header>
        <Card.Content>
          <dl class="grid grid-cols-2 gap-6">
            <div>
              <dt class="text-xs font-medium text-muted-foreground">
                Системна подія
              </dt>
              <dd class="mt-1 text-sm text-foreground font-mono">
                {#if event.system_event_id}
                  <a
                    href={`/events/system/${event.system_event_id}`}
                    class="underline hover:text-blue-500 transition-colors"
                    >#{event.system_event_id}</a
                  >
                {:else}
                  -
                {/if}
              </dd>
            </div>
            <div>
              <dt class="text-xs font-medium text-muted-foreground">
                Перепустка
              </dt>
              <dd class="mt-1 text-sm text-foreground font-mono">
                {#if event.permit_id}
                  <a
                    href={`/permits/${event.permit_id}`}
                    class="underline hover:text-blue-500 transition-colors"
                    >#{event.permit_id}</a
                  >
                {:else}
                  -
                {/if}
              </dd>
            </div>
          </dl>
        </Card.Content>
      </Card.Root>
    </div>
  </div>
</div>
