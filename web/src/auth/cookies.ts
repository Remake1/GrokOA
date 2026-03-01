const COOKIE_NAME = "token";

export function getToken(): string | null {
    const match = document.cookie.match(
        new RegExp(`(?:^|;\\s*)${COOKIE_NAME}=([^;]*)`),
    );
    return match ? decodeURIComponent(match[1] ?? "") : null;
}

export function setToken(token: string): void {
    document.cookie = `${COOKIE_NAME}=${encodeURIComponent(token)}; path=/; SameSite=Strict`;
}

export function removeToken(): void {
    document.cookie = `${COOKIE_NAME}=; path=/; max-age=0`;
}
