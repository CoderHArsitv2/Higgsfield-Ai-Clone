"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

export function Nav({ signedIn }: { signedIn: boolean }) {
  const [solid, setSolid] = useState(false);
  const raf = useRef(0);

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
          <span className="grid h-6 w-6 place-items-center rounded-[7px] bg-accent">
            <span className="h-1.5 w-1.5 rounded-full bg-void" />
          </span>
          <span className="font-medium">Aperture</span>
        </Link>

        <div className="hidden items-center gap-8 text-sm text-muted md:flex">
          <a href="#loop" className="transition-colors hover:text-fg">
            How it works
          </a>
          <a href="#models" className="transition-colors hover:text-fg">
            Models
          </a>
          <a href="#keys" className="transition-colors hover:text-fg">
            Your keys
          </a>
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
