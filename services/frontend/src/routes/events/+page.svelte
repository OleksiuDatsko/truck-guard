<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import * as Tabs from "$lib/components/ui/tabs";
  import * as Table from "$lib/components/ui/table";
  import { Button } from "$lib/components/ui/button";
  import {
    RefreshCcw,
    Camera,
    Image as ImageIcon,
    UserCog,
    Cpu,
    Scale,
    DoorOpen,
    Settings,
    ChevronLeft,
    ChevronRight,
    FileText,
    Database,
    Calendar as CalendarIcon,
    Funnel,
    X,
    Activity,
  } from "@lucide/svelte";
  import { type ApiResponse } from "$lib/types/events";

  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import * as Collapsible from "$lib/components/ui/collapsible";
  import * as Popover from "$lib/components/ui/popover";
  import { RangeCalendar } from "$lib/components/ui/range-calendar";
  import {
    CalendarDate,
    DateFormatter,
    type DateValue,
    getLocalTimeZone,
  } from "@internationalized/date";
  import { cn } from "$lib/utils";

  const df = new DateFormatter("uk-UA", {
    dateStyle: "medium",
  });

  let { data } = $props<{
    data: {
      events: ApiResponse<any>;
      tab: string;
      page: number;
      limit: number;
    };
  }>();

  const activeTab = $derived(data.tab);
  const currentPage = $derived(data.page);
  const items = $derived(data.events.data || []);
  const metadata = $derived(
    data.events.metadata || {
      total_items: 0,
      total_pages: 0,
      current_page: 1,
      limit: 10,
    },
  );
  const totalPages = $derived(metadata.total_pages);

  let loading = $state(false);
  let isFiltersOpen = $state(false);

  // Parse initial date from URL or default to empty
  const initialFrom = page.url.searchParams.get("from");
  const initialTo = page.url.searchParams.get("to");

  let range = $state({
    start: initialFrom
      ? new CalendarDate(...parseIsoDate(initialFrom))
      : undefined,
    end: initialTo ? new CalendarDate(...parseIsoDate(initialTo)) : undefined,
  });

  let filters = $state({
    plate: page.url.searchParams.get("plate") || "",
    gate: page.url.searchParams.get("gate") || "",
    type: page.url.searchParams.get("type") || "",
  });

  function parseIsoDate(iso: string): [number, number, number] {
    const d = new Date(iso);
    return [d.getFullYear(), d.getMonth() + 1, d.getDate()];
  }

  function handleTabChange(value: string) {
    if (value !== activeTab) {
      loading = true;
      goto(`?tab=${value}&page=1`).then(() => (loading = false));
    }
  }

  function handlePageChange(newPage: number) {
    if (newPage >= 1 && newPage <= totalPages) {
      loading = true;
      goto(`?tab=${activeTab}&page=${newPage}`).then(() => (loading = false));
    }
  }

  function refresh() {
    loading = true;
    goto(page.url, { invalidateAll: true }).then(() => (loading = false));
  }

  function formatDate(dateStr: string) {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleString("uk-UA");
  }

  function applyFilters() {
    loading = true;
    const query = new URLSearchParams(page.url.searchParams);

    if (range.start) {
      // Set to beginning of day
      const start = range.start.toDate(getLocalTimeZone());
      query.set("from", start.toISOString());
    } else {
      query.delete("from");
    }

    if (range.end) {
      // Set to end of day
      const end = range.end.toDate(getLocalTimeZone());
      end.setHours(23, 59, 59, 999);
      query.set("to", end.toISOString());
    } else {
      query.delete("to");
    }

    if (filters.plate) query.set("plate", filters.plate);
    else query.delete("plate");
    if (filters.gate) query.set("gate", filters.gate);
    else query.delete("gate");
    if (filters.type) query.set("type", filters.type);
    else query.delete("type");

    // Reset page on filter change
    query.set("page", "1");

    goto(`?${query.toString()}`).then(() => (loading = false));
  }

  function resetFilters() {
    filters = { plate: "", gate: "", type: "" };
    range = { start: undefined, end: undefined };
    applyFilters();
  }
</script>

