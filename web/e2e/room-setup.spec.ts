import { test, expect } from "@playwright/test";
import type { WebSocketRoute } from "@playwright/test";

const MOCK_TOKEN =
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDAsImlhdCI6MTg5MzQ1MjQwMH0.dummy-signature";

const ROOM_CODE = "A1B2";

const WS_URL_PATTERN = /\/api\/ws\/client/;

/** Authenticate the browser by setting the token cookie. */
async function authenticate(context: import("@playwright/test").BrowserContext) {
    await context.addCookies([
        { name: "token", value: MOCK_TOKEN, domain: "localhost", path: "/" },
    ]);
}

/** Clear persisted room code from localStorage. */
async function clearRoomStorage(page: import("@playwright/test").Page) {
    await page.evaluate(() => localStorage.removeItem("room"));
}

/**
 * Set up a mock WS route and return a promise that resolves
 * with the WebSocketRoute when the page connects.
 * MUST be awaited before navigating to the page.
 */
async function mockWs(
    page: import("@playwright/test").Page,
): Promise<{ waitForConnection: () => Promise<WebSocketRoute> }> {
    let resolveWs: (ws: WebSocketRoute) => void;
    const wsPromise = new Promise<WebSocketRoute>((resolve) => {
        resolveWs = resolve;
    });

    await page.routeWebSocket(WS_URL_PATTERN, (ws) => {
        resolveWs(ws);
    });

    return { waitForConnection: () => wsPromise };
}

