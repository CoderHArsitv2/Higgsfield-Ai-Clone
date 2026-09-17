"use client";

import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { SplitText } from "gsap/SplitText";

/**
 * Plugins register once. Doing it at module scope rather than inside a
 * component avoids double registration under React strict mode.
 */
if (typeof window !== "undefined") {
  gsap.registerPlugin(ScrollTrigger, SplitText);

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

export { gsap, ScrollTrigger, SplitText };
