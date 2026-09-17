import type { NextRequest } from "next/server";
import { auth0 } from "./lib/auth0";

/**
 * Mounts /auth/login, /auth/logout, /auth/callback and /auth/access-token, and
 * keeps the session rolling. Routes are not protected here -- pages decide for
 * themselves -- so the landing page stays public and fast.
 */
export async function middleware(request: NextRequest) {
  return auth0.middleware(request);
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt|.*\\.(?:svg|png|jpg|jpeg|gif|webp|mp4|woff2?)$).*)",
  ],
};
