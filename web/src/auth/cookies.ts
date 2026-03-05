const COOKIE_NAME = "token";
/** Minimum remaining lifetime (in seconds) before a token is considered expired. */
const MIN_TOKEN_LIFETIME = 3600; // 1 hour

interface JwtPayload {
    exp?: number;
}

function parsePayload(token: string): JwtPayload | null {
    const parts = token.split(".");
    if (parts.length < 2 || !parts[1]) {
        return null;
    }

    try {
        const b64url = parts[1];
        const b64 = b64url.replace(/-/g, "+").replace(/_/g, "/");
        const padded = b64 + "=".repeat((4 - (b64.length % 4)) % 4);

        return JSON.parse(atob(padded)) as JwtPayload;
    } catch {
        return null;
    }
}

function tokenMaxAge(token: string): number | null {
    const payload = parsePayload(token);
    if (typeof payload?.exp !== "number") {
        return null;
    }

    return payload.exp - Math.floor(Date.now() / 1000);
}

export function getToken(): string | null {
    const match = document.cookie.match(
        new RegExp(`(?:^|;\\s*)${COOKIE_NAME}=([^;]*)`),
    );
    const token = match ? decodeURIComponent(match[1] ?? "") : null;
    if (!token) {
        return null;
    }

    const maxAge = tokenMaxAge(token);
    if (maxAge !== null && maxAge < MIN_TOKEN_LIFETIME) {
        removeToken();
        return null;
    }

    return token;
}

export function setToken(token: string): void {
    const maxAge = tokenMaxAge(token);
    if (maxAge !== null && maxAge <= 0) {
        removeToken();
        return;
    }

    const base = `${COOKIE_NAME}=${encodeURIComponent(token)}; path=/; SameSite=Strict`;
    document.cookie =
        maxAge !== null ? `${base}; max-age=${maxAge}` : base;
}

export function removeToken(): void {
    document.cookie = `${COOKIE_NAME}=; path=/; max-age=0`;
}
