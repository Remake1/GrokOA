import type { NavigationGuardWithThis } from "vue-router";
import { useAuth } from "@/auth/useAuth";

export const authGuard: NavigationGuardWithThis<undefined> = () => {
    const { isAuthenticated } = useAuth();
    if (!isAuthenticated.value) {
        return { name: "auth" };
    }
};

export const guestGuard: NavigationGuardWithThis<undefined> = () => {
    const { isAuthenticated } = useAuth();
    if (isAuthenticated.value) {
        return { name: "home" };
    }
};
