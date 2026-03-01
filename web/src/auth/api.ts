export class AuthError extends Error {
    status: number;

    constructor(
        message: string,
        status: number,
    ) {
        super(message);
        this.name = "AuthError";
        this.status = status;
    }
}

interface AuthResponse {
    token: string;
}

interface ErrorResponse {
    error: string;
}

export async function login(key: string): Promise<AuthResponse> {
    const res = await fetch("/api/auth", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ key }),
    });

    if (!res.ok) {
        const body = (await res.json()) as ErrorResponse;
        throw new AuthError(body.error, res.status);
    }

    return res.json() as Promise<AuthResponse>;
}
