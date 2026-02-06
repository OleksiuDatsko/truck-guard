<script lang="ts" module>
  import { can } from "$lib/auth";
  // Іконки
  import Activity from "@lucide/svelte/icons/activity";
  import Camera from "@lucide/svelte/icons/camera";
  import ClipboardList from "@lucide/svelte/icons/clipboard-list";
  import LayoutDashboard from "@lucide/svelte/icons/layout-dashboard";
  import Scale from "@lucide/svelte/icons/scale";
  import Settings from "@lucide/svelte/icons/settings";
  import Users from "@lucide/svelte/icons/users";
  import Key from "@lucide/svelte/icons/key";
  import ShieldCheck from "@lucide/svelte/icons/shield-check";
  import SlidersHorizontal from "@lucide/svelte/icons/sliders-horizontal";

  const data = {
    navMain: [
      {
        title: "Події",
        url: "/events",
        icon: LayoutDashboard,
        permissions: ["read:events"],
      },
      {
        title: "Журнал перепусток",
        url: "/permits",
        icon: ClipboardList,
        permissions: ["read:permits"],
      },
      {
        title: "Конфігурація",
        url: "#",
        icon: Settings,
        permissions: ["read:settings"],
        items: [
          {
            title: "Камери",
            url: "/config/cameras",
            icon: Camera,
            permissions: ["read:cameras"],
          },
          {
            title: "Ваги",
            url: "/config/scales",
            icon: Scale,
            permissions: ["read:scales"],
          },
          {
            title: "Налаштування",
            url: "/config/settings",
            icon: SlidersHorizontal,
            permissions: ["update:settings"],
          },
        ],
      },
      {
        title: "Адміністрування",
        url: "#",
        icon: ShieldCheck,
        permissions: ["read:users"],
        items: [
          {
            title: "Користувачі",
            url: "/admin/users",
            icon: Users,
            permissions: ["read:users"],
          },
          {
            title: "Ролі та права",
            url: "/admin/roles",
            icon: ShieldCheck,
            permissions: ["read:roles"],
          },
          {
            title: "API Ключі",
            url: "/admin/keys",
            icon: Key,
            permissions: ["read:keys"],
          },
        ],
      },
      {
        title: "Системний аудит",
        url: "/system/audit",
        icon: Activity,
        permissions: ["read:audit"],
      },
    ],
  };

  function filterNavItems(items: any[], user: any): any[] {
    return items
      .filter((item) => {
        if (!item.permissions) return true;
        return item.permissions.some((p: string) => can(user, p));
      })
      .map((item) => {
        if (item.items) {
          return { ...item, items: filterNavItems(item.items, user) };
        }
        return item;
      })
      .filter((item) => {
        if (item.items && item.items.length === 0 && item.url === "#") {
          return false;
        }
        return true;
      });
  }
</script>

<script lang="ts">
  import NavMain from "./nav-main.svelte";
  import NavUser from "./nav-user.svelte";
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import { useSidebar } from "$lib/components/ui/sidebar/index.js";
  import type { ComponentProps } from "svelte";

  let {
    ref = $bindable(null),
    collapsible = "icon",
    user,
    ...restProps
  }: ComponentProps<typeof Sidebar.Root> & {
    user: any;
  } = $props();

  const sidebar = useSidebar();

  let filteredNavMain = $derived(
    user?.permissions ? filterNavItems(data.navMain, user) : [],
  );
</script>

<Sidebar.Root {collapsible} {...restProps}>
  <Sidebar.Header>
    <div class="flex items-center gap-2 mx-auto py-2">
      <div
        class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
      >
        <span class="font-bold">TG</span>
      </div>
      {#if sidebar.state !== "collapsed"}
        <div class="flex flex-col gap-0.5 leading-none">
          <span class="font-semibold text-lg tracking-tight">TruckGuard</span>
          <span class="text-xs text-muted-foreground">Logistics Control</span>
        </div>
      {/if}
    </div>
  </Sidebar.Header>

  <Sidebar.Content>
    <NavMain items={filteredNavMain} />
  </Sidebar.Content>

  <Sidebar.Footer>
    <NavUser {user} />
  </Sidebar.Footer>
  <Sidebar.Rail />
</Sidebar.Root>
