import { NextRequest, NextResponse } from "next/server";
import { auth0 } from "@/lib/auth0";

const API = process.env.BACKEND_URL ?? "http://localhost:8080";

/**
 * Server-side proxy to the Go API.
 *
 * Client components never see the Auth0 access token: they call this route,
 * and the token is attached here. That keeps a bearer credential out of the
 * browser entirely, which is the whole reason for the extra hop.
 */
async function forward(req: NextRequest, path: string[]) {
  let token: string;
  try {
    const result = await auth0.getAccessToken();
    token = result.token;
  } catch {
    return NextResponse.json(
      { error: { code: "unauthorized", message: "Sign in to continue" } },
      { status: 401 },
    );
  }

  const url = new URL(`${API}/api/v1/${path.join("/")}`);
  req.nextUrl.searchParams.forEach((v, k) => url.searchParams.set(k, v));

  const init: RequestInit = {
    method: req.method,
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    cache: "no-store",
  };
  if (req.method !== "GET" && req.method !== "DELETE") {
    init.body = await req.text();
  }

  try {
    const upstream = await fetch(url, init);
    if (upstream.status === 204) return new NextResponse(null, { status: 204 });
    const body = await upstream.text();
    return new NextResponse(body, {
      status: upstream.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    // The API being down should read as a service problem, not a broken page.
    return NextResponse.json(
      {
        error: {
          code: "backend_unreachable",
          message: "The generation service is not responding.",
        },
      },
      { status: 503 },
    );
  }
}

type Ctx = { params: Promise<{ path: string[] }> };

export async function GET(req: NextRequest, ctx: Ctx) {
  return forward(req, (await ctx.params).path);
}
export async function POST(req: NextRequest, ctx: Ctx) {
  return forward(req, (await ctx.params).path);
}
export async function PUT(req: NextRequest, ctx: Ctx) {
  return forward(req, (await ctx.params).path);
}
export async function DELETE(req: NextRequest, ctx: Ctx) {
  return forward(req, (await ctx.params).path);
}
