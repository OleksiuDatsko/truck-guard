<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import * as Table from "$lib/components/ui/table";
  import {
    ArrowLeft,
    Clock,
    MapPin,
    Hash,
    FileText,
    Camera,
    Scale,
    Image as ImageIcon,
    UserCog,
    Cpu,
    Database,
    Info,
  } from "@lucide/svelte";
  import type { GateEvent } from "$lib/types/events";

  let { data } = $props<{ data: { event: GateEvent | null } }>();

  const event = $derived(data.event);

  function formatDate(dateStr: string) {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleString("uk-UA");
  }
</script>

<div class="container mx-auto py-6 space-y-6">
  <div class="flex items-center gap-4">
    <Button
      variant="outline"
      size="icon"
      href="/events?tab=gate"
      class="rounded-full shadow-sm h-9 w-9 shrink-0"
    >
      <ArrowLeft class="h-4 w-4" />
    </Button>
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-foreground">
        Деталі події на КПП
      </h1>
      <p class="text-sm text-muted-foreground">
        Інформація про проїзд, розпізнані номери та зважування
      </p>
    </div>
  </div>

  {#if event}
    <div class="grid gap-6 md:grid-cols-3">
      <Card.Root class="md:col-span-1 shadow-sm border-border h-fit gap-0 py-0">
        <Card.Header class="border-b bg-muted/50 pt-6 ">
          <Card.Title class="flex items-center gap-2 text-lg">
            <Info class="h-5 w-5 text-emerald-600 dark:text-emerald-500" />
            Загальна інформація
          </Card.Title>
        </Card.Header>
        <Card.Content class="space-y-0 text-sm p-0">
          <div
            class="flex items-center justify-between p-4 border-b border-border last:border-0 hover:bg-muted/50 transition-colors"
          >
            <div class="flex items-center gap-2 text-muted-foreground">
              <Hash class="h-4 w-4" />
              <span>ID події</span>
            </div>
            <span class="font-mono font-bold text-foreground">{event.ID}</span>
          </div>

          <div
            class="flex items-center justify-between p-4 border-b border-border last:border-0 hover:bg-muted/50 transition-colors"
          >
            <div class="flex items-center gap-2 text-muted-foreground">
              <Clock class="h-4 w-4" />
              <span>Час реєстрації</span>
            </div>
            <span class="font-medium">{formatDate(event.timestamp)}</span>
          </div>

          <div
            class="flex items-center justify-between p-4 border-b border-border last:border-0 hover:bg-muted/50 transition-colors"
          >
            <div class="flex items-center gap-2 text-muted-foreground">
              <MapPin class="h-4 w-4" />
              <span>Локація (КПП)</span>
            </div>
            <span
              class="inline-flex items-center px-2 py-1 rounded-md bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-400 text-xs font-bold border border-emerald-100 dark:border-emerald-900/50"
            >
              {event.gate_id}
            </span>
          </div>

          <div
            class="flex items-center justify-between p-4 border-b border-border last:border-0 hover:bg-muted/50 transition-colors"
          >
            <div class="flex items-center gap-2 text-muted-foreground">
              <FileText class="h-4 w-4" />
              <span>Перепустка</span>
            </div>
            {#if event.permit_id}
              <Button
                variant="link"
                class="p-0 h-auto font-bold text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"
                href={`/permits/${event.permit_id}`}
              >
                #{event.permit_id}
              </Button>
            {:else}
              <span
                class="text-xs text-muted-foreground italic bg-muted/50 px-2 py-0.5 rounded border border-dashed border-border"
              >
                Не призначено
              </span>
            {/if}
          </div>
        </Card.Content>
      </Card.Root>

      <div class="md:col-span-2 space-y-6">
        <Card.Root class="shadow-sm border-border overflow-hidden gap-0 py-0">
          <Card.Header
            class="flex flex-row items-center justify-between border-b pt-4 bg-muted/50"
          >
            <div class="space-y-1">
              <Card.Title class="flex items-center gap-2 text-lg">
                <Camera class="h-5 w-5 text-indigo-600 dark:text-indigo-500" />
                Розпізнані номери
              </Card.Title>
              <Card.Description
                >Події з камер ANPR під час цього проїзду</Card.Description
              >
            </div>
            <div
              class="bg-indigo-100 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 px-3 py-1 rounded-full text-xs font-black shadow-sm"
            >
              {event.plate_events?.length || 0}
            </div>
          </Card.Header>
          <Card.Content class="p-0">
            <Table.Root>
              <Table.Header>
                <Table.Row>
                  <Table.Head class="w-[80px] text-center">ID</Table.Head>
                  <Table.Head>Камера</Table.Head>
                  <Table.Head>Номер</Table.Head>
                  <Table.Head>Метод</Table.Head>
                  <Table.Head class="text-right pr-6">Дія</Table.Head>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {#if event.plate_events && event.plate_events.length > 0}
                  {#each event.plate_events as plateEvent}
                    <Table.Row
                      class="hover:bg-muted/30 transition-colors border-b last:border-0 border-border"
                    >
                      <Table.Cell class="pl-6 font-mono">
                        <Button variant="link" href={`/events/plate/${plateEvent.ID}`}>
                          #{plateEvent.ID}
                        </Button>
                      </Table.Cell>
                      <Table.Cell>
                        <div class="flex flex-col">
                          <span class="font-medium text-sm text-foreground">
                            {plateEvent.camera_name || plateEvent.camera_id}
                          </span>
                          <span class="text-[10px] text-muted-foreground">
                            {formatDate(plateEvent.timestamp)}
                          </span>
                        </div>
                      </Table.Cell>
                      <Table.Cell>
                        <div
                          class="inline-flex items-center border border-border rounded-sm bg-white dark:bg-slate-950 px-2 py-1 shadow-sm select-none"
                        >
                          <span
                            class="font-bold text-slate-900 dark:text-slate-100 tracking-wider font-mono text-base uppercase leading-none"
                          >
                            {plateEvent.plate}
                          </span>
                        </div>
                      </Table.Cell>
                      <Table.Cell>
                        {#if plateEvent.is_manual}
                          <span
                            class="inline-flex items-center gap-1.5 text-amber-700 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 px-2 py-1 rounded-full border border-amber-200 dark:border-amber-800 text-[10px] font-bold uppercase"
                          >
                            <UserCog class="h-3 w-3" /> Ручне
                          </span>
                        {:else}
                          <span
                            class="inline-flex items-center gap-1.5 text-indigo-700 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/30 px-2 py-1 rounded-full border border-indigo-200 dark:border-indigo-800 text-[10px] font-bold uppercase"
                          >
                            <Cpu class="h-3 w-3" /> ANPR
                          </span>
                        {/if}
                      </Table.Cell>
                      <Table.Cell class="text-right pr-6">
                        {#if plateEvent.image_key}
                          <Button
                            variant="outline"
                            size="sm"
                            class="h-8 gap-2 hover:bg-slate-900 dark:hover:bg-slate-100 hover:text-white dark:hover:text-slate-900 transition-all shadow-sm font-semibold"
                            href={`/api/images/${plateEvent.image_key}`}
                            target="_blank"
                          >
                            <ImageIcon class="h-3.5 w-3.5" />
                            Фото
                          </Button>
                        {/if}
                      </Table.Cell>
                    </Table.Row>
                  {/each}
                {:else}
                  <Table.Row>
                    <Table.Cell
                      colspan={5}
                      class="h-32 text-center text-muted-foreground italic"
                    >
                      <div
                        class="flex flex-col items-center justify-center gap-2"
                      >
                        <Camera class="h-8 w-8 opacity-20" />
                        <span>Події розпізнавання номерів відсутні</span>
                      </div>
                    </Table.Cell>
                  </Table.Row>
                {/if}
              </Table.Body>
            </Table.Root>
          </Card.Content>
        </Card.Root>

        <!-- Weight Events Card -->
        <Card.Root class="shadow-sm border-border overflow-hidden gap-0 py-0">
          <Card.Header
            class="flex flex-row items-center justify-between border-b bg-muted/50 pt-4"
          >
            <div class="space-y-1">
              <Card.Title class="flex items-center gap-2 text-lg">
                <Scale class="h-5 w-5 text-blue-600 dark:text-blue-500" />
                Зважування
              </Card.Title>
              <Card.Description>Показники вагової системи</Card.Description>
            </div>
            <div
              class="bg-blue-100 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 px-3 py-1 rounded-full text-xs font-black shadow-sm"
            >
              {event.weight_events?.length || 0}
            </div>
          </Card.Header>
          <Card.Content class="p-0">
            <Table.Root>
              <Table.Header>
                <Table.Row>
                  <Table.Head class="pl-6 w-[80px]">ID</Table.Head>
                  <Table.Head>Час</Table.Head>
                  <Table.Head>Обладнання</Table.Head>
                  <Table.Head class="text-right pr-6">Показник ваги</Table.Head>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {#if event.weight_events && event.weight_events.length > 0}
                  {#each event.weight_events as weightEvent}
                    <Table.Row
                      class="hover:bg-muted/30 transition-colors border-b last:border-0 border-border"
                    >
                      <Table.Cell
                        class="pl-6 font-mono text-xs text-muted-foreground"
                      >
                        {weightEvent.ID}
                      </Table.Cell>
                      <Table.Cell class="text-xs">
                        {formatDate(weightEvent.timestamp)}
                      </Table.Cell>
                      <Table.Cell>
                        <div class="flex items-center gap-2">
                          <div
                            class="p-1 bg-blue-50 dark:bg-blue-950/30 rounded"
                          >
                            <Scale
                              class="h-3 w-3 text-blue-600 dark:text-blue-400"
                            />
                          </div>
                          <span class="text-sm font-medium text-foreground">
                            {weightEvent.scale_id}
                          </span>
                        </div>
                      </Table.Cell>
                      <Table.Cell class="text-right pr-6">
                        <span
                          class="font-mono font-bold text-lg text-blue-700 dark:text-blue-400"
                        >
                          {weightEvent.weight}
                          <span
                            class="text-xs font-sans text-muted-foreground ml-1"
                            >кг</span
                          >
                        </span>
                      </Table.Cell>
                    </Table.Row>
                  {/each}
                {:else}
                  <Table.Row>
                    <Table.Cell
                      colspan={4}
                      class="h-32 text-center text-muted-foreground italic"
                    >
                      <div
                        class="flex flex-col items-center justify-center gap-2"
                      >
                        <Scale class="h-8 w-8 opacity-20" />
                        <span>Події зважування відсутні</span>
                      </div>
                    </Table.Cell>
                  </Table.Row>
                {/if}
              </Table.Body>
            </Table.Root>
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
        <div
          class="p-4 bg-background rounded-full shadow-sm ring-1 ring-border"
        >
          <Database class="h-10 w-10 text-muted-foreground/50" />
        </div>
        <div class="space-y-1">
          <h2 class="text-2xl font-bold text-foreground">Подію не знайдено</h2>
          <p class="text-muted-foreground text-sm">
            Можливо, запис було видалено або ID вказано невірно
          </p>
        </div>
        <Button
          href="/events?tab=gate"
          variant="default"
          class="rounded-full px-6"
        >
          Повернутися до списку
        </Button>
      </Card.Content>
    </Card.Root>
  {/if}
</div>

<style>
  :global(.font-mono) {
    font-variant-ligatures: none;
  }
</style>
