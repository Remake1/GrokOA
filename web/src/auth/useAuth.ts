import { ref, computed } from "vue";
import { getToken, setToken, removeToken } from "./cookies";
import { login as apiLogin } from "./api";

const token = ref<string | null>(getToken());

export function useAuth() {
    const isAuthenticated = computed(() => token.value !== null);

    async function login(key: string) {
        const res = await apiLogin(key);
        setToken(res.token);
        token.value = res.token;
    }

    function logout() {
        removeToken();
        token.value = null;
    }

    return { isAuthenticated, login, logout };
}
