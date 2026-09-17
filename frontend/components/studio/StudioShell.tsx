"use client";

import { useState } from "react";
import { AppNav } from "./AppNav";
import { Studio } from "./Studio";

/**
 * Owns the credit balance so the nav and the composer never disagree: a
 * generation decrements it in one place and both surfaces re-render.
 */
export function StudioShell({
  initialCredits,
  name,
  picture,
}: {
  initialCredits: number;
  name?: string;
  picture?: string;
}) {
  const [credits, setCredits] = useState(initialCredits);

  return (
    <>
      <AppNav credits={credits} name={name} picture={picture} />
      <Studio credits={credits} onCreditsChange={setCredits} />
    </>
  );
}
