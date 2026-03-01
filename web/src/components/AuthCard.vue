<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useForm } from "@tanstack/vue-form";
import { authSchema } from "@/auth/schema";
import { useAuth } from "@/auth/useAuth";
import { AuthError } from "@/auth/api";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";

const router = useRouter();
const { login } = useAuth();
const serverError = ref<string | null>(null);

const form = useForm({
  defaultValues: {
    key: "",
  },
  validators: {
    onSubmit: authSchema,
  },
  onSubmit: async ({ value }) => {
    serverError.value = null;
    try {
      await login(value.key);
      router.push("/");
    } catch (err) {
      if (err instanceof AuthError) {
        serverError.value = err.message;
      } else {
        serverError.value = "Something went wrong. Please try again.";
      }
    }
  },
});

function isInvalid(field: any) {
  return field.state.meta.isTouched && !field.state.meta.isValid;
}
</script>

<template>
  <Card class="w-full max-w-sm">
    <CardHeader>
      <CardTitle>Authorization</CardTitle>
      <CardDescription>
        Enter your access key to continue
      </CardDescription>
    </CardHeader>
    <CardContent>
      <form id="form-auth" @submit.prevent="form.handleSubmit">
        <FieldGroup>
          <form.Field name="key">
            <template #default="{ field }">
              <Field :data-invalid="isInvalid(field)">
                <FieldLabel :for="field.name">
                  Access Key
                </FieldLabel>
                <Input
                  :id="field.name"
                  :name="field.name"
                  :model-value="field.state.value"
                  :aria-invalid="isInvalid(field)"
                  type="password"
                  placeholder="Enter your access key"
                  autocomplete="off"
                  @blur="field.handleBlur"
                  @input="field.handleChange(($event.target as HTMLInputElement).value)"
                />
                <FieldError
                  v-if="isInvalid(field)"
                  :errors="field.state.meta.errors"
                />
              </Field>
            </template>
          </form.Field>

          <p v-if="serverError" class="text-destructive text-sm">
            {{ serverError }}
          </p>

          <Button type="submit" class="w-full" form="form-auth">
            Continue
          </Button>
        </FieldGroup>
      </form>
    </CardContent>
  </Card>
</template>
