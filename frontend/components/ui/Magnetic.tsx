"use client";

import { useRef, type ReactNode } from "react";
import { gsap, prefersReducedMotion } from "@/lib/gsap";

/**
 * Pulls a button gently toward the cursor. Small effect, but it makes the
 * primary CTA feel like the only interactive thing on a very still page.
 */
export function Magnetic({
  children,
  strength = 0.35,
  className,
}: {
  children: ReactNode;
  strength?: number;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);

  const move = (e: React.MouseEvent) => {
    const el = ref.current;
    if (!el || prefersReducedMotion()) return;
    const r = el.getBoundingClientRect();
    gsap.to(el, {
      x: (e.clientX - (r.left + r.width / 2)) * strength,
      y: (e.clientY - (r.top + r.height / 2)) * strength,
      duration: 0.6,
      ease: "power3.out",
    });
  };

  const reset = () => {
    if (ref.current) {
      gsap.to(ref.current, {
        x: 0,
        y: 0,
        duration: 0.8,
        ease: "elastic.out(1, 0.4)",
      });
    }
  };

  return (
    <div
      ref={ref}
      onMouseMove={move}
      onMouseLeave={reset}
      className={className}
      style={{ display: "inline-block", willChange: "transform" }}
    >
      {children}
    </div>
  );
}
