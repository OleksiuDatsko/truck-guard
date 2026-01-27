<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import { enhance } from "$app/forms";
  import { toast } from "svelte-sonner";
  import { User, Mail, Phone, FileText } from "@lucide/svelte";

  let { data } = $props();

  let { profile, user } = $derived(data);

  let rawRole = $derived(profile?.role || user.role);
  let roleName = $derived(
    typeof rawRole === "object" ? rawRole?.name : rawRole,
  );
</script>

<div class="container max-w-2xl py-10 mx-auto">
  <div class="mb-8 space-y-2">
    <h1 class="text-3xl font-bold tracking-tight">Мій Профіль</h1>
    <p class="text-muted-foreground">
      Керуйте своєю особистою інформацією та налаштуваннями.
    </p>
  </div>

  <div class="space-y-6">
    <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6">
      <h2 class="text-lg font-semibold mb-4">Обліковий запис</h2>
      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-1">
          <Label class="text-xs text-muted-foreground">Логін</Label>
          <div class="font-medium">{user.username || profile?.username}</div>
        </div>
        <div class="space-y-1">
          <Label class="text-xs text-muted-foreground">Роль</Label>
          <div class="font-medium capitalize">
            {roleName}
          </div>
        </div>
      </div>
    </div>

    <form
      method="POST"
      class="space-y-8 rounded-lg border bg-card text-card-foreground shadow-sm p-6"
      use:enhance={() => {
        toast.loading("Збереження змін...");
        return async ({ result, update }) => {
          if (result.type === "success") {
            toast.success("Профіль оновлено успішно");
            await update({ reset: false });
          } else {
            toast.error("Не вдалося оновити профіль");
          }
        };
      }}
    >
      <div class="space-y-4">
        <h2 class="text-lg font-semibold flex items-center gap-2">
          <User class="h-5 w-5" /> Особисті дані
        </h2>

        <div class="grid gap-4 md:grid-cols-3">
          <div class="space-y-2">
            <Label for="last_name">Прізвище</Label>
            <Input
              id="last_name"
              name="last_name"
              value={user.last_name || ""}
            />
          </div>
          <div class="space-y-2">
            <Label for="first_name">Ім'я</Label>
            <Input
              id="first_name"
              name="first_name"
              value={user.first_name || ""}
            />
          </div>
          <div class="space-y-2">
            <Label for="third_name">По батькові</Label>
            <Input
              id="third_name"
              name="third_name"
              value={user.third_name || ""}
            />
          </div>
        </div>

        <div class="space-y-2">
          <Label for="email" class="flex items-center gap-2">
            <Mail class="h-4 w-4" /> Email
          </Label>
          <Input
            id="email"
            name="email"
            type="email"
            value={user.email || ""}
          />
        </div>

        <div class="space-y-2">
          <Label for="phone_number" class="flex items-center gap-2">
            <Phone class="h-4 w-4" /> Телефон
          </Label>
          <Input
            id="phone_number"
            name="phone_number"
            value={user.phone_number || ""}
          />
        </div>

        <div class="space-y-2">
          <Label for="notes" class="flex items-center gap-2">
            <FileText class="h-4 w-4" /> Примітки
          </Label>
          <Textarea
            id="notes"
            name="notes"
            value={user.notes || ""}
            placeholder="Додаткова інформація про себе..."
            class="min-h-[100px]"
          />
        </div>
      </div>

      <div class="flex justify-end pt-4">
        <Button type="submit">Зберегти зміни</Button>
      </div>
    </form>
  </div>
</div>
