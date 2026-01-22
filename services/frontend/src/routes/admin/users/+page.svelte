<script lang="ts">
  import * as Table from "$lib/components/ui/table";
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import { format } from "date-fns";
  import { can } from "$lib/auth";
  import { Pencil, Plus, Trash2 } from "@lucide/svelte";
  import * as Dialog from "$lib/components/ui/dialog";
  import { enhance } from "$app/forms";
  import { toast } from "svelte-sonner";

  let { data } = $props();

  let isDeleteOpen = $state(false);
  let userToDelete: any = $state(null);

  function openDelete(user: any) {
    userToDelete = user;
    isDeleteOpen = true;
  }
</script>

<div class="p-6 space-y-6 overflow-auto">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Користувачі</h1>
      <p class="text-muted-foreground">
        Список всіх користувачів системи з ролями та статусом
      </p>
    </div>
    {#if can(data.user, "create:users")}
      <Button href="/admin/users/create">
        <Plus class="mr-2 h-4 w-4" />
        Створити користувача
      </Button>
    {/if}
  </div>

  <div class="rounded-md border">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>Username</Table.Head>
          <Table.Head>Роль</Table.Head>
          <Table.Head>ПІБ</Table.Head>
          <Table.Head>Email</Table.Head>
          <Table.Head>Телефон</Table.Head>
          <Table.Head>Останній вхід</Table.Head>
          <Table.Head class="text-right">Дії</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each data.users as user (user.id)}
          <Table.Row>
            <Table.Cell>
              {user.username}
              {#if user.id === data.user.id}
                <Badge variant="outline">Ви</Badge>
              {/if}

            </Table.Cell>
            <Table.Cell>
              <Badge variant="outline">{user.role?.name || "No Role"}</Badge>
            </Table.Cell>
            <Table.Cell>
              {#if user.profile}
                {user.profile.last_name || ""}
                {user.profile.first_name || ""}
                {user.profile.third_name || ""}
              {:else}
                <span class="text-muted-foreground italic"
                  >Профіль не знайдено</span
                >
              {/if}
            </Table.Cell>
            <Table.Cell>
              {#if user.profile?.email}
                {user.profile.email}
              {:else}
                -
              {/if}
            </Table.Cell>
            <Table.Cell>
              {#if user.profile?.phone_number}
                {user.profile.phone_number}
              {:else}
                -
              {/if}
            </Table.Cell>
            <Table.Cell>
              {#if user.last_login}
                {format(new Date(user.last_login), "dd.MM.yyyy HH:mm")}
              {:else}
                <span class="text-muted-foreground">Ніколи</span>
              {/if}
            </Table.Cell>
            <Table.Cell class="text-right space-x-2">
              {#if can(data.user, "update:users") && user.profile && user.id != data.user.id}
                <Button
                  variant="ghost"
                  size="icon"
                  href={`/admin/users/${user.id}`}
                  title="Змінити"
                >
                  <Pencil class="h-4 w-4" />
                </Button>
              {/if}
              {#if can(data.user, "delete:users") && user.id != data.user.id}
                <Button
                  variant="ghost"
                  size="icon"
                  onclick={() => openDelete(user)}
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

  <Dialog.Root bind:open={isDeleteOpen}>
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>Видалити користувача?</Dialog.Title>
        <Dialog.Description>
          Ви впевнені, що хочете видалити користувача <strong
            >{userToDelete?.username}</strong
          >? Цю дію неможливо скасувати.
        </Dialog.Description>
      </Dialog.Header>
      <Dialog.Footer>
        <Button variant="outline" onclick={() => (isDeleteOpen = false)}>
          Скасувати
        </Button>
        <form
          action="?/delete"
          method="POST"
          use:enhance={() => {
            isDeleteOpen = false;
            toast.loading("Видалення користувача...");
            return async ({ result, update }) => {
              if (result.type === "success") {
                toast.success("Користувача видалено");
                await update();
              } else {
                toast.error("Не вдалося видалити користувача");
                // Optional: revert optimistic UI if we implemented it
              }
            };
          }}
        >
          <input type="hidden" name="id" value={userToDelete?.id} />
          <Button type="submit" variant="destructive">Видалити</Button>
        </form>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>
</div>
