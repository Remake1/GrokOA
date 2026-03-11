<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import wakeLockVideo from "@/assets/wakelock.mp4";

type WakeLockSentinelLike = {
  released: boolean;
  release: () => Promise<void>;
  addEventListener: (
    type: "release",
    listener: () => void,
    options?: AddEventListenerOptions,
  ) => void;
  removeEventListener: (
    type: "release",
    listener: () => void,
    options?: EventListenerOptions,
  ) => void;
};

type NavigatorWithWakeLock = Navigator & {
  wakeLock?: {
    request: (type: "screen") => Promise<WakeLockSentinelLike>;
  };
};

const wakeLock = ref<WakeLockSentinelLike | null>(null);
const useVideoFallback = ref(false);
const videoEl = ref<HTMLVideoElement | null>(null);
const shouldUseWakeLockApi = computed(() => !useVideoFallback.value);

function handleWakeLockRelease() {
  wakeLock.value = null;

  if (document.visibilityState === "visible") {
    void requestScreenWakeLock();
  }
}

async function releaseWakeLock() {
  const sentinel = wakeLock.value;
  if (!sentinel) return;

  sentinel.removeEventListener("release", handleWakeLockRelease);
  wakeLock.value = null;

  if (!sentinel.released) {
    await sentinel.release();
  }
}

async function startVideoFallback() {
  useVideoFallback.value = true;
  await nextTick();

  try {
    await videoEl.value?.play();
  } catch {
    // Muted inline autoplay should typically succeed; ignore if the browser blocks it.
  }
}

async function requestScreenWakeLock() {
  if (document.visibilityState !== "visible") return;

  const navigatorWithWakeLock = navigator as NavigatorWithWakeLock;
  if (!navigatorWithWakeLock.wakeLock) {
    await startVideoFallback();
    return;
  }

  try {
    await releaseWakeLock();
    const sentinel = await navigatorWithWakeLock.wakeLock.request("screen");
    sentinel.addEventListener("release", handleWakeLockRelease);
    wakeLock.value = sentinel;
    useVideoFallback.value = false;
  } catch {
    await startVideoFallback();
  }
}

function handleVisibilityChange() {
  if (document.visibilityState === "visible") {
    if (shouldUseWakeLockApi.value) {
      void requestScreenWakeLock();
    } else {
      void videoEl.value?.play();
    }

    return;
  }

  if (shouldUseWakeLockApi.value) {
    void releaseWakeLock();
    return;
  }

  videoEl.value?.pause();
}

watch(useVideoFallback, async (enabled) => {
  if (enabled) {
    await nextTick();
    try {
      await videoEl.value?.play();
    } catch {
      // Ignore blocked autoplay for the fallback media element.
    }
    return;
  }

  videoEl.value?.pause();
});

onMounted(() => {
  document.addEventListener("visibilitychange", handleVisibilityChange);
  void requestScreenWakeLock();
});

onUnmounted(async () => {
  document.removeEventListener("visibilitychange", handleVisibilityChange);
  videoEl.value?.pause();
  await releaseWakeLock();
});
</script>

<template>
  <video
    v-if="useVideoFallback"
    ref="videoEl"
    autoplay
    muted
    loop
    playsinline
    style="position: fixed; width: 2px; height: 2px; opacity: 0"
  >
    <source :src="wakeLockVideo" type="video/mp4" />
  </video>
</template>
