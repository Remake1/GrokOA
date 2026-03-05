import { test, expect } from "@playwright/test";

const MOCK_TOKEN =
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDAsImlhdCI6MTg5MzQ1MjQwMH0.dummy-signature";

function mockAuthRoute(page: import("@playwright/test").Page) {
    return page.route("**/api/auth", async (route) => {
        if (route.request().method() !== "POST") {
            return route.fallback();
        }

        const body = route.request().postDataJSON() as { key?: string };

        if (!body.key) {
            return route.fulfill({
                status: 400,
                contentType: "application/json",
                body: JSON.stringify({ error: "key is required" }),
            });
        }

        if (body.key !== "test1234") {
            return route.fulfill({
                status: 401,
                contentType: "application/json",
                body: JSON.stringify({ error: "wrong key" }),
            });
        }

        return route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({ token: MOCK_TOKEN }),
        });
    });
}

test.describe("Auth", () => {
    test("auth guard redirects unauthenticated user to /auth", async ({
        page,
    }) => {
        await page.goto("/");
        await expect(page).toHaveURL(/\/auth/);
    });

    test("guest guard redirects authenticated user away from /auth", async ({
        page,
        context,
    }) => {
        await context.addCookies([
            { name: "token", value: MOCK_TOKEN, domain: "localhost", path: "/" },
        ]);

        await page.goto("/auth");
        await expect(page).toHaveURL("/");
    });

    test("successful login redirects to /", async ({ page }) => {
        await mockAuthRoute(page);

        await page.goto("/auth");
        await page.getByLabel("Access Key").fill("test1234");
        await page.getByRole("button", { name: "Continue" }).click();

        await expect(page).toHaveURL("/");
    });

    test("invalid key shows error message", async ({ page }) => {
        await mockAuthRoute(page);

        await page.goto("/auth");
        await page.getByLabel("Access Key").fill("wrongkey");
        await page.getByRole("button", { name: "Continue" }).click();

        await expect(page.getByText("wrong key")).toBeVisible();
        await expect(page).toHaveURL(/\/auth/);
    });
});
