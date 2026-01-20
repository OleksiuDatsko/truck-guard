<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { enhance } from "$app/forms";
  import type { ActionData } from "./$types";
  import { Eye, EyeOff } from "@lucide/svelte";

  let { form } = $props<ActionData>();

  let loading = $state(false);

  let showPassword = $state(false);
</script>

<div class="flex min-h-[100dvh] items-center justify-center bg-background p-4">
  <Card.Root class="w-full max-w-sm">
    <Card.Header>
      <Card.Title class="text-2xl font-bold">Вхід у TruckGuard</Card.Title>
      <Card.Description>
        Введіть ваші дані для доступу до системи.
      </Card.Description>
    </Card.Header>
    <Card.Content>
      <form
        method="POST"
        use:enhance={() => {
          loading = true;
          return async ({ update }) => {
            await update();
            loading = false;
          };
        }}
      >
        <div class="grid w-full items-center gap-4">
          <div class="flex flex-col space-y-1.5">
            <Label for="username">Логін</Label>
            <Input
              id="username"
              name="username"
              required
              disabled={loading}
              class="h-10"
            />
          </div>
          <div class="flex flex-col space-y-1.5">
            <Label for="password">Пароль</Label>
            <div class="relative">
              <Input
                id="password"
                name="password"
                type={showPassword ? "text" : "password"}
                required
                disabled={loading}
                class="h-10 pr-10"
              />
              <Button
                variant="ghost"
                size="icon"
                class="absolute right-0 top-0 h-10 w-10 text-muted-foreground hover:bg-transparent"
                type="button"
                onclick={() => (showPassword = !showPassword)}
              >
                {#if showPassword}
                  <Eye class="h-4 w-4" />
                {:else}
                  <EyeOff class="h-4 w-4" />
                {/if}
                <span class="sr-only"
                  >{showPassword ? "Hide password" : "Show password"}</span
                >
              </Button>
            </div>
          </div>
        </div>

        {#if form?.status === 401}
          <div
            class="mt-4 rounded-md bg-destructive/15 p-3 text-sm text-destructive"
          >
            Невірний логін або пароль
          </div>
        {/if}

        {#if form?.status === 500}
          <div
            class="mt-4 rounded-md bg-destructive/15 p-3 text-sm text-destructive"
          >
            Виникла помилка на сервері
          </div>
        {/if}

        <Button class="mt-6 w-full" size="lg" type="submit" disabled={loading}>
          {#if loading}
            Вхід...
          {:else}
            Увійти
          {/if}
        </Button>
      </form>
    </Card.Content>
  </Card.Root>
</div>
