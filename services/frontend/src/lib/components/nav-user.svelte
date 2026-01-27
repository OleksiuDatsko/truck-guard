<script lang="ts">
  import * as Avatar from "$lib/components/ui/avatar/index.js";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import { useSidebar } from "$lib/components/ui/sidebar/index.js";
  import BadgeCheckIcon from "@lucide/svelte/icons/badge-check";
  import ChevronsUpDownIcon from "@lucide/svelte/icons/chevrons-up-down";
  import SunIcon from "@lucide/svelte/icons/sun";
  import LogOutIcon from "@lucide/svelte/icons/log-out";
  import type { CoreUser as User } from "$lib/server/core-client";
  import { toggleMode } from "mode-watcher";
  import { MoonIcon } from "@lucide/svelte";
  import { goto } from "$app/navigation";

  let { user }: { user: User } = $props();
  const sidebar = useSidebar();
  let logoutForm: HTMLFormElement;
</script>

<form
  action="/logout"
  method="POST"
  bind:this={logoutForm}
  class="hidden"
></form>

<Sidebar.Menu>
  <Sidebar.MenuItem>
    <DropdownMenu.Root>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <Sidebar.MenuButton
            size="lg"
            class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            {...props}
          >
            <Avatar.Root class="size-8 rounded-lg">
              <Avatar.Fallback class="rounded-lg">
                {#if user.first_name && user.last_name}
                  {user.first_name.substring(0, 1).toUpperCase()}{user.last_name
                    .substring(0, 1)
                    .toUpperCase()}
                {:else}
                  AD
                {/if}
              </Avatar.Fallback>
            </Avatar.Root>
            <div class="grid flex-1 text-start text-sm leading-tight">
              <span class="font-medium"
                >{#if user.first_name && user.last_name}{user.first_name}
                  {user.last_name}{:else}Admin{/if}</span
              >
              <span class="text-xs"
                >{#if user.email}{user.email}{:else}admin@admin.com{/if}</span
              >
            </div>
            <ChevronsUpDownIcon class="ms-auto size-4" />
          </Sidebar.MenuButton>
        {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
        side={sidebar.isMobile ? "bottom" : "right"}
        align="end"
        sideOffset={4}
      >
        <DropdownMenu.Label class="p-0 font-normal">
          <div class="flex items-center gap-2 px-1 py-1.5 text-start text-sm">
            <Avatar.Root class="size-8 rounded-lg">
              <Avatar.Fallback class="rounded-lg">
                {#if user.first_name && user.last_name}
                  {user.first_name.substring(0, 1).toUpperCase()}{user.last_name
                    .substring(0, 1)
                    .toUpperCase()}
                {:else}
                  AD
                {/if}
              </Avatar.Fallback>
            </Avatar.Root>
            <div class="grid flex-1 text-start text-sm leading-tight">
              <span class="truncate font-medium"
                >{#if user.first_name && user.last_name}{user.first_name}
                  {user.last_name}{:else}Admin{/if}</span
              >
              <span class="truncate text-xs"
                >{#if user.email}{user.email}{:else}admin{/if}</span
              >
            </div>
          </div>
        </DropdownMenu.Label>
        <DropdownMenu.Separator />
        <DropdownMenu.Group>
          <DropdownMenu.Item onclick={() => goto('/profile')}>
            <BadgeCheckIcon />
            Профіль
          </DropdownMenu.Item>
          <DropdownMenu.Item onclick={toggleMode}>
            <SunIcon
              class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 !transition-all dark:scale-0 dark:-rotate-90"
            />
            <MoonIcon
              class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 !transition-all dark:scale-100 dark:rotate-0"
            />
            Тема
          </DropdownMenu.Item>
        </DropdownMenu.Group>
        <DropdownMenu.Separator />
        <DropdownMenu.Item onclick={() => logoutForm.requestSubmit()}>
          <LogOutIcon />
          Вийти
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </Sidebar.MenuItem>
</Sidebar.Menu>
