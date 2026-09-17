"use client";

import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { ScrollToPlugin } from "gsap/ScrollToPlugin";
import { SplitText } from "gsap/SplitText";

/**
 * Plugins register once. Doing it at module scope rather than inside a
 * component avoids double registration under React strict mode.
 */
if (typeof window !== "undefined") {
  gsap.registerPlugin(ScrollTrigger, ScrollToPlugin, SplitText);

  // The `js` class is what activates `.reveal-target { opacity: 0 }`. It is
  // added here, not in the server-rendered HTML, so that if this bundle fails
  // to load the content is simply never hidden in the first place.
  document.documentElement.classList.add("js");

  // Second line of defence: if something throws after elements are hidden but
  // before their animation runs, reveal them anyway. An invisible headline is a
  // far worse failure than a missing animation.
  window.setTimeout(() => {
    document.querySelectorAll<HTMLElement>(".reveal-target").forEach((el) => {
      if (getComputedStyle(el).opacity === "0") {
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
 * element's own offset is no longer where the section visually begins. Jumping
 * to the raw offset lands the viewer mid-pin, part-way through an animation
 * that looks broken because it was never played. Targeting the spacer lands on
 * the section's true start.
 */
export function scrollToSection(hash: string, smooth = true) {
  const el = document.querySelector<HTMLElement>(hash);
  if (!el) return;

  ScrollTrigger.refresh();

  // If this section is pinned, its own offsetTop is meaningless while pinned --
  // the element is transformed and sits inside a spacer. The trigger itself
  // knows the exact scroll position where the section begins, so ask it.
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
