"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { scrollToSection } from "@/lib/gsap";

const NAV_LINKS: [string, string][] = [
  ["#loop", "How it works"],
  ["#models", "Models"],
  ["#keys", "Your keys"],
];

export function Nav({ signedIn }: { signedIn: boolean }) {
  const [solid, setSolid] = useState(false);
  const raf = useRef(0);

  // A page loaded with a hash is positioned by the browser before ScrollTrigger
  // has created its pin spacers, so it lands part-way into a pinned section.
  // Re-resolve the target once layout has settled.
  useEffect(() => {
    let t = 0;
    const settle = (smooth: boolean) => {
      const hash = window.location.hash;
      if (!hash) return;
      window.clearTimeout(t);
      t = window.setTimeout(() => scrollToSection(hash, smooth), 450);
    };

    settle(false);
    // Back/forward between sections only changes the hash, so the effect above
    // never re-runs. Without this the browser drops the viewer at the raw
    // offset, part-way into a pinned section.
    const onHashChange = () => settle(true);
    window.addEventListener("hashchange", onHashChange);
    return () => {
      window.clearTimeout(t);
      window.removeEventListener("hashchange", onHashChange);
    };
  }, []);

  useEffect(() => {
    const onScroll = () => {
      cancelAnimationFrame(raf.current);
      raf.current = requestAnimationFrame(() => setSolid(window.scrollY > 24));
    };
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => {
      window.removeEventListener("scroll", onScroll);
      cancelAnimationFrame(raf.current);
    };
  }, []);

  return (
    <header
      className={`fixed inset-x-0 top-0 z-50 transition-colors duration-500 ${
        solid
          ? "border-b border-line bg-void/80 backdrop-blur-xl"
          : "border-b border-transparent"
      }`}
    >
      <nav className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <Link
          href="/"
          className="flex items-center gap-2.5 text-sm tracking-tight"
        >
          {/* A 600-byte SVG: next/image would add a request and an optimizer
              pass for no benefit, and it does not optimise SVG anyway.
              Served straight from public/ so the mark has one source of truth. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src="/logo-mark.svg"
            alt=""
            width={24}
            height={24}
            className="h-6 w-6"
          />
          <span className="font-medium">Aperture</span>
        </Link>

        <div className="hidden items-center gap-8 text-sm text-muted md:flex">
          {NAV_LINKS.map(([hash, label]) => (
            <a
              key={hash}
              href={hash}
              onClick={(e) => {
                e.preventDefault();
                scrollToSection(hash);
              }}
              className="transition-colors hover:text-fg"
            >
              {label}
            </a>
          ))}
        </div>

        <div className="flex items-center gap-3">
          {signedIn ? (
            <>
              <Link
                href="/studio"
                className="text-sm text-muted transition-colors hover:text-fg"
              >
                Studio
              </Link>
              <a
                href="/auth/logout"
                className="rounded-full border border-line px-4 py-1.5 text-sm transition-colors hover:border-dim"
              >
                Sign out
              </a>
            </>
          ) : (
            <a
              href="/auth/login"
              className="rounded-full bg-fg px-4 py-1.5 text-sm font-medium text-void transition-transform hover:scale-[1.03]"
            >
              Sign in
            </a>
          )}
        </div>
      </nav>
    </header>
  );
}
