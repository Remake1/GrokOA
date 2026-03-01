import { z } from "zod";

export const authSchema = z.object({
    key: z
        .string()
        .min(1, "Access key is required")
        .max(32, "Access key must be at most 32 characters"),
});

export type AuthForm = z.infer<typeof authSchema>;
