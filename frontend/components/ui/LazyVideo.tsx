"use client";

import { useEffect, useRef } from "react";

/**
 * A looping clip that only decodes while it is on screen.
 *
 * The page has around a dozen of these. Letting them all autoplay means a
 * dozen simultaneous video decodes competing with scroll for main-thread and
 * GPU time, which is what makes a scroll-driven page feel heavy. An
 * IntersectionObserver plays the one or two in view and pauses the rest, and
 * `preload="none"` means off-screen clips cost no bandwidth at all until they
 * are needed.
 */
export function LazyVideo({
  src,
  poster,
  className,
  eager = false,
  active = true,
}: {
  src: string;
  poster?: string;
  className?: string;
  /** Above the fold: fetch metadata up front so playback starts promptly. */
  eager?: boolean;
  /** False keeps the clip paused on its poster even while on screen. */
  active?: boolean;
}) {
  const ref = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    if (!active) {
      if (!el.paused) el.pause();
      return;
    }

    const io = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          // Autoplay can still be refused; a paused poster is an acceptable
          // outcome and must not throw.
          void el.play().catch(() => {});
        } else if (!el.paused) {
          el.pause();
        }
      },
      { threshold: 0.2 },
    );

    io.observe(el);
    return () => io.disconnect();
  }, [active]);

  return (
    <video
      ref={ref}
      src={src}
      poster={poster}
      className={className}
      muted
      loop
      playsInline
      preload={eager ? "metadata" : "none"}
    />
  );
}
