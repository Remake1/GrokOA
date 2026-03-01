<script setup lang="ts">
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp";

const REGEXP_ONLY_CHARS = "[a-zA-Z0-9]";

defineProps<{
  code: string;
}>();
</script>

<template>
  <InputOTP :model-value="code" :maxlength="4" :pattern="REGEXP_ONLY_CHARS" disabled>
    <InputOTPGroup
      class="gap-2.5 *:data-[slot=input-otp-slot]:h-16 *:data-[slot=input-otp-slot]:w-12 *:data-[slot=input-otp-slot]:rounded-md *:data-[slot=input-otp-slot]:border *:data-[slot=input-otp-slot]:text-xl"
    >
      <InputOTPSlot :index="0" />
      <InputOTPSlot :index="1" />
      <InputOTPSlot :index="2" />
      <InputOTPSlot :index="3" />
    </InputOTPGroup>
  </InputOTP>
</template>
