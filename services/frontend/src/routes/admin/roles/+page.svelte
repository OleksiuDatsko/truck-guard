<script lang="ts">
  import { enhance } from "$app/forms";
  import * as Table from "$lib/components/ui/table";
  import * as Dialog from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Checkbox } from "$lib/components/ui/checkbox";
  import type { PageData } from "./$types";
  import type { Role, Permission } from "$lib/server/auth-client";
  import Pencil from "@lucide/svelte/icons/pencil";
  import Trash2 from "@lucide/svelte/icons/trash-2";
  import Shield from "@lucide/svelte/icons/shield";
  import Plus from "@lucide/svelte/icons/plus";
  import { toast } from "svelte-sonner";
  import { can } from "$lib/auth";

  let { data }: { data: PageData } = $props();

  let isCreateOpen = $state(false);
  let isEditOpen = $state(false);
  let isPermsOpen = $state(false);
  let isDeleteOpen = $state(false);

  let currentRole: Role | null = $state(null);
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
    currentRole = null;
    isCreateOpen = true;
  }

  function openEdit(role: Role) {
    currentRole = role;
    isEditOpen = true;
  }

  function openDelete(role: Role) {
    currentRole = role;
    isDeleteOpen = true;
  }

  function openPerms(role: Role) {
    currentRole = role;
    if (role.permissions) {
      selectedPermissions = role.permissions.map((p) => p.id);
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
</script>

<div class="p-6 space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Ролі та права</h1>
      <p class="text-muted-foreground">
        Керування ролями користувачів та їх доступом.
      </p>
    </div>
    {#if can(data.user,"create:roles")}
      <Button onclick={openCreate}>
        <Plus class="mr-2 h-4 w-4" />
        Створити роль
      </Button>
    {/if}
  </div>

  <div class="border rounded-md">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>Назва</Table.Head>
          <Table.Head>Опис</Table.Head>
          <Table.Head class="text-right">Дії</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each data.roles as role (role.id)}
          <Table.Row>
            <Table.Cell class="font-medium">{role.name}</Table.Cell>
            <Table.Cell>{role.description}</Table.Cell>
            <Table.Cell class="text-right space-x-2">
              {#if can(data.user,"update:roles")}
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => openPerms(role)}
                  title="Права доступу"
                >
                  <Shield class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => openEdit(role)}
                  title="Редагувати"
                >
                  <Pencil class="h-4 w-4" />
                </Button>
              {/if}
              {#if can(data.user,"delete:roles")}
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => openDelete(role)}
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

  <!-- Create Role Dialog -->
  <Dialog.Root bind:open={isCreateOpen}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>Створити нову роль</Dialog.Title>
      </Dialog.Header>
      <form
        action="?/create"
        method="POST"
        use:enhance={({ formData }) => {
          const name = formData.get("name") as string;
          const description = formData.get("description") as string;

          // Optimistic UI
          const newRole: Role = {
            id: Date.now(), // Temporary id
            name,
            description,
            permissions: [],
          };

          const originalRoles = data.roles;
          data.roles = [...data.roles, newRole];
          isCreateOpen = false;
          toast.loading("Створення ролі...");

          return async ({ result, update }) => {
            if (result.type === "error" || result.type === "failure") {
              data.roles = originalRoles;
              toast.error("Не вдалося створити роль");
              isCreateOpen = true;
            } else {
              toast.success("Роль створено");
              await update();
            }
          };
        }}
      >
        <div class="space-y-4 py-4">
          <div class="grid gap-2">
            <Label for="name">Назва</Label>
            <Input id="name" name="name" required placeholder="admin" />
          </div>
          <div class="grid gap-2">
            <Label for="desc">Опис</Label>
            <Input
              id="desc"
              name="description"
              placeholder="Адміністратор системи"
            />
          </div>
        </div>
        <Dialog.Footer>
          <Button type="submit">Створити</Button>
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  </Dialog.Root>

  <!-- Edit Role Dialog -->
  <Dialog.Root bind:open={isEditOpen}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>Редагувати роль</Dialog.Title>
      </Dialog.Header>
      <form
        action="?/update"
        method="POST"
        use:enhance={() => {
          toast.loading("Оновлення ролі...");
          return async ({ result, update }) => {
            if (result.type === "success") {
              isEditOpen = false;
              toast.success("Роль оновлено");
              await update();
            } else {
              toast.error("Не вдалося оновити роль");
            }
          };
        }}
      >
        <input type="hidden" name="id" value={currentRole?.id} />
        <div class="space-y-4 py-4">
          <div class="grid gap-2">
            <Label for="edit-name">Назва</Label>
            <Input
              id="edit-name"
              name="name"
              bind:value={currentRole!.name}
              required
            />
          </div>
          <div class="grid gap-2">
            <Label for="edit-desc">Опис</Label>
            <Input
              id="edit-desc"
              name="description"
              bind:value={currentRole!.description}
            />
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
        <Dialog.Title
          >Налаштування прав доступу: {currentRole?.name}</Dialog.Title
        >
      </Dialog.Header>
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
        <input type="hidden" name="id" value={currentRole?.id} />
        {#each selectedPermissions as pId}
          <input type="hidden" name="permissions" value={pId} />
        {/each}
        <div class="px-1 py-4">
          <Input
            placeholder="Пошук прав..."
            bind:value={permSearch}
            class="mb-4"
          />
        </div>
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
                    class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                  >
                    <div class="flex flex-col">
                      {perm.name}
                      <span class="text-xs text-muted-foreground">
                        {perm.id}
                      </span>
                    </div>
                  </Label>
                </div>
              </div>
            {/each}
          </div>
        </div>
        <Dialog.Footer>
          <Button type="submit">Зберегти права</Button>
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  </Dialog.Root>

  <!-- Delete Confirmation Dialog -->
  <Dialog.Root bind:open={isDeleteOpen}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>Видалити роль?</Dialog.Title>
        <Dialog.Description>
          Ви впевнені, що хочете видалити роль <strong
            >{currentRole?.name}</strong
          >? Цю дію неможливо скасувати.
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
            const roleId = currentRole?.id;
            const originalRoles = data.roles;
            if (roleId) {
              data.roles = data.roles.filter((r: Role) => r.id !== roleId);
            }
            isDeleteOpen = false;
            toast.info("Видалення ролі...");
            return async ({ result, update }) => {
              if (result.type === "error" || result.type === "failure") {
                data.roles = originalRoles;
                toast.error("Не вдалося видалити роль");
              } else {
                toast.success("Роль видалено");
                await update();
              }
            };
          }}
        >
          <input type="hidden" name="id" value={currentRole?.id} />
          <Button type="submit" variant="destructive">Видалити</Button>
        </form>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>
</div>
