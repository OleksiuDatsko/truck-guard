<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import { enhance } from "$app/forms";
  import { toast } from "svelte-sonner";

  export let data;
  export let form;

  $: user = data.user;
</script>

<div class="p-6 space-y-6">
  <div>
    <h1 class="text-3xl font-bold tracking-tight">
      Редагування користувача: {user.username}
    </h1>
    <p class="text-sm text-muted-foreground">
      Змініть дані користувача та роль.
    </p>
  </div>

  {#if form?.error}
    <div class="rounded-md bg-destructive/15 p-3 text-sm text-destructive">
      {form.error}
    </div>
  {/if}

  <form
    method="POST"
    class="space-y-8"
    use:enhance={() => {
      toast.loading("Оновлення даних...");
      return async ({ result, update }) => {
        if (result.type === "redirect") {
          toast.success("Дані користувача оновлено");
          await update();
        } else if (result.type === "failure" || result.type === "error") {
          toast.error("Не вдалося оновити дані");
          await update();
        } else {
          await update();
        }
      };
    }}
  >
    <div class="grid gap-4 md:grid-cols-2">
      <div class="space-y-2">
        <Label for="username">Логін (Username)</Label>
        <Input id="username" value={user.username} disabled class="bg-muted" />
        <p class="text-[0.8rem] text-muted-foreground">
          Логін не можна змінити.
        </p>
      </div>
      <div class="space-y-2">
        <Label for="role_id">Роль *</Label>
        <select
          id="role_id"
          name="role_id"
          value={user.role_id}
          class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          required
        >
          {#if data.roles}
            {#each data.roles as role}
              <option value={role.id}>{role.name} ({role.description})</option>
            {/each}
          {/if}
        </select>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <div class="space-y-2">
        <Label for="last_name">Прізвище</Label>
        <Input
          id="last_name"
          name="last_name"
          value={user.profile.last_name || ""}
        />
      </div>
      <div class="space-y-2">
        <Label for="first_name">Ім'я</Label>
        <Input
          id="first_name"
          name="first_name"
          value={user.profile.first_name || ""}
        />
      </div>
      <div class="space-y-2">
        <Label for="third_name">По батькові</Label>
        <Input
          id="third_name"
          name="third_name"
          value={user.profile.third_name || ""}
        />
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      <div class="space-y-2">
        <Label for="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          value={user.profile.email || ""}
        />
      </div>
      <div class="space-y-2">
        <Label for="phone_number">Телефон</Label>
        <Input
          id="phone_number"
          name="phone_number"
          value={user.profile.phone_number || ""}
        />
      </div>
    </div>

    <div class="space-y-2">
      <Label for="notes">Примітки</Label>
      <Textarea
        id="notes"
        name="notes"
        value={user.profile.notes || ""}
        placeholder="Додаткова інформація..."
      />
    </div>

    <div class="flex justify-end gap-4">
      <Button variant="outline" href="/admin/users">Скасувати</Button>
      <Button type="submit">Зберегти зміни</Button>
    </div>
  </form>
</div>
