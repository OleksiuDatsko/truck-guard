<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import {
    ArrowLeft,
    Clock,
    Activity,
    Database,
    FileQuestionMark
  } from "@lucide/svelte";
  import type { SystemEvent } from "$lib/types/events";

  // Svelte 5: Отримання пропсів
  let { data } = $props<{ data: { event: SystemEvent | null } }>();

  // Svelte 5: Похідний стан
  const event = $derived(data.event);

  function formatDate(dateStr: string) {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleString("uk-UA");
  }

  // Функція для красивого форматування JSON
  const formattedJson = $derived.by(() => {
    if (!event?.payload) return "";
    try {
      // Якщо це вже об'єкт — форматуємо, якщо рядок — парсимо і форматуємо
      const obj =
        typeof event.payload === "string"
          ? JSON.parse(event.payload)
          : event.payload;
      return JSON.stringify(obj, null, 2);
    } catch (e) {
      // Якщо це не JSON, повертаємо як є
      return event.payload;
    }
  });
</script>

<div class="container mx-auto py-6 space-y-6">
  <div class="flex items-center gap-4">
    <Button
      variant="outline"
      size="icon"
      href="/events?tab=system"
      class="rounded-full shadow-sm h-9 w-9 shrink-0"
    >
      <ArrowLeft class="h-4 w-4" />
    </Button>
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-foreground">
        Деталі події
      </h1>
      <p class="text-sm text-muted-foreground">
        Технічна інформація та корисне навантаження
      </p>
    </div>
  </div>

  {#if event}
    <div class="grid gap-6 lg:grid-cols-3">
      <!-- Sidebar Info -->
      <div class="lg:col-span-1 space-y-6">
        <Card.Root class="shadow-sm border-border gap-0 py-0">
          <Card.Header class="bg-muted/50 border-b pt-4">
            <Card.Title class="flex items-center gap-2 text-lg">
              <Activity class="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
              Метадані
            </Card.Title>
          </Card.Header>
          <Card.Content class="space-y-0 text-sm p-0">
            <div
              class="flex items-center justify-between p-4 border-b border-border last:border-0 hover:bg-muted/50 transition-colors"
            >
              <div class="flex items-center gap-2 text-muted-foreground">
                <Database class="h-4 w-4" />
                <span>ID події</span>
              </div>
              <span class="font-mono font-bold text-foreground">{event.ID}</span>
            </div>

            <div
              class="flex items-center justify-between p-4 border-b border-border last:border-0 hover:bg-muted/50 transition-colors"
            >
              <div class="flex items-center gap-2 text-muted-foreground">
                <Clock class="h-4 w-4" />
                <span>Час</span>
              </div>
              <span>{formatDate(event.timestamp)}</span>
            </div>

            <div
              class="flex items-center justify-between p-4 border-b border-border last:border-0 hover:bg-muted/50 transition-colors"
            >
              <div class="flex items-center gap-2 text-muted-foreground">
                <FileQuestionMark class="h-4 w-4" />
                <span>Тип</span>
              </div>
              <span
                class="inline-flex items-center px-2 py-1 rounded-md bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 text-xs font-bold font-mono"
              >
                {event.type}
              </span>
            </div>

          </Card.Content>
        </Card.Root>
      </div>

      <!-- Main Content (JSON) -->
      <div class="lg:col-span-2">
        <Card.Root class="h-full shadow-sm border-border overflow-hidden py-0 gap-0">
          <Card.Header
            class="bg-muted/50 border-b pb-4 pt-4 flex flex-row items-center justify-between space-y-0"
          >
            <div class="flex items-center gap-2">
              <FileQuestionMark class="h-5 w-5 text-blue-600 dark:text-blue-400" />
              <Card.Title>Корисне навантаження</Card.Title>
            </div>
            <div
              class="text-[10px] font-black tracking-wider text-muted-foreground uppercase bg-muted/50 px-2 py-1 rounded border border-border"
            >
              JSON Payload
            </div>
          </Card.Header>
          <Card.Content class="p-0">
            <div class="bg-[#0d1117] p-4 sm:p-6 overflow-x-auto min-h-[300px]">
              <pre
                class="text-xs sm:text-sm font-mono text-slate-300 leading-relaxed"><code
                  >{formattedJson}</code
                ></pre>
            </div>
          </Card.Content>
        </Card.Root>
      </div>
    </div>
  {:else}
    <Card.Root
      class="flex flex-col items-center justify-center h-[50vh] border-2 border-dashed border-border bg-muted/20 shadow-none"
    >
      <Card.Content
        class="flex flex-col items-center space-y-4 pt-6 text-center"
      >
        <div class="p-4 bg-background rounded-full shadow-sm ring-1 ring-border">
          <Database class="h-10 w-10 text-muted-foreground/30" />
        </div>
        <div class="space-y-1">
          <h2 class="text-2xl font-bold text-foreground">Подію не знайдено</h2>
          <p class="text-muted-foreground text-sm">
            Можливо, запис було видалено або ID вказано невірно
          </p>
        </div>
        <Button
          href="/events?tab=system"
          variant="default"
          class="rounded-full px-6"
        >
          Повернутися до списку
        </Button>
      </Card.Content>
    </Card.Root>
  {/if}
</div>
