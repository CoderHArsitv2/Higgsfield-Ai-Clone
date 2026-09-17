import { auth0 } from "@/lib/auth0";
import { getPublicCatalog } from "@/lib/catalog";
import { Nav } from "@/components/landing/Nav";
import { Hero } from "@/components/landing/Hero";
import { LoopSection } from "@/components/landing/LoopSection";
import { ModelWall } from "@/components/landing/ModelWall";
import { Showcase } from "@/components/landing/Showcase";
import { Byok } from "@/components/landing/Byok";
import { FooterCta } from "@/components/landing/FooterCta";

export default async function Home() {
  const [session, catalog] = await Promise.all([
    auth0.getSession(),
    getPublicCatalog(),
  ]);

  return (
    <>
      <Nav signedIn={Boolean(session)} />
      <main>
        <Hero modelCount={catalog.models.length} />
        <LoopSection />
        <ModelWall models={catalog.models} />
        <Showcase />
        <Byok />
      </main>
      <FooterCta />
    </>
  );
}
