<script setup lang="ts">
import { ref, watch } from "vue";
import { useRouter } from "vue-router";
import { LogOut, DoorOpen, Save } from "lucide-vue-next";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import LanguageSelector from "@/settings/components/LanguageSelector.vue";
import PromptEditor from "@/settings/components/PromptEditor.vue";
import {
  useSettingsStore,
  LANGUAGE_OPTIONS,
  DEFAULT_CODING_PROMPT,
  DEFAULT_MCQ_PROMPT,
} from "@/settings/store";
import { useRoomStore } from "@/room/store";
import { useAuth } from "@/auth/useAuth";

const open = defineModel<boolean>("open", { required: true });

const props = defineProps<{
  disconnect: () => void;
}>();

const router = useRouter();
const settings = useSettingsStore();
const roomStore = useRoomStore();
const { logout } = useAuth();

// --- Prompt tab draft state (mirrors SettingsView pattern) ---
const draftLanguage = ref(settings.codingLanguage);
const draftCodingPrompt = ref(settings.codingPrompt);
const draftMcqPrompt = ref(settings.mcqPrompt);

const hasChanges = ref(false);

watch(
  [draftLanguage, draftCodingPrompt, draftMcqPrompt],
  () => {
    hasChanges.value =
      draftLanguage.value !== settings.codingLanguage ||
      draftCodingPrompt.value !== settings.codingPrompt ||
      draftMcqPrompt.value !== settings.mcqPrompt;
  },
  { immediate: true },
);

// Reset drafts when dialog opens
watch(open, (isOpen) => {
  if (isOpen) {
    draftLanguage.value = settings.codingLanguage;
    draftCodingPrompt.value = settings.codingPrompt;
    draftMcqPrompt.value = settings.mcqPrompt;
  }
});

function savePrompts() {
  settings.codingLanguage = draftLanguage.value;
  settings.codingPrompt = draftCodingPrompt.value;
  settings.mcqPrompt = draftMcqPrompt.value;
}

function exitSession() {
  props.disconnect();
  roomStore.reset();
  localStorage.removeItem("room");
  open.value = false;
  router.push("/");
}

function logoutUser() {
  props.disconnect();
  roomStore.reset();
  localStorage.removeItem("room");
  logout();
  open.value = false;
  router.push("/auth");
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent
      class="flex h-[90dvh] w-[95vw] max-w-none flex-col gap-0 p-0"
    >
      <Tabs default-value="session" class="flex min-h-0 flex-1 flex-col">
        <DialogHeader class="flex shrink-0 items-center justify-center border-b border-border px-4 py-2">
          <DialogTitle class="sr-only">Settings</DialogTitle>
          <DialogDescription class="sr-only">
            Manage session and prompt settings
          </DialogDescription>
          <TabsList class="h-8">
            <TabsTrigger value="session" class="text-xs px-3 py-1">Session</TabsTrigger>
            <TabsTrigger value="prompt" class="text-xs px-3 py-1">Prompt</TabsTrigger>
          </TabsList>
        </DialogHeader>

        <!-- ─── Session Tab ─── -->
        <TabsContent
          value="session"
          class="flex-1 overflow-y-auto px-5 py-4"
        >
          <div class="flex flex-col gap-4">
            <div>
              <h3 class="mb-1 text-sm font-medium">Exit Session</h3>
              <p class="mb-3 text-xs text-muted-foreground">
                Disconnect from the current room and return to the home screen.
                The room code will be cleared.
              </p>
              <Button
                id="settings-exit-session"
                variant="outline"
                class="w-full gap-2"
                @click="exitSession"
              >
                <DoorOpen class="size-4" />
                Exit Session
              </Button>
            </div>

            <Separator />

            <div>
              <h3 class="mb-1 text-sm font-medium">Logout</h3>
              <p class="mb-3 text-xs text-muted-foreground">
                End the current session and log out of your account.
              </p>
              <Button
                id="settings-logout"
                variant="destructive"
                class="w-full gap-2"
                @click="logoutUser"
              >
                <LogOut class="size-4" />
                Logout
              </Button>
            </div>
          </div>
        </TabsContent>

        <!-- ─── Prompt Tab ─── -->
        <TabsContent
          value="prompt"
          class="flex-1 overflow-y-auto px-5 py-4"
        >
          <div class="flex flex-col gap-5">
            <LanguageSelector
              v-model="draftLanguage"
              :options="LANGUAGE_OPTIONS"
            />

            <Separator />

            <PromptEditor
              v-model="draftCodingPrompt"
              label="Coding Prompt"
              :default-value="DEFAULT_CODING_PROMPT"
            />

            <Separator />

            <PromptEditor
              v-model="draftMcqPrompt"
              label="MCQ Prompt"
              :default-value="DEFAULT_MCQ_PROMPT"
            />

            <Separator />

            <Button
              id="settings-modal-save"
              class="w-full gap-2"
              :disabled="!hasChanges"
              @click="savePrompts"
            >
              <Save class="size-4" />
              Save
            </Button>
          </div>
        </TabsContent>
      </Tabs>
    </DialogContent>
  </Dialog>
</template>
