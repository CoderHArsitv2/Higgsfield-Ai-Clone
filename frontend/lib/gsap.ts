"use client";

import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { ScrollToPlugin } from "gsap/ScrollToPlugin";
import { SplitText } from "gsap/SplitText";

declare global {
  interface Window {
    __apertureReady?: boolean;
  }
}

if (typeof window !== "undefined") {
  gsap.registerPlugin(ScrollTrigger, ScrollToPlugin, SplitText);

  // Tells the inline boot-watchdog in the document head that the app came up,
  // so it leaves the `js` class alone.
  window.__apertureReady = true;

  /**
   * Safety net for the case where this bundle loaded but a section's animation
   * threw before running.
   *
   * It must not touch anything GSAP is deliberately holding at zero opacity --
   * the inactive panels of the loop section are supposed to be invisible. GSAP
   * writes inline styles, so an element with an inline opacity was reached by
   * an animation and is left alone; one still at zero purely from the
   * stylesheet never was, and is revealed.
   */
  window.setTimeout(() => {
    document.querySelectorAll<HTMLElement>(".reveal-target").forEach((el) => {
      const untouched = el.style.opacity === "" && el.style.transform === "";
      if (untouched && getComputedStyle(el).opacity === "0") {
        el.style.opacity = "1";
        el.style.transform = "none";
      }
    });
  }, 4000);
}

/** Honour the OS setting. Motion is decoration; content is not. */
export const prefersReducedMotion = () =>
  typeof window !== "undefined" &&
  window.matchMedia("(prefers-reduced-motion: reduce)").matches;

/**
 * Scroll to a section by id.
 *
 * Pinned sections are wrapped by ScrollTrigger in a `.pin-spacer`, so the
 * element's own offset is no longer where the section visually begins. The
 * trigger knows its real start, so ask it.
 */
export function scrollToSection(hash: string, smooth = true) {
  const el = document.querySelector<HTMLElement>(hash);
  if (!el) return;

  ScrollTrigger.refresh();
  const pinned = ScrollTrigger.getAll().find((t) => t.trigger === el);
  const y = pinned
    ? pinned.start
    : (el.closest<HTMLElement>(".pin-spacer") ?? el).getBoundingClientRect()
        .top + window.scrollY;

  if (!smooth || prefersReducedMotion()) {
    window.scrollTo({ top: y, behavior: "auto" });
    return;
  }
  gsap.to(window, {
    duration: 1,
    ease: "power2.inOut",
    scrollTo: { y, autoKill: true },
  });
}

export { gsap, ScrollTrigger, ScrollToPlugin, SplitText };
