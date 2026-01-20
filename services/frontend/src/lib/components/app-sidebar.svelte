<script lang="ts" module>
  // Іконки
  import Activity from "@lucide/svelte/icons/activity";
  import Camera from "@lucide/svelte/icons/camera";
  import ClipboardList from "@lucide/svelte/icons/clipboard-list";
  import DoorOpen from "@lucide/svelte/icons/door-open";
  import GitFork from "@lucide/svelte/icons/git-fork";
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
        title: "Моніторинг",
        url: "/dashboard",
        icon: LayoutDashboard,
        permissions: ["read:events", "read:trips"],
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
        permissions: ["manage:configs", "read:cameras", "read:scales", "read:gates"],
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
            title: "Гейти",
            url: "/config/gates",
            icon: DoorOpen,
            permissions: ["read:gates"],
          },
          {
            title: "Маршрути",
            url: "/config/flows",
            icon: GitFork,
            permissions: ["read:flows"],
          },
          {
            title: "Налаштування",
            url: "/config/settings",
            icon: SlidersHorizontal,
            permissions: ["read:settings"],
          },
        ],
      },
      {
        title: "Адміністрування",
        url: "#",
        icon: ShieldCheck,
        permissions: ["read:users", "read:roles", "read:keys"],
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
        permissions: ["view:audit"],
      },
    ],
  };

  function filterNavItems(items: any[], permissions: string[]): any[] {
    return items
      .filter((item) => {
        if (!item.permissions) return true;
        return item.permissions.some((p: string) => permissions.includes(p));
      })
      .map((item) => {
        if (item.items) {
          return { ...item, items: filterNavItems(item.items, permissions) };
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
    user?.permissions ? filterNavItems(data.navMain, user.permissions) : []
  );
</script>

<Sidebar.Root {collapsible} {...restProps}>
  <Sidebar.Header>
    <div class="flex items-center gap-2 mx-auto py-2">
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
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