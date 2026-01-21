<script lang="ts">
  import { enhance } from "$app/forms";
  import * as Table from "$lib/components/ui/table";
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Checkbox } from "$lib/components/ui/checkbox";
  import { Switch } from "$lib/components/ui/switch";
  import type { PageData } from "./$types";
  import type { APIKey, Permission } from "$lib/server/auth-client";
  import Pencil from "@lucide/svelte/icons/pencil";
  import Trash2 from "@lucide/svelte/icons/trash-2";
  import Shield from "@lucide/svelte/icons/shield";
  import Plus from "@lucide/svelte/icons/plus";
  import Copy from "@lucide/svelte/icons/copy";
  import { toast } from "svelte-sonner";

  let { data }: { data: PageData } = $props();

  let isCreateOpen = $state(false);
  let isEditOpen = $state(false);
  let isPermsOpen = $state(false);
  let isDeleteOpen = $state(false);
  let isSecretOpen = $state(false);
  let isSystemWorkerUpdateAllowed = $state(false);
  let isSystemWorkerDeleteAllowed = $state(false);

  let currentKey: APIKey | null = $state(null);
  let generatedSecret: string = $state("");
  let selectedPermissions: string[] = $state([]);
  let permSearch = $state("");

  let filteredPermissions = $derived(
    data.permissions.filter(
      (p: Permission) =>
        p.id.toLowerCase().includes(permSearch.toLowerCase()) ||
        p.description.toLowerCase().includes(permSearch.toLowerCase()),
    ),
  );

  function openCreate() {
    currentKey = null;
    selectedPermissions = [];
    permSearch = "";
    isCreateOpen = true;
  }

  function openEdit(key: APIKey) {
    currentKey = key;
    isEditOpen = true;
  }

  function openDelete(key: APIKey) {
    currentKey = key;
    isDeleteOpen = true;
  }

  function openPerms(key: APIKey) {
    currentKey = key;
    if (key.permissions) {
      selectedPermissions = key.permissions.map((p) => p.id);
    } else {
      selectedPermissions = [];
    }
    permSearch = "";
    isPermsOpen = true;
  }

  function togglePermission(id: string) {
    if (selectedPermissions.includes(id)) {
      selectedPermissions = selectedPermissions.filter((p) => p !== id);
    } else {
      selectedPermissions = [...selectedPermissions, id];
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    toast.success("Скопійовано в буфер обміну");
  }

  function can(permission: string): boolean {
    return data.user?.permissions?.includes(permission) ?? false;
  }
</script>

<div class="p-6 space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">API Ключі</h1>
      <p class="text-muted-foreground">
        Керування ключами доступу для зовнішніх інтеграцій.
      </p>
    </div>
    {#if can("create:keys")}
      <Button onclick={openCreate}>
        <Plus class="mr-2 h-4 w-4" />
        Створити ключ
      </Button>
    {/if}
  </div>

  <div class="border rounded-md">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>ID</Table.Head>
          <Table.Head>Назва (Власник)</Table.Head>
          <Table.Head>Статус</Table.Head>
          <Table.Head>Створено</Table.Head>
          <Table.Head class="text-right">Дії</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each data.keys as key (key.id)}
          <Table.Row>
            <Table.Cell>{key.id}</Table.Cell>
            <Table.Cell class="font-medium">{key.owner_name}</Table.Cell>
            <Table.Cell>
              {#if key.is_active}
                <span
                  class="inline-flex items-center rounded-full bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20"
                  >Активний</span
                >
              {:else}
                <span
                  class="inline-flex items-center rounded-full bg-red-50 px-2 py-1 text-xs font-medium text-red-700 ring-1 ring-inset ring-red-600/10"
                  >Неактивний</span
                >
              {/if}
            </Table.Cell>
            <Table.Cell>{new Date(key.created_at).toLocaleString()}</Table.Cell>
            <Table.Cell class="text-right space-x-2">
              {#if can("update:keys")}
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => openPerms(key)}
                  title="Права доступу"
                >
                  <Shield class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => openEdit(key)}
                  title="Редагувати"
                >
                  <Pencil class="h-4 w-4" />
                </Button>
              {/if}
              {#if can("delete:keys")}
                <!-- Don't allow to delete default system worker key -->
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => openDelete(key)}
                  class="text-destructive hover:text-destructive"
                  title="Видалити"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              {/if}
            </Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
  </div>

  <!-- Create Key Dialog -->
  <Dialog.Root bind:open={isCreateOpen}>
    <Dialog.Content class="max-w-2xl">
      <Dialog.Header>
        <Dialog.Title>Створити новий API ключ</Dialog.Title>
      </Dialog.Header>
      <form
        action="?/create"
        method="POST"
        use:enhance={({ formData }) => {
          toast.loading("Створення ключа...");

          return async ({ result, update }) => {
            if (result.type === "success" && result.data?.newKey) {
              const newKeyData = result.data.newKey as { api_key: string };
              generatedSecret = newKeyData.api_key;
              isCreateOpen = false;
              isSecretOpen = true; // Show secret to user
              toast.success("Ключ створено");
              await update();
            } else {
              toast.error("Не вдалося створити ключ");
            }
          };
        }}
      >
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <Label for="name">Назва (Власник)</Label>
            <Input id="name" name="name" required placeholder="IoT Sensor 1" />
          </div>

          <!-- Permissions Selection for New Key -->
          <div class="grid gap-2">
            <Label>Початкові права</Label>
            <Input
              placeholder="Пошук прав..."
              bind:value={permSearch}
              class="mb-2"
            />
            <div class="h-[200px] overflow-y-auto border rounded-md p-4">
              <div class="grid grid-cols-2 gap-4">
                {#each filteredPermissions as perm}
                  <div class="flex items-start space-x-2">
                    <Checkbox
                      id="new-perm-{perm.id}"
                      value={perm.id}
                      checked={selectedPermissions.includes(perm.id)}
                      onCheckedChange={() => togglePermission(perm.id)}
                    />
                    <Label
                      for="new-perm-{perm.id}"
                      class="text-sm font-medium leading-none"
                    >
                      <div class="flex flex-col">
                        {perm.name}
                        <span class="text-xs text-muted-foreground"
                          >{perm.id}</span
                        >
                      </div>
                    </Label>
                  </div>
                {/each}
              </div>
            </div>
            <!-- Hidden inputs for permissions -->
            {#each selectedPermissions as pId}
              <input type="hidden" name="permissions" value={pId} />
            {/each}
          </div>
        </div>
        <Dialog.Footer>
          <Button type="submit">Створити</Button>
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  </Dialog.Root>

  <!-- Secret Key Display Dialog -->
  <Dialog.Root bind:open={isSecretOpen}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>API Ключ Створено</Dialog.Title>
        <Dialog.Description>
          Збережіть цей ключ зараз. Ви не зможете побачити його знову.
        </Dialog.Description>
      </Dialog.Header>
      <div class="flex items-center space-x-2 mt-4">
        <Input
          readonly
          value={generatedSecret}
          class="font-mono text-center bg-muted"
        />
        <Button
          variant="outline"
          size="icon"
          onclick={() => copyToClipboard(generatedSecret)}
        >
          <Copy class="h-4 w-4" />
        </Button>
      </div>
      <Dialog.Footer class="mt-4">
        <Button onclick={() => (isSecretOpen = false)}>Закрити</Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>

  <!-- Edit Key Dialog -->
  <Dialog.Root bind:open={isEditOpen}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>Редагувати ключ</Dialog.Title>
      </Dialog.Header>
      <form
        action="?/update"
        method="POST"
        use:enhance={() => {
          toast.loading("Оновлення ключа...");
          return async ({ result, update }) => {
            if (result.type === "success") {
              isEditOpen = false;
              toast.success("Ключ оновлено");
              await update();
            } else {
              toast.error("Не вдалося оновити ключ");
            }
          };
        }}
      >
        <input type="hidden" name="id" value={currentKey?.id} />
        <div class="space-y-4 py-4">
          <div class="grid gap-2">
            <Label for="edit-owner">Назва (Власник)</Label>
            <Input
              id="edit-owner"
              name="owner_name"
              bind:value={currentKey!.owner_name}
              required
            />
          </div>
          <div class="flex items-center space-x-2">
            <Switch
              id="edit-active"
              name="is_active_toggle"
              checked={currentKey!.is_active}
              onCheckedChange={(v) => (currentKey!.is_active = v)}
            />
            <input
              type="hidden"
              name="is_active"
              value={currentKey!.is_active}
            />
            <Label for="edit-active">Активний</Label>
          </div>
        </div>
        <Dialog.Footer>
          <Button type="submit">Зберегти</Button>
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  </Dialog.Root>

  <!-- Permissions Dialog -->
  <Dialog.Root bind:open={isPermsOpen}>
    <Dialog.Content class="max-w-2xl">
      <Dialog.Header>
        <Dialog.Title>Налаштування прав: {currentKey?.owner_name}</Dialog.Title>
      </Dialog.Header>
      {#if currentKey?.id === 1}
        Ви намагаєтеся змінити права для системного ключа!
        <Button onclick={() => (isSystemWorkerUpdateAllowed = true)}
          >Дозволити змінювати</Button
        >
      {/if}
      {#if currentKey?.id !== 1 || isSystemWorkerUpdateAllowed}
        <form
          action="?/assignPermissions"
          method="POST"
          use:enhance={() => {
            toast.loading("Збереження прав...");
            return async ({ result, update }) => {
              if (result.type === "success") {
                isPermsOpen = false;
                toast.success("Права оновлено");
                await update();
              } else {
                toast.error("Не вдалося оновити права");
              }
            };
          }}
        >
          <input type="hidden" name="id" value={currentKey?.id} />
          {#each selectedPermissions as pId}
            <input type="hidden" name="permissions" value={pId} />
          {/each}
          <div class="px-1 py-4">
            <Input
              placeholder="Пошук прав..."
              bind:value={permSearch}
              class="mb-4"
            />
            <div class="py-4 h-[50vh] overflow-y-auto">
              <div class="grid grid-cols-2 gap-4">
                {#each filteredPermissions as perm}
                  <div class="flex items-start space-x-2">
                    <Checkbox
                      id="perm-{perm.id}"
                      value={perm.id}
                      checked={selectedPermissions.includes(perm.id)}
                      onCheckedChange={() => togglePermission(perm.id)}
                    />
                    <div class="grid gap-1.5 leading-none">
                      <Label
                        for="perm-{perm.id}"
                        class="text-sm font-medium leading-none"
                      >
                        <div class="flex flex-col">
                          {perm.name}
                          <span class="text-xs text-muted-foreground"
                            >{perm.id}</span
                          >
                        </div>
                      </Label>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          </div>
          <Dialog.Footer>
            <Button type="submit">Зберегти права</Button>
          </Dialog.Footer>
        </form>
      {/if}
    </Dialog.Content>
  </Dialog.Root>

  <!-- Delete Confirmation Dialog -->
  <Dialog.Root bind:open={isDeleteOpen}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>Видалити ключ?</Dialog.Title>
        <Dialog.Description>
          Ви впевнені, що хочете видалити ключ <strong
            >{currentKey?.owner_name}</strong
          >? Цю дію неможливо скасувати.
          {#if currentKey?.id === 1}
            <br />
            <span class="text-red-500"
              >Ви намагаєтесь видалити системний ключ!</span
            >
            <Button
              variant="destructive"
              onclick={() => (isSystemWorkerDeleteAllowed = true)}
              >Дозволити видалення</Button
            >
          {/if}
        </Dialog.Description>
      </Dialog.Header>
      <Dialog.Footer>
        <Button variant="outline" onclick={() => (isDeleteOpen = false)}
          >Скасувати</Button
        >
        <form
          action="?/delete"
          method="POST"
          use:enhance={() => {
            isDeleteOpen = false;
            toast.info("Видалення ключа...");
            return async ({ result, update }) => {
              if (result.type === "error" || result.type === "failure") {
                toast.error("Не вдалося видалити ключ");
              } else {
                toast.success("Ключ видалено");
                await update();
              }
            };
          }}
        >
          <input type="hidden" name="id" value={currentKey?.id} />
          <Button
            type="submit"
            variant="destructive"
            disabled={currentKey?.id === 1 && !isSystemWorkerDeleteAllowed}
            >Видалити</Button
          >
        </form>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>
</div>