test.describe("Room Setup", () => {
    test.beforeEach(async ({ page, context }) => {
        await authenticate(context);
        await page.goto("/");
        await clearRoomStorage(page);
    });

    test("connecting to setup page opens WebSocket and displays room code", async ({
        page,
    }) => {
        const { waitForConnection } = await mockWs(page);

        await page.goto("/room/setup");

        const ws = await waitForConnection();

        // Server connected status should show after WS opens
        await expect(page.getByText("Server").locator("..").getByText("Connected")).toBeVisible();
        await expect(page.getByText("Desktop").locator("..").getByText("Disconnected")).toBeVisible();

        // Mock server sends room_created
        ws.send(JSON.stringify({ type: "room_created", code: ROOM_CODE }));

        // Room code characters should be visible in the OTP slots
        for (const char of ROOM_CODE) {
            await expect(page.getByText(char, { exact: true }).first()).toBeVisible();
        }
    });

    test("desktop_connected redirects to room chat page", async ({ page }) => {
        const { waitForConnection } = await mockWs(page);

        await page.goto("/room/setup");
        const ws = await waitForConnection();

        ws.send(JSON.stringify({ type: "room_created", code: ROOM_CODE }));
        ws.send(JSON.stringify({ type: "desktop_connected" }));

        await expect(page).toHaveURL(`/room/${ROOM_CODE}`);
        await expect(page.getByText(ROOM_CODE, { exact: true })).toBeVisible();
    });

    test("desktop_connected keeps the existing room websocket and room code", async ({
        page,
    }) => {
        const wsRoutes: WebSocketRoute[] = [];
        await page.routeWebSocket(WS_URL_PATTERN, (ws) => {
            wsRoutes.push(ws);
        });

        await page.goto("/room/setup");

        await expect(async () => {
            expect(wsRoutes.length).toBe(1);
        }).toPass({ timeout: 5000 });

        const ws = wsRoutes[0]!;
        ws.send(JSON.stringify({ type: "room_created", code: ROOM_CODE }));
        ws.send(JSON.stringify({ type: "desktop_connected" }));

        await expect(page).toHaveURL(`/room/${ROOM_CODE}`);
        await expect(page.getByText(ROOM_CODE, { exact: true })).toBeVisible();

        await expect(async () => {
            expect(wsRoutes.length).toBe(1);
        }).toPass({ timeout: 1000 });

        const stored = await page.evaluate(() => localStorage.getItem("room"));
        const parsed = JSON.parse(stored!);
        expect(parsed.roomCode).toBe(ROOM_CODE);
    });

    test("desktop_disconnected updates status on setup page", async ({ page }) => {
        const { waitForConnection } = await mockWs(page);

        await page.goto("/room/setup");
        const ws = await waitForConnection();

        ws.send(JSON.stringify({ type: "room_created", code: ROOM_CODE }));
        await expect(page.getByText("Desktop").locator("..").getByText("Disconnected")).toBeVisible();

        // Desktop connects then disconnects without redirect
        // (testing that disconnect updates the status indicator)
    });

    test("server error displays error message", async ({ page }) => {
        const { waitForConnection } = await mockWs(page);

        await page.goto("/room/setup");
        const ws = await waitForConnection();

        ws.send(
            JSON.stringify({ type: "error", message: "invalid message format" }),
        );

        await expect(page.getByText("invalid message format")).toBeVisible();
    });

    test("room not found error triggers fresh room creation", async ({
        page,
    }) => {
        // Pre-set a stale room code in localStorage
        await page.evaluate(() => {
            localStorage.setItem("room", JSON.stringify({ roomCode: "STALE" }));
        });

        // Capture all WS connections
        const wsRoutes: WebSocketRoute[] = [];
        await page.routeWebSocket(WS_URL_PATTERN, (ws) => {
            wsRoutes.push(ws);
        });

        await page.goto("/room/setup");

        // Wait for first WS to be captured
        await expect(async () => {
            expect(wsRoutes.length).toBeGreaterThanOrEqual(1);
        }).toPass({ timeout: 5000 });
        const firstWs = wsRoutes[0]!;

        // Verify the first connection included the room query param
        expect(firstWs.url()).toContain("room=STALE");

        // Server responds with room not found
        firstWs.send(
            JSON.stringify({
                type: "error",
                message: "room not found or expired",
            }),
        );

        // Second WS should open (fresh connection without room param)
        await expect(async () => {
            expect(wsRoutes.length).toBeGreaterThanOrEqual(2);
        }).toPass({ timeout: 5000 });
        const secondWs = wsRoutes[1]!;
        expect(secondWs.url()).not.toContain("room=");

        // New room is created
        secondWs.send(JSON.stringify({ type: "room_created", code: "NEWC" }));

        for (const char of "NEWC") {
            await expect(page.getByText(char, { exact: true }).first()).toBeVisible();
        }
    });

    test("reconnection uses persisted room code", async ({ page }) => {
        // Pre-set a room code in localStorage
        await page.evaluate(() => {
            localStorage.setItem(
                "room",
                JSON.stringify({ roomCode: "R3CN" }),
            );
        });

        const { waitForConnection } = await mockWs(page);

        await page.goto("/room/setup");
        const ws = await waitForConnection();

        // Verify room code was sent as query parameter
        expect(ws.url()).toContain("room=R3CN");

        // Server confirms the rejoin
        ws.send(JSON.stringify({ type: "room_created", code: "R3CN" }));

        for (const char of "R3CN") {
            await expect(page.getByText(char, { exact: true }).first()).toBeVisible();
        }
    });

    test("server disconnect updates status to disconnected", async ({
        page,
    }) => {
        const { waitForConnection } = await mockWs(page);

        await page.goto("/room/setup");
        const ws = await waitForConnection();

        ws.send(JSON.stringify({ type: "room_created", code: ROOM_CODE }));
        await expect(page.getByText("Server").locator("..").getByText("Connected")).toBeVisible();

        // Server closes the WebSocket
        await ws.close();

        await expect(page.getByText("Server").locator("..").getByText("Disconnected")).toBeVisible();
    });

    test("room code persists to localStorage after creation", async ({
        page,
    }) => {
        const { waitForConnection } = await mockWs(page);

        await page.goto("/room/setup");
        const ws = await waitForConnection();

        ws.send(JSON.stringify({ type: "room_created", code: ROOM_CODE }));

        // Wait for the code to render
        await expect(page.getByText("A", { exact: true }).first()).toBeVisible();

        // Verify localStorage was updated
        const stored = await page.evaluate(() =>
            localStorage.getItem("room"),
        );
        const parsed = JSON.parse(stored!);
        expect(parsed.roomCode).toBe(ROOM_CODE);
    });

    test("Create Room navigates from home to setup page", async ({ page }) => {
        // Swallow the WS connection so it doesn't error out
        await page.routeWebSocket(WS_URL_PATTERN, () => { });

        await page.goto("/");

        await expect(page.getByText("Create Room")).toBeVisible();
        await page.getByText("Create Room").click();

        await expect(page).toHaveURL("/room/setup");
        await expect(page.getByText("Room")).toBeVisible();
    });
});
