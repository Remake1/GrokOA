<script setup lang="ts">
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
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
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
      <form>
        <FieldGroup>
          <Field>
            <FieldLabel for="auth-access-key">
              Access Key
            </FieldLabel>
            <Input
              id="auth-access-key"
              type="password"
              placeholder="Enter your access key"
              required
            />
          </Field>
          <Button type="submit" class="w-full">
            Continue
          </Button>
        </FieldGroup>
      </form>
    </CardContent>
  </Card>
</template>