<div class="container mx-auto py-6 space-y-6">
  <div
    class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
  >
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-foreground">
        Моніторинг подій
      </h1>
      <p class="text-muted-foreground">
        Централізований перегляд активності системи
      </p>
    </div>
    <Button
      variant="outline"
      size="sm"
      onclick={refresh}
      disabled={loading}
      class="w-fit shadow-sm"
    >
      <RefreshCcw class="mr-2 h-4 w-4 {loading ? 'animate-spin' : ''}" />
      Оновити дані
    </Button>
  </div>

  <Tabs.Root value={activeTab} onValueChange={handleTabChange} class="w-full">
    <div class="flex items-center justify-between">
      <Tabs.List class="w-full justify-start grid-cols-4 lg:w-[640px] grid">
        <Tabs.Trigger value="plate"
          ><Camera class="mr-2 h-4 w-4" /> Номери</Tabs.Trigger
        >
        <Tabs.Trigger value="weight"
          ><Scale class="mr-2 h-4 w-4" /> Вага</Tabs.Trigger
        >
        <Tabs.Trigger value="gate"
          ><DoorOpen class="mr-2 h-4 w-4" /> Ворота</Tabs.Trigger
        >
        <Tabs.Trigger value="system"
          ><Settings class="mr-2 h-4 w-4" /> Система</Tabs.Trigger
        >
      </Tabs.List>
      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          class="shadow-sm gap-2"
          onclick={() => (isFiltersOpen = !isFiltersOpen)}
        >
          <Funnel class="h-4 w-4" />
          Фільтри
          {#if range.start || filters.plate || filters.gate || filters.type}
            <div
              class="h-1.5 w-1.5 rounded-full bg-emerald-500 absolute top-2 right-2"
            ></div>
          {/if}
        </Button>
      </div>
    </div>

    <Collapsible.Root
      open={isFiltersOpen}
      onOpenChange={(v) => (isFiltersOpen = v)}
      class="w-full"
    >
      <Collapsible.Content
        class="overflow-hidden data-[state=closed]:animate-collapsible-up data-[state=open]:animate-collapsible-down"
      >
        <div
          class="rounded-lg border bg-card p-4 shadow-sm grid gap-6 md:grid-cols-2 lg:grid-cols-4 items-end"
        >
          <!-- Date Range Picker -->
          <div class="grid gap-2 col-span-1 md:col-span-2 lg:col-span-2">
            <Label>Період</Label>
            <Popover.Root>
              <Popover.Trigger>
                {#snippet child({ props })}
                  <Button
                    variant="outline"
                    class={cn(
                      "w-full justify-start text-left font-normal",
                      !range && "text-muted-foreground",
                    )}
                    {...props}
                  >
                    <CalendarIcon class="mr-2 h-4 w-4 opacity-50" />
                    {#if range && range.start}
                      {#if range.end}
                        {df.format(range.start.toDate(getLocalTimeZone()))} - {df.format(
                          range.end.toDate(getLocalTimeZone()),
                        )}
                      {:else}
                        {df.format(range.start.toDate(getLocalTimeZone()))}
                      {/if}
                    {:else}
                      <span>Оберіть період</span>
                    {/if}
                  </Button>
                {/snippet}
              </Popover.Trigger>
              <Popover.Content class="w-auto p-0" align="start">
                <RangeCalendar
                  bind:value={range}
                  placeholder={range?.start}
                  numberOfMonths={2}
                />
              </Popover.Content>
            </Popover.Root>
          </div>

          <!-- Context Filters -->
          {#if activeTab === "plate"}
            <div class="grid gap-2">
              <Label for="plate">Номер авто</Label>
              <Input
                id="plate"
                placeholder="Пошук..."
                bind:value={filters.plate}
              />
            </div>
          {:else if activeTab === "gate"}
            <div class="grid gap-2">
              <Label for="gate">ID Гейта</Label>
              <Input id="gate" placeholder="ID..." bind:value={filters.gate} />
            </div>
          {:else if activeTab === "system"}
            <div class="grid gap-2">
              <Label for="type">Тип</Label>
              <Input
                id="type"
                placeholder="Тип події..."
                bind:value={filters.type}
              />
            </div>
          {/if}

          <!-- Actions -->
          <div class="flex items-center gap-2">
            <Button onclick={applyFilters} class="flex-1">Застосувати</Button>
            <Button
              variant="ghost"
              size="icon"
              onclick={resetFilters}
              title="Скинути"
            >
              <X class="h-4 w-4" />
            </Button>
          </div>
        </div>
      </Collapsible.Content>
    </Collapsible.Root>

    <div
      class="rounded-xl border bg-card shadow-md overflow-hidden transition-all"
    >
      <Table.Root>
        <Table.Header class="bg-muted/30">
          <Table.Row>
            <Table.Head class="w-[80px] text-center">ID</Table.Head>
            {#if activeTab === "plate"}
              <Table.Head>Час</Table.Head>
              <Table.Head>Джерело</Table.Head>
              <Table.Head>Номер автомобіля</Table.Head>
              <Table.Head>Метод</Table.Head>
              <Table.Head class="text-right">Фото</Table.Head>
            {:else if activeTab === "weight"}
              <Table.Head>Час</Table.Head>
              <Table.Head>Обладнання</Table.Head>
              <Table.Head class="text-right">Показник ваги</Table.Head>
            {:else if activeTab === "gate"}
              <Table.Head>Час</Table.Head>
              <Table.Head>Локація</Table.Head>
              <Table.Head>Документ</Table.Head>
              <Table.Head>Пов'язані дані</Table.Head>
            {:else if activeTab === "system"}
              <Table.Head>Час</Table.Head>
              <Table.Head>Категорія</Table.Head>
              <Table.Head>Дані</Table.Head>
            {/if}
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {#if items.length === 0}
            <Table.Row>
              <Table.Cell
                colspan={6}
                class="h-40 text-center text-muted-foreground"
              >
                <div class="flex flex-col items-center justify-center gap-2">
                  <Database class="h-8 w-8 opacity-20" />
                  <span class="text-lg font-medium italic"
                    >Дані не знайдено</span
                  >
                </div>
              </Table.Cell>
            </Table.Row>
          {:else}
            {#each items as item}
              <Table.Row
                class="hover:bg-muted/40 transition-colors border-b last:border-0"
              >
                <Table.Cell>
                  <Button
                    variant="link"
                    href={`/events/${activeTab}/${item.ID}`}
                  >
                    #{item.ID}
                  </Button>
                </Table.Cell>
                {#if activeTab === "plate"}
                  <Table.Cell class="whitespace-nowrap text-sm"
                    >{formatDate(item.timestamp)}</Table.Cell
                  >
                  <Table.Cell>
                    <div class="flex items-center gap-2">
                      <div class="p-1.5 bg-muted rounded">
                        <Camera class="h-3.5 w-3.5 text-muted-foreground" />
                      </div>
                      <span class="font-medium text-sm text-foreground"
                        >{item.camera_name || item.camera_id}</span
                      >
                    </div>
                  </Table.Cell>
                  <Table.Cell>
                    <div
                      class="inline-flex items-center border border-border rounded-sm bg-white dark:bg-slate-950 px-2 py-1 shadow-sm select-none"
                    >
                      <span
                        class="font-bold text-slate-900 dark:text-slate-50 tracking-[0.15em] font-mono text-base uppercase leading-none"
                      >
                        {item.plate}
                      </span>
                    </div>
                  </Table.Cell>
                  <Table.Cell>
                    {#if item.is_manual}
                      <span
                        class="inline-flex items-center gap-1.5 text-amber-700 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 px-2.5 py-1 rounded-full border border-amber-200 dark:border-amber-800 font-bold uppercase tracking-tight"
                      >
                        <UserCog class="h-3 w-3" /> Ручне
                      </span>
                    {:else}
                      <span
                        class="inline-flex items-center gap-1.5 text-indigo-700 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/30 px-2.5 py-1 rounded-full border border-indigo-200 dark:border-indigo-800 font-bold uppercase tracking-tight"
                      >
                        <Cpu class="h-3 w-3" /> ANPR
                      </span>
                    {/if}
                  </Table.Cell>
                  <Table.Cell class="text-right">
                    {#if item.image_key}
                      <Button
                        variant="outline"
                        size="sm"
                        class="h-8 gap-2 hover:bg-foreground hover:text-background transition-all shadow-sm"
                        href={`/api/images/${item.image_key}`}
                        target="_blank"
                      >
                        <ImageIcon class="h-3.5 w-3.5" /> Фото
                      </Button>
                    {/if}
                  </Table.Cell>
                {:else if activeTab === "weight"}
                  <Table.Cell class="whitespace-nowrap text-sm"
                    >{formatDate(item.timestamp)}</Table.Cell
                  >
                  <Table.Cell>
                    <div class="flex items-center gap-2 font-medium text-sm">
                      <div class="p-1.5 bg-blue-50 dark:bg-blue-950/30 rounded">
                        <Scale
                          class="h-3.5 w-3.5 text-blue-600 dark:text-blue-400"
                        />
                      </div>
                      <span class="text-foreground">{item.scale_id}</span>
                    </div>
                  </Table.Cell>
                  <Table.Cell class="text-right">
                    <span
                      class="font-mono font-bold text-xl text-blue-700 dark:text-blue-400"
                    >
                      {item.weight}
                      <span class="text-xs font-sans text-muted-foreground ml-1"
                        >кг</span
                      >
                    </span>
                  </Table.Cell>
                {:else if activeTab === "gate"}
                  <Table.Cell class="whitespace-nowrap text-sm"
                    >{formatDate(item.timestamp)}</Table.Cell
                  >
                  <Table.Cell>
                    <div class="flex items-center gap-2">
                      <div
                        class="p-1.5 bg-emerald-50 dark:bg-emerald-950/30 rounded"
                      >
                        <DoorOpen
                          class="h-3.5 w-3.5 text-emerald-600 dark:text-emerald-400"
                        />
                      </div>
                      <span class="font-semibold text-sm text-foreground"
                        >{item.gate_id}</span
                      >
                    </div>
                  </Table.Cell>
                  <Table.Cell>
                    {#if item.permit_id}
                      <Button
                        variant="link"
                        class="p-0 h-auto text-sm font-bold flex gap-1 items-center text-blue-600 dark:text-blue-400"
                        href={`/permits/${item.permit_id}`}
                      >
                        <FileText class="h-3 w-3" /> #{item.permit_id}
                      </Button>
                    {:else}
                      <span
                        class="text-muted-foreground text-xs bg-muted px-2 py-1 rounded border border-dashed border-border"
                        >Немає перепустки</span
                      >
                    {/if}
                  </Table.Cell>
                  <Table.Cell>
                    <div
                      class="flex items-center gap-4
                     font-medium"
                    >
                      <div
                        class="flex items-center gap-1.5 text-muted-foreground"
                      >
                        <Camera class="h-3 w-3" />
                        {item.plate_events?.length || 0}
                      </div>
                      <div
                        class="flex items-center gap-1.5 text-muted-foreground"
                      >
                        <Scale class="h-3 w-3" />
                        {item.weight_events?.length || 0}
                      </div>
                    </div>
                  </Table.Cell>
                {:else if activeTab === "system"}
                  <Table.Cell class="whitespace-nowrap text-sm"
                    >{formatDate(item.timestamp)}</Table.Cell
                  >
                  <Table.Cell>
                    <div
                      class="inline-flex items-center gap-1.5 bg-primary text-primary-foreground px-2 py-0.5 rounded font-mono font-bold"
                    >
                      <Activity class="h-3 w-3" />
                      {item.type}
                    </div>
                  </Table.Cell>
                  <Table.Cell>
                    <div
                      class="max-w-[55vw] truncate font-mono bg-muted p-1.5 rounded border border-border"
                      title={item.payload}
                    >
                      {item.payload}
                    </div>
                  </Table.Cell>
                {/if}
              </Table.Row>
            {/each}
          {/if}
        </Table.Body>
      </Table.Root>
    </div>

    <div
      class="flex flex-col sm:flex-row items-center justify-between gap-4 pt-6"
    >
      <div
        class="text-sm text-muted-foreground order-2 sm:order-1 flex items-center gap-2"
      >
        Сторінка
        <span
          class="font-bold text-foreground underline underline-offset-4 decoration-emerald-500/30"
          >{currentPage}</span
        >
        з <span class="font-bold text-foreground">{totalPages || 1}</span>
      </div>
      <div class="flex items-center gap-2 order-1 sm:order-2">
        <Button
          variant="outline"
          size="sm"
          class="shadow-sm h-9"
          onclick={() => handlePageChange(currentPage - 1)}
          disabled={currentPage <= 1 || loading}
        >
          <ChevronLeft class="h-4 w-4 mr-1.5" /> Назад
        </Button>
        <Button
          variant="outline"
          size="sm"
          class="shadow-sm h-9"
          onclick={() => handlePageChange(currentPage + 1)}
          disabled={currentPage >= totalPages || loading}
        >
          Далі <ChevronRight class="h-4 w-4 ml-1.5" />
        </Button>
      </div>
    </div>
  </Tabs.Root>
</div>
