import { http, HttpResponse } from "msw";

export const handlers = [
  http.post("/api/auth", async ({ request }) => {
    const body = (await request.json()) as { key?: string };

    if (!body.key) {
      return HttpResponse.json({ error: "key is required" }, { status: 400 });
    }

    if (body.key !== "test1234") {
      return HttpResponse.json({ error: "wrong key" }, { status: 401 });
    }

    return HttpResponse.json({
      token:
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzIzNTA4NjYsImlhdCI6MTc3MjMzNjQ2Nn0.eW2piuiBYRlhzE0Wqfw6anjVPGc5le0Y7FuKmjPL14k",
    });
  }),
];
