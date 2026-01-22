<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Separator } from "$lib/components/ui/separator";
  import { ChevronLeft, UserPlus, ShieldCheck, Contact } from "@lucide/svelte";
  import { enhance } from "$app/forms";
  import { toast } from "svelte-sonner";

  let { data, form } = $props();
</script>

<div class="container mx-auto max-w-5xl py-10 px-4">
  <div
    class="mb-10 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
  >
    <div class="space-y-1">
      <h3 class="text-3xl font-bold tracking-tight text-foreground">
        Створити користувача
      </h3>
      <p class="text-muted-foreground">
        Додайте нового співробітника до системи та призначте йому права доступу.
      </p>
    </div>
    <Button variant="outline" href="/admin/users" class="w-fit gap-2">
      <ChevronLeft class="h-4 w-4" /> До списку
    </Button>
  </div>

  <form
    method="POST"
    class="space-y-12"
    use:enhance={() => {
      toast.loading("Створення користувача...");
      return async ({ result, update }) => {
        if (result.type === "redirect") {
          toast.success("Користувача створено успішно");
          await update();
        } else if (result.type === "failure" || result.type === "error") {
          toast.error("Не вдалося створити користувача");
          // Keep the form state if showing error, but update will reset if we don't handle it.
          // Default update behavior is fine for failure as it populates form.
          await update();
        } else {
          await update();
        }
      };
    }}
  >
    <div class="grid grid-cols-1 gap-8">
      <div class="space-y-1">
        <div class="flex items-center gap-2 font-semibold text-primary">
          <ShieldCheck class="h-5 w-5" />
          <span>Акаунт</span>
        </div>
        <p class="text-sm text-muted-foreground">
          Основні дані для авторизації та рівень доступу.
        </p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
        <div class="space-y-2">
          <Label
            for="username"
            class="after:content-['*'] after:ml-0.5 after:text-destructive text-sm font-medium"
            >Логін</Label
          >
          <Input
            id="username"
            name="username"
            placeholder="n.petrenko"
            required
            class="bg-background"
          />
        </div>
        <div class="space-y-2">
          <Label
            for="password"
            class="after:content-['*'] after:ml-0.5 after:text-destructive text-sm font-medium"
            >Пароль</Label
          >
          <Input
            id="password"
            name="password"
            type="password"
            required
            class="bg-background"
          />
        </div>
        <div class="sm:col-span-2 space-y-2">
          <Label
            for="role"
            class="after:content-['*'] after:ml-0.5 after:text-destructive text-sm font-medium"
            >Роль в системі</Label
          >
          <select
            id="role"
            name="role"
            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:ring-2 focus:ring-ring focus:ring-offset-2 transition-all outline-none"
            required
          >
            <option value="" disabled selected
              >Оберіть роль для користувача...</option
            >
            {#if data.roles}
              {#each data.roles as role}
                <option value={role.name}
                  >{role.name} — {role.description}</option
                >
              {/each}
            {/if}
          </select>
        </div>
      </div>
    </div>

    <Separator />

    <div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
      <div class="space-y-1">
        <div class="flex items-center gap-2 font-semibold text-primary">
          <UserPlus class="h-5 w-5" />
          <span>Персональні дані</span>
        </div>
        <p class="text-sm text-muted-foreground">
          Прізвище та ім'я користувача для відображення в системі.
        </p>
      </div>

      <div class="lg:col-span-2 grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div class="space-y-2">
          <Label for="last_name" class="text-sm font-medium">Прізвище</Label>
          <Input
            id="last_name"
            name="last_name"
            placeholder="Петренко"
            class="bg-background"
          />
        </div>
        <div class="space-y-2">
          <Label for="first_name" class="text-sm font-medium">Ім'я</Label>
          <Input
            id="first_name"
            name="first_name"
            placeholder="Микола"
            class="bg-background"
          />
        </div>
        <div class="space-y-2">
          <Label for="third_name" class="text-sm font-medium">По батькові</Label
          >
          <Input
            id="third_name"
            name="third_name"
            placeholder="Іванович"
            class="bg-background"
          />
        </div>
      </div>
    </div>

    <Separator />

    <div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
      <div class="space-y-1">
        <div class="flex items-center gap-2 font-semibold text-primary">
          <Contact class="h-5 w-5" />
          <span>Контакти</span>
        </div>
        <p class="text-sm text-muted-foreground">
          Зворотний зв'язок та додаткова інформація.
        </p>
      </div>

      <div class="lg:col-span-2 space-y-6">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="email" class="text-sm font-medium"
              >Електронна пошта</Label
            >
            <Input
              id="email"
              name="email"
              type="email"
              placeholder="example@mail.com"
              class="bg-background"
            />
          </div>
          <div class="space-y-2">
            <Label for="phone_number" class="text-sm font-medium"
              >Номер телефону</Label
            >
            <Input
              id="phone_number"
              name="phone_number"
              placeholder="+380..."
              class="bg-background"
            />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="notes" class="text-sm font-medium"
            >Додаткові примітки</Label
          >
          <Textarea
            id="notes"
            name="notes"
            placeholder="Напишіть коментар..."
            class="min-h-[120px] bg-background"
          />
        </div>
      </div>
    </div>

    <div class="flex items-center justify-end gap-4 pt-6 border-t">
      <Button
        variant="ghost"
        href="/admin/users"
        type="button"
        class="hover:bg-accent"
      >
        Скасувати
      </Button>
      <Button type="submit" class="min-w-[150px] shadow-sm">
        Створити акаунт
      </Button>
    </div>
  </form>
</div>

<style>
  select:focus {
    border-color: hsl(var(--ring));
  }
</style>
